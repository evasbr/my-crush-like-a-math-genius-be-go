package handler

import (
	"net/http"
	"sync"

	"evasbr/mclamg/app"
	"evasbr/mclamg/configuration"

	fiberadaptor "github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/go-redis/redis/v9"
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

// Handler is the entrypoint for Vercel Serverless Functions.
func Handler(w http.ResponseWriter, r *http.Request) {
	initApp()
	fiberadaptor.HTTPHandler(fiberApp.Handler()).ServeHTTP(w, r)
}
