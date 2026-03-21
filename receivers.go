package yggtm

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type ReceiverFunc func(c *gin.Context) (string, error)
type ResourceIDReceiverFunc ReceiverFunc

func ReceiveURLParam(param string) ResourceIDReceiverFunc {
	return func(c *gin.Context) (string, error) {
		return c.Param(param), nil
	}
}

func ReceiveFromBody(key string) ResourceIDReceiverFunc {
	return func(c *gin.Context) (string, error) {
		body := map[string]any{}
		if err := c.ShouldBindJSON(&body); err != nil {
			return "", fmt.Errorf("failed bind json body: %s", err)
		}

		value, ok := body[key]
		if !ok {
			return "", fmt.Errorf("key %s not found in body", key)
		}
		str, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("value for key %s is not a string", key)
		}
		return str, nil
	}
}

func ReceiveQueryParam(param string) ReceiverFunc {
	return func(c *gin.Context) (string, error) {
		return c.Query(param), nil
	}
}

func ReceiveHeader(header string) ReceiverFunc {
	return func(c *gin.Context) (string, error) {
		return c.GetHeader(header), nil
	}
}

type AuthHeaderReceiverFunc func(string) (string, error)

func ReceiveFromJWT(key string) AuthHeaderReceiverFunc {
	return func(header string) (string, error) {
		panic("not implemented")
	}
}

func ReceiveFromAuthHeader(receiver AuthHeaderReceiverFunc) ReceiverFunc {
	return func(c *gin.Context) (string, error) {
		return receiver(c.GetHeader("Authorization"))
	}
}
