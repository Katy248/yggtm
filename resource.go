package yggtm

import (
	"fmt"

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

		panic("not implemented")
	}
}
