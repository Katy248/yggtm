package yggtm

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var client = &http.Client{}

type ErrorResponse struct {
	Details string `json:"details"`
}

func (service *Service) Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Errors != nil {
			err := c.Errors.Last()
			c.JSON(http.StatusInternalServerError, gin.H{
				"details": err.Error(),
			})
			return
		}

		path := c.Request.URL.Path
		if path[0] == '/' {
			path = path[1:]
		}
		req, err := http.NewRequest(c.Request.Method, fmt.Sprintf("%s/%s", service.Uri, path), c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Details: "proxy error"})
			return
		}

		for key, values := range c.Request.Header {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Details: "proxy error"})
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)

	}
}
