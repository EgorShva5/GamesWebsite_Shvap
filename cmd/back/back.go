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

	handler.GameCount, err = db.UpdateGames()
	if err != nil {
		log.Fatal(err)
	}
	handler.BannerSlice, err = db.UpdateBannerSlice()
	if err != nil {
		log.Fatal(err)
	}
	handler.MaxPage = uint64((len(handler.BannerSlice) + handler.PerPage - 1) / handler.PerPage)

	r := gin.New()
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	r.GET("/", handler.RedirectHome)
	r.GET("home", handler.LoadHome)
	r.GET("auth", handler.LoadAuth)

	r.GET("api/banners", handler.RetrieveBanners)
	r.POST("api/register", handler.Register(db))
	r.POST("api/login", handler.Login(db))
	r.POST("api/newbanner", middleware.JWTAuthMiddleware(), handler.NewBanner(db))

	r.GET("/newgame", middleware.JWTAuthMiddleware(), handler.LoadGameMaker)

	r.Run(":3000")
}
