package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Redirect to /home with status code 302.
func RedirectHome(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, "/home")
}

// Load home page HTML.
func LoadHome(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "TEST.html", gin.H{})
}
