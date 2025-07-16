package main

import (
	"log"

	"GamesWebsite.Shvap/internal/handler"
	"GamesWebsite.Shvap/internal/middleware"
	"GamesWebsite.Shvap/internal/store"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := store.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	handler.UpdateBannerCache(db)

	r := gin.New()
	r.Use(middleware.JWTAuthMiddleware())

	r.LoadHTMLGlob("./web/templates/*")
	r.Static("/static", "./web/static")

	r.GET("/", handler.RedirectHome)
	r.GET("/catalog", handler.LoadHomePage)
	r.GET("/home", handler.LoadMainPage)
	r.GET("/auth", handler.LoadAuthPage)
	r.GET("/newgame", middleware.EnsureAuth(), handler.LoadBannerCreationPage)

	api := r.Group("/api")
	api.GET("/banners", handler.RetrieveBanners)
	api.POST("/register", handler.Register(db))
	api.POST("/login", handler.Login(db))
	api.POST("/newbanner", middleware.EnsureAuth(), handler.NewBanner(db))
	api.POST("/logout", handler.Logout())

	r.Run(":3000")
}
