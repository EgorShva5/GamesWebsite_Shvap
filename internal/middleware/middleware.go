package middleware

import (
	"net/http"

	"GamesWebsite.Shvap/internal/handler"
	"GamesWebsite.Shvap/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Middleware for checking JWT.
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr, err := ctx.Cookie("jwt_token")
		if err != nil {
			//ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no token found"})
			ctx.Next()
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &handler.CustomClaims{}, func(token *jwt.Token) (any, error) {
			return []byte(store.Cfg.Keys.JWT), nil
		})

		if err != nil || !token.Valid {
			//ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			ctx.Next()
			return
		}

		if claims, ok := token.Claims.(*handler.CustomClaims); ok {
			ctx.Set("userData", gin.H{
				"Display": claims.Display,
				"Login":   claims.Login,
				"Role":    claims.Role,
			})
		} else {
			//ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to read token"})
			return
		}

		ctx.Next()
	}
}

// Check if user is authorized
func EnsureAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if _, exists := ctx.Get("userData"); !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		ctx.Next()
	}
}
