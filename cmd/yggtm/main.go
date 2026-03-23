package main

import (
	"fmt"
	"log"
	"os"
	"yggtm"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func setupConfig() {

	viper.SetConfigName("yggtm")
	viper.AddConfigPath(".")

	viper.RegisterAlias("server.port", "port")
	viper.SetEnvPrefix("YG")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, "[WARNING] Error occured while reading config file:", err)
	}

	serverCmd := pflag.NewFlagSet("yggtm", pflag.ExitOnError)
	serverCmd.Int("port", 80, "port to listen on")
	serverCmd.Parse(os.Args)

	if err := viper.BindPFlags(serverCmd); err != nil {
		panic(err)
	}
}

func main() {

	setupConfig()

	resMiddle, err := yggtm.NewResourcesMiddleware(viper.Sub("spicedb"))
	if err != nil {
		panic(err)
	}

	authMiddle := yggtm.NewAuthenticationMiddleware(viper.Sub("auth"))

	userService := &yggtm.Service{
		Name: "users service",
		Uri:  "http://localhost:8081",
	}
	orgService := &yggtm.Service{
		Name: "organizations service",
		Uri:  "http://localhost:8080",
	}

	userSubject := yggtm.Subject{
		Name: "user",
		ID:   yggtm.ReceiveFromAuthHeader(yggtm.UserIDFromClaims()),
	}

	server := gin.Default()

	server.POST("/api/auth/login", userService.Proxy())
	server.POST("/api/auth/register", userService.Proxy())
	server.POST("/api/auth/refresh", userService.Proxy())

	server.GET("/api/users/:id", userService.Proxy(), authMiddle.RequireAuth())
	server.POST("/api/users/update-email", userService.Proxy(), authMiddle.RequireAuth())
	server.POST("/api/users/update-password", userService.Proxy(), authMiddle.RequireAuth())

	server.GET("/api/organizations", orgService.Proxy(), authMiddle.RequireAuth())
	server.POST("/api/organizations", orgService.Proxy(), authMiddle.RequireAuth())
	server.GET("/api/organizations/my", orgService.Proxy(), authMiddle.RequireAuth())
	server.GET(
		"/api/organizations/:id",
		orgService.Proxy(),
		authMiddle.RequireAuth(),
		resMiddle.ForResource(
			yggtm.Resource{
				Name:       "organizations",
				ResourceID: yggtm.ReceiveURLParam("id"),
			},
			userSubject,
			"view",
		))
	server.POST(
		"/api/organizations/:id",
		orgService.Proxy(),
		authMiddle.RequireAuth(),
		resMiddle.ForResource(
			yggtm.Resource{
				Name:       "organizations",
				ResourceID: yggtm.ReceiveURLParam("id"),
			},
			userSubject,
			"edit",
		))
	server.DELETE(
		"/api/organizations/:id",
		orgService.Proxy(),
		authMiddle.RequireAuth(),
		resMiddle.ForResource(
			yggtm.Resource{
				Name:       "organizations",
				ResourceID: yggtm.ReceiveURLParam("id"),
			},
			userSubject,
			"delete",
		))

	server.GET(
		"/api/organizations/:id/members",
		orgService.Proxy(),
		authMiddle.RequireAuth(),
		resMiddle.ForResource(
			yggtm.Resource{
				Name:       "organizations",
				ResourceID: yggtm.ReceiveURLParam("id"),
			},
			userSubject,
			"view",
		))
	server.POST(
		"/api/organizations/:id/members",
		orgService.Proxy(),
		authMiddle.RequireAuth(),
		resMiddle.ForResource(
			yggtm.Resource{
				Name:       "organizations",
				ResourceID: yggtm.ReceiveURLParam("id"),
			},
			userSubject,
			"edit",
		))
	server.DELETE(
		"/api/organizations/:id/members",
		orgService.Proxy(),
		authMiddle.RequireAuth(),
		resMiddle.ForResource(
			yggtm.Resource{
				Name:       "organizations",
				ResourceID: yggtm.ReceiveURLParam("id"),
			},
			userSubject,
			"edit",
		))

	server.GET("/api/invitations", orgService.Proxy(), authMiddle.RequireAuth())
	server.POST("/api/invitations/:id/accept", orgService.Proxy(), authMiddle.RequireAuth())
	server.POST("/api/invitations/:id/reject", orgService.Proxy(), authMiddle.RequireAuth())

	if err := server.Run(fmt.Sprintf(":%d", viper.GetInt("server.port"))); err != nil {
		log.Fatal(err)
	}

	// for future
	server.POST(
		"/api/invitations/",
		orgService.Proxy(),
		resMiddle.ForResource(
			yggtm.Resource{
				Name:       "organizations",
				ResourceID: yggtm.ReceiveFromBody("organizationId"),
			},
			userSubject,
			"edit",
		))
}
