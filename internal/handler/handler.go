package handler

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"GamesWebsite.Shvap/internal/store"
	"GamesWebsite.Shvap/web/templates"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

//
/* VARIABLES AND STRUCTS */
//

// Amount of games created.
var BannerCount int

// Slice of all banners.
var BannerSlice []store.Banner

// Items per page.
const PerPage = 9

// Max page, that's it.
var MaxPage uint64

// Variable for template rendering
var t = templates.New()

// Allowed extensions. WIP -> check magic numbers instead.
var Extensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
}

// Struct for registration data.
type RegisterRequest struct {
	DisplayName string `json:"display" binding:"required,min=4,max=32"`
	LoginRequest
}

// Struct for login credentials.
type LoginRequest struct {
	Login    string `json:"login" binding:"required,min=2,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// Data inside a JWT.
type CustomClaims struct {
	Display string `json:"display"`
	Login   string `json:"username"`
	Role    string `json:"role"`
	jwt.RegisteredClaims
}

//
/* VARIABLES AND STRUCTS END */
//

// Redirect to /home with status code 302.
func RedirectHome(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, "/home")
}

// Load home page.
func LoadHomePage(ctx *gin.Context) {
	str := ctx.Query("page")

	page, err := strconv.ParseUint(str, 10, 64)
	if err != nil || page == 0 {
		page = 1
	}
	if page > MaxPage {
		page = MaxPage
	}
	end := min(PerPage*page, uint64(len(BannerSlice)))

	t.AutoRender(ctx, "Home.html", gin.H{
		"BannerCount": BannerCount,
		"Banners":     BannerSlice[(page-1)*PerPage : end],
		"Page":        page,
		"MaxPage":     MaxPage,
	})
}

// Load auth page.
func LoadAuthPage(ctx *gin.Context) {
	t.AutoRender(ctx, "Auth.html", nil)
}

// Load banner creation page.
func LoadBannerCreationPage(ctx *gin.Context) {
	t.AutoRender(ctx, "NewBanner.html", nil)
}

// Retrieve all banners.
func RetrieveBanners(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, BannerSlice)
}

// Create a banner.
func NewBanner(db *store.Database) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var display string
		userData, exists := ctx.Get("userData")
		if exists {
			if data, ok := userData.(gin.H); ok {
				display = data["Display"].(string)
			}
		}
		if len(display) == 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "can't get your name :("})
			return
		}

		if !strings.Contains(ctx.ContentType(), "multipart/form-data") {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "multipart/form-data required"})
			return
		}

		title := ctx.PostForm("title")
		description := ctx.PostForm("description")
		url := ctx.PostForm("url")

		file, err := ctx.FormFile("image")
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to get image"})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !Extensions[ext] {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "allowed extensions: png, jpg, jpeg, webp"})
			return
		}

		var imageName string
		var uploadPath string
		for {
			imageName = uuid.New().String() + ext
			uploadPath = "./web/static/img/banners/" + imageName
			if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
				break
			}
		}
		if err := ctx.SaveUploadedFile(file, uploadPath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save image"})
			return
		}

		if err := db.CheckBannerExists(title); err != nil {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		err = db.NewBanner(title, description, display, url, imageName)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		BannerCount, err = db.UpdateBannerCount()
		if err != nil {
			log.Println(err.Error())
		}
		BannerSlice, err = db.UpdateBannerSlice()
		if err != nil {
			log.Println(err.Error())
		}
		MaxPage = uint64((len(BannerSlice) + PerPage - 1) / PerPage)

		ctx.JSON(http.StatusCreated, gin.H{})
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

		display, err := db.GetDisplay(req.Login)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		token, err := GenerateJWT(display, req.Login)
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
func GenerateJWT(display, login string) (string, error) {
	claims := CustomClaims{
		Display: display,
		Login:   login,
		Role:    "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(store.Cfg.Keys.JWT))
}

// Can't make a db method :( (I actually can but it does not worth it)
func UpdateBannerCache(db *store.Database) error {
	var err error
	BannerCount, err = db.UpdateBannerCount()
	if err != nil {
		return err
	}
	BannerSlice, err = db.UpdateBannerSlice()
	if err != nil {
		return err
	}
	MaxPage = uint64((len(BannerSlice) + PerPage - 1) / PerPage)
	return nil
}

// Delete JWT
func Logout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.SetCookie(
			"jwt_token",
			"",
			-1,
			"/",
			"",
			false,
			true,
		)

		ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}
