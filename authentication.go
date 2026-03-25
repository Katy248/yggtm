package yggtm

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type AuthenticationMiddleware struct {
	config *viper.Viper
}

func NewAuthenticationMiddleware(config *viper.Viper) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		config: config,
	}
}

const AuthorizationHeader = "Authorization"

const UserIDHeader = "YGGTM-User-ID"
const UserEmailHeader = "YGGTM-User-Email"

type Claims struct {
	jwt.RegisteredClaims
	UserID    string `json:"userId"`
	UserEmail string `json:"userEmail"`
}

func (am *AuthenticationMiddleware) parseClaims(token string) (Claims, error) {
	var parsedClaims Claims
	_, err := jwt.ParseWithClaims(token, &parsedClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte(am.config.GetString("secret")), nil
	})
	if err != nil {
		return Claims{}, err
	}
	return parsedClaims, nil
}
func (am *AuthenticationMiddleware) WithMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(AuthMiddlewareKey, am)
	}
}
func (am *AuthenticationMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(AuthMiddlewareKey, am)

		header := c.GetHeader(AuthorizationHeader)
		if header == "" {
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("missing %q header", AuthorizationHeader))
			return
		}

		token := strings.TrimSpace(header)
		if token == "" {
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("missing token in %q header", AuthorizationHeader))
			return
		}

		claims, err := am.parseClaims(token)
		if err != nil {
			c.AbortWithError(http.StatusUnauthorized, err)
			return
		}
		c.Header(UserIDHeader, claims.UserID)
		c.Header(UserEmailHeader, claims.UserEmail)

		c.Set(AuthClaimsKey, claims)

		c.Next()
	}
}

const AuthClaimsKey = "yggtm-claims"
const AuthMiddlewareKey = "yggtm-auth-middleware"

func getClaims(c *gin.Context) Claims {
	claims := c.MustGet(AuthClaimsKey).(Claims)
	return claims
}

func getAuthenticationMiddleware(c *gin.Context) (*AuthenticationMiddleware, error) {
	data, ok := c.Get(AuthMiddlewareKey)
	if !ok {
		return nil, fmt.Errorf("no auth middleware was provided")
	}
	m, ok := data.(*AuthenticationMiddleware)
	if !ok {
		return nil, fmt.Errorf("invalid auth middleware type was provided in context")
	}
	return m, nil
}

func mustGetAuthenticationMiddleware(c *gin.Context) *AuthenticationMiddleware {
	m, err := getAuthenticationMiddleware(c)
	if err != nil {
		panic(fmt.Errorf("failed to get auth middleware: %s", err))
	}
	return m
}
