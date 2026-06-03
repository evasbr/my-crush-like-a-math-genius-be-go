package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"evasbr/mclamg/app"
	"evasbr/mclamg/configuration"

	"github.com/go-redis/redis/v9"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var (
	fiberApp *fiber.App
	once     sync.Once
	db       *gorm.DB
	rdb      *redis.Client
)

func initApp() {
	once.Do(func() {
		// Load config (reads from OS environment variables in Vercel)
		config := configuration.New()
		db = configuration.NewDatabase(config)
		rdb = configuration.NewRedis(config)
		fiberApp = app.BuildApp(config, db, rdb)
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// This is needed to set the proper request path in `fiber.Ctx`
	r.RequestURI = r.URL.String()

	handler().ServeHTTP(w, r)
}

func handler() http.HandlerFunc {
	initApp()
	return adaptor.FiberApp(fiberApp)
}

func main() {
	initApp()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(fiberApp.Listen(":" + port))
}
