package configuration

import (
	"evasbr/mclamg/exception"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func NewDatabase(config Config) *gorm.DB {
	databaseURL := config.Get("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = config.Get("POSTGRES_URL")
	}

	var dsn string
	if databaseURL != "" {
		dsn = databaseURL
	} else {
		username := config.Get("DATASOURCE_USERNAME")
		password := config.Get("DATASOURCE_PASSWORD")
		host := config.Get("DATASOURCE_HOST")
		port := config.Get("DATASOURCE_PORT")
		dbName := config.Get("DATASOURCE_DB_NAME")
		sslMode := config.Get("DATASOURCE_SSL_MODE")
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta", host, username, password, dbName, port, sslMode)
	}
	maxPoolOpen := 10
	if val := config.Get("DATASOURCE_POOL_MAX_CONN"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			maxPoolOpen = i
		}
	}

	maxPoolIdle := 5
	if val := config.Get("DATASOURCE_POOL_IDLE_CONN"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			maxPoolIdle = i
		}
	}

	maxPollLifeTime := 30000
	if val := config.Get("DATASOURCE_POOL_LIFE_TIME"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			maxPollLifeTime = i
		}
	}

	loggerDb := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: loggerDb,
	})
	exception.PanicLogging(err)

	sqlDB, err := db.DB()
	exception.PanicLogging(err)

	sqlDB.SetMaxOpenConns(maxPoolOpen)
	sqlDB.SetMaxIdleConns(maxPoolIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(rand.Int31n(int32(maxPollLifeTime))) * time.Millisecond)

	//autoMigrate
	//err = db.AutoMigrate(&entity.Product{})
	//err = db.AutoMigrate(&entity.Transaction{})
	//err = db.AutoMigrate(&entity.TransactionDetail{})
	//err = db.AutoMigrate(&entity.User{})
	//err = db.AutoMigrate(&entity.UserRole{})
	//exception.PanicLogging(err)
	return db
}
