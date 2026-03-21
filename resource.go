package yggtm

import "github.com/gin-gonic/gin"

type Subject struct {
	Name string
	ID   ReceiverFunc
}

type Resource struct {
	Name       string
	ResourceID ResourceIDReceiverFunc
}

func ForResource(resource Resource, subject Subject, permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {

		panic("not implemented")
	}
}
