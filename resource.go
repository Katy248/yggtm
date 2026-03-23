package yggtm

import (
	"context"
	"fmt"
	"net/http"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	authzed "github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Subject struct {
	Name string
	ID   ReceiverFunc
}

type Resource struct {
	Name       string
	ResourceID ResourceIDReceiverFunc
}

type ResourcesMiddleware struct {
	spiceDB *authzed.Client
}

func getCredentials(config *viper.Viper) ([]grpc.DialOption, error) {
	secure := !config.GetBool("insecure")
	token := config.GetString("token")

	if secure {
		systemCerts, err := grpcutil.WithSystemCerts(grpcutil.VerifyCA)
		if err != nil {
			return nil, fmt.Errorf("unable to load CA certificates: %s", err)
		}
		return []grpc.DialOption{
			systemCerts,
			grpcutil.WithBearerToken(token),
		}, nil
	}

	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(token),
	}, nil
}

func NewResourcesMiddleware(config *viper.Viper) (*ResourcesMiddleware, error) {

	creds, err := getCredentials(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get credentials: %s", err)
	}

	addr := config.GetString("address")
	if addr == "" {
		return nil, fmt.Errorf("spicedb address is required")
	}

	client, err := authzed.NewClient(
		addr,
		creds...,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize spicedb client: %s", err)
	}

	return &ResourcesMiddleware{
		spiceDB: client,
	}, nil
}

func (rm *ResourcesMiddleware) ForResource(resource Resource, subject Subject, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		resourceID, err := resource.ResourceID(c)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		subjectID, err := subject.ID(c)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		rm.spiceDB.CheckPermission(context.TODO(), &v1.CheckPermissionRequest{
			Resource: &v1.ObjectReference{
				ObjectType: resource.Name,
				ObjectId:   resourceID,
			},
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectType: subject.Name,
					ObjectId:   subjectID,
				},
			},
			Permission: permissions[0],
		})

		requestItems := make([]*v1.CheckBulkPermissionsRequestItem, len(permissions))

		resourceItem := &v1.ObjectReference{
			ObjectType: resource.Name,
			ObjectId:   resourceID,
		}
		subjectItem := &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: subject.Name,
				ObjectId:   subjectID,
			},
		}

		for i, permission := range permissions {
			requestItems[i] = &v1.CheckBulkPermissionsRequestItem{
				Resource:   resourceItem,
				Subject:    subjectItem,
				Permission: permission,
			}
		}

		result, err := rm.spiceDB.CheckBulkPermissions(context.TODO(), &v1.CheckBulkPermissionsRequest{
			Items: requestItems,
			Consistency: &v1.Consistency{
				Requirement: &v1.Consistency_FullyConsistent{
					FullyConsistent: true,
				},
			},
		})

		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		for _, pair := range result.Pairs {
			if err := pair.GetError(); err != nil {
				c.AbortWithError(http.StatusForbidden, fmt.Errorf("failed to check permission: %s", err))
				return
			}

			item := pair.GetItem()
			if item == nil {
				c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("permission check response item in nil"))
			}
			if item.Permissionship != v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
				c.AbortWithError(http.StatusForbidden,
					fmt.Errorf("%s:%s has not required permissions (%w) for %s:%s", subjectItem.Object.ObjectType, subjectItem.Object.ObjectId, permissions, resourceItem.ObjectType, resourceItem.ObjectId))
				return
			}
		}

	}
}
