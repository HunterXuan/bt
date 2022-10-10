package main

import (
	"context"
	"github.com/HunterXuan/bt/app/infra/config"
	"github.com/HunterXuan/bt/app/infra/db"
	"github.com/HunterXuan/bt/app/infra/dht"
	"github.com/HunterXuan/bt/app/infra/scheduler"
	"github.com/HunterXuan/bt/app/router"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Init
	config.InitConfig()
	db.InitRedisClient()
	scheduler.InitScheduler()
	dht.InitDHT()

	// Construct server
	r := router.InitRouter()
	srv := &http.Server{
		Addr:    config.Config.GetString("APP_LISTEN_ADDR"),
		Handler: r,
	}

	go func() {
		// Listen and serve
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Err: %s\n", err)
		}
	}()

	log.Println("Server URL: http://" + config.Config.GetString("APP_LISTEN_ADDR") + "/")
	log.Println("Enter Control + C Shutdown Server")

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
