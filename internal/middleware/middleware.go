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
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no token found"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &handler.CustomClaims{}, func(token *jwt.Token) (any, error) {
			return []byte(store.Cfg.Keys.JWT), nil
		})

		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			return
		}

		if claims, ok := token.Claims.(*handler.CustomClaims); ok {
			ctx.Set("login", claims.Login)
			ctx.Set("role", claims.Role)
		} else {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to read token"})
			return
		}

		ctx.Next()
	}
}
