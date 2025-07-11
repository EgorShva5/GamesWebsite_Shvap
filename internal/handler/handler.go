package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"GamesWebsite.Shvap/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Amount of games created.
var GameCount int

// Slice of all banners.
var BannerSlice []store.Banner

// Struct for registration data.
type RegisterRequest struct {
	DisplayName string `json:"display" binding:"required,min=4,max=32"`
	LoginRequest
}

// Struct for login credentials.
type LoginRequest struct {
	Login    string `json:"login" binding:"required,min=2,max=32"`
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
	str := ctx.Query("page")

	page, err := strconv.ParseUint(str, 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	if page > MaxPage {
		page = MaxPage
	}

	end := min(PerPage*page, uint64(len(BannerSlice)))
	ctx.HTML(http.StatusOK, "Home.html", gin.H{
		"GameCount": GameCount,
		"banners":   BannerSlice[(page-1)*PerPage : end],
		"page":      page,
		"maxpage":   MaxPage,
	})
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

// Retrieve all banners.
func RetrieveBanners(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, BannerSlice)
}

// Create a banner.
func NewBanner(db *store.Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		login := ctx.MustGet("login").(string)
		_ = ctx.MustGet("role").(string)

		var req store.BannerParse

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.CheckBannerExists(req.Title); err != nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		err := db.NewBanner(req.Title, req.Description, login, req.Url)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{})
		GameCount, err = db.UpdateGames()
		if err != nil {
			log.Println(err.Error())
		}
		BannerSlice, err = db.UpdateBannerSlice()
		if err != nil {
			log.Println(err.Error())
		}
		MaxPage = uint64((len(BannerSlice) + PerPage - 1) / PerPage)
	}
}

// Check if user input is valid and register a new account.
func Register(db *store.Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RegisterRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.CheckUserExists(req.DisplayName, req.Login); err != nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		if err := db.Register(req.DisplayName, req.Login, req.Password); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{})
	}
}

// Check password and authorize user, store JWT-token as a cookie.
func Login(db *store.Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req LoginRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.CheckPassword(req.Login, req.Password); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

// * Pagination * //
const PerPage = 3

var MaxPage uint64
