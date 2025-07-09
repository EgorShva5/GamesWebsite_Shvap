package handler

import (
	"net/http"
	"time"

	"GamesWebsite.Shvap/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Struct for auth credentials.
type AuthRequest struct {
	Login    string `json:"login" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

// Data inside a JWT.
type CustomClaims struct {
	Login string `json:"username"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// Redirect to /home with status code 302.
func RedirectHome(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, "/home")
}

// Load home page.
func LoadHome(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "Home.html", gin.H{})
}

// Load auth page.
func LoadAuth(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "Auth.html", gin.H{})
}

// Load banner creation page.
func LoadGameMaker(ctx *gin.Context) {
	_ = ctx.MustGet("login").(string)
	_ = ctx.MustGet("role").(string)

	ctx.HTML(http.StatusOK, "NewBanner.html", gin.H{})
}

// Check if user input is valid and register a new account.
func Register(db *store.Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AuthRequest

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

// Check password and authorize user, store JWT-token as a cookie.
func Login(db *store.Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AuthRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if r := db.CheckPassword(req.Login, req.Password); r != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to login"})
			return
		}

		token, err := GenerateJWT(req.Login)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not create a token 4 u :-("})
		}

		ctx.SetCookie(
			"jwt_token",
			token,
			int(24*time.Hour.Seconds()),
			"/",
			"",
			false,
			true,
		)

		ctx.JSON(http.StatusAccepted, gin.H{})
	}
}

// JWT generation.
func GenerateJWT(login string) (string, error) {
	claims := CustomClaims{
		Login: login,
		Role:  "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(store.Cfg.Keys.JWT))
}
