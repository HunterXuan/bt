package router

import (
	"github.com/HunterXuan/bt/app/controller/handler/stats"
	"github.com/HunterXuan/bt/app/controller/handler/tracker"
	"github.com/gin-gonic/gin"
	"log"
)

func InitRouter() *gin.Engine {
	log.Println("Router Initializing...")

	r := gin.New()

	// Global middleware
	r.Use(gin.Logger(), gin.Recovery())

	// index.html
	r.StaticFile("/", "./storage/public/index.html")
	// Robots.txt
	r.StaticFile("robots.txt", "./storage/public/robots.txt")
	// favicon.ico
	r.StaticFile("favicon.ico", "./storage/public/favicon.ico")
	// assets
	r.Static("/assets", "./storage/public/assets")

	{
		// Tracker related API
		r.GET("/announce", tracker.Announce)
		r.GET("/scrape", tracker.Scrape)
	}

	{
		// Stats related API
		r.GET("/stats", stats.GetAllStats)
	}

	log.Println("Router Initialized!")

	return r
}
