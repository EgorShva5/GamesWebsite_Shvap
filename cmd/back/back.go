package main

import (
	"log"

	"GamesWebsite.Shvap/internal/handler"
	"GamesWebsite.Shvap/internal/store"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := store.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	r := gin.New()
	r.LoadHTMLGlob("web/templates/*")
	r.Static("/static", "./web/static")

	r.GET("/", handler.RedirectHome)
	r.GET("home", handler.LoadHome)
	r.GET("auth", handler.LoadAuth)

	r.POST("auth", handler.Register(db))

	r.Run(":3000")
}
