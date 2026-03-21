package main

import (
	"yggtm"

	"github.com/gin-gonic/gin"
)

func requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		panic("not implemented")
	}
}

func main() {
	userService := &yggtm.Service{
		Name: "users service",
		Uri:  "http://localhost:8080",
	}
	orgService := &yggtm.Service{
		Name: "organizations service",
		Uri:  "http://localhost:8081",
	}

	userSubject := yggtm.Subject{
		Name: "user",
		ID:   yggtm.ReceiveFromAuthHeader(yggtm.ReceiveFromJWT("userId")),
	}

	server := gin.Default()

	server.POST("/api/auth/login", userService.Proxy())
	server.POST("/api/auth/register", userService.Proxy())
	server.POST("/api/auth/refresh", userService.Proxy())

	server.GET("/api/users/:id", userService.Proxy(), requireAuth())
	server.POST("/api/users/update-email", userService.Proxy(), requireAuth())
	server.POST("/api/users/update-password", userService.Proxy(), requireAuth())

	server.GET("/api/organizations", orgService.Proxy())
	server.POST("/api/organizations", orgService.Proxy())
	server.GET("/api/organizations/my", orgService.Proxy())
	server.GET(
		"/api/organizations/:id",
		orgService.Proxy(),
		yggtm.ForResource(
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
		yggtm.ForResource(
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
		yggtm.ForResource(
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
		yggtm.ForResource(
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
		yggtm.ForResource(
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
		yggtm.ForResource(
			yggtm.Resource{
				Name:       "organizations",
				ResourceID: yggtm.ReceiveURLParam("id"),
			},
			userSubject,
			"edit",
		))

	server.GET("/api/invitations", orgService.Proxy(), requireAuth())
	server.POST("/api/invitations/:id/accept", orgService.Proxy(), requireAuth())
	server.POST("/api/invitations/:id/reject", orgService.Proxy(), requireAuth())

	server.Run(":80")

	// for future
	server.POST(
		"/api/invitations/",
		orgService.Proxy(),
		yggtm.ForResource(
			yggtm.Resource{
				Name:       "organizations",
				ResourceID: yggtm.ReceiveFromBody("organizationId"),
			},
			userSubject,
			"edit",
		))
}
