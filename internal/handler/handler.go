package handler

import (
	"net/http"

	"GamesWebsite.Shvap/internal/store"
	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// Redirect to /home with status code 302.
func RedirectHome(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, "/home")
}

// Load home page HTML.
func LoadHome(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "Home.html", gin.H{})
}

// Load auth page HTML.
func LoadAuth(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "Auth.html", gin.H{})
}

// Check if user input is valid and register a new account.
func Register(db *store.Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if r, _ := db.CheckUserExists(req.Login); r {
			ctx.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}

		if err := db.Register(req.Login, req.Password); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{})
	}

}
