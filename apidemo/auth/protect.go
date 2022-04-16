package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt"
)

func Protect(signature []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		tokenString := strings.TrimPrefix(auth, "Bearer ")

		_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected  signing method: %v", token.Header["alg"])
			}
			return signature, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"msg": "Unauthorized",
			})
			return
		}
		c.Next()
	}
}
