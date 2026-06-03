package main

import (
	"context"
	"evasbr/mclamg/app"
	"evasbr/mclamg/common"
	"evasbr/mclamg/configuration"
	"evasbr/mclamg/db/seed"
	_ "evasbr/mclamg/docs"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

// @title Go Fiber Clean Architecture
// @version 1.0.0
// @description Baseline project using Go Fiber
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email fiber@swagger.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:9999
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey JWT
// @in header
// @name Authorization
// @description Authorization For JWT
func main() {
	// Parse CLI flags
	runSeed := flag.Bool("seed", false, "run database seeders")
	runRollback := flag.Bool("seed-rollback", false, "rollback database seeders")
	createSeed := flag.String("seed-create", "", "generate a new empty seeder file with timestamp prefix")
	runMigrate := flag.Bool("migrate", false, "run all database migrations")
	runMigrateRollback := flag.Bool("migrate-rollback", false, "rollback the last database migration")
	runMigrateRollbackAll := flag.Bool("migrate-rollback-all", false, "rollback all database migrations")
	flag.Parse()

	if *createSeed != "" {
		fmt.Printf("Generating new seeder: %s...\n", *createSeed)
		if err := seed.Create(*createSeed); err != nil {
			log.Fatalf("Failed to create seeder: %v", err)
		}
		return
	}

	//setup configuration
	config := configuration.New()

	if *runMigrate {
		fmt.Println("Running database migrations...")
		if err := runMigrateCommand(config, "up"); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migration completed successfully.")
		return
	}

	if *runMigrateRollback {
		fmt.Println("Rolling back the last database migration...")
		if err := runMigrateCommand(config, "down", "1"); err != nil {
			log.Fatalf("Migration rollback failed: %v", err)
		}
		fmt.Println("Migration rollback completed successfully.")
		return
	}

	if *runMigrateRollbackAll {
		fmt.Println("Rolling back all database migrations...")
		if err := runMigrateCommand(config, "down", "-all"); err != nil {
			log.Fatalf("Migration rollback all failed: %v", err)
		}
		fmt.Println("Migration rollback all completed successfully.")
		return
	}

	database := configuration.NewDatabase(config)

	if *runSeed {
		fmt.Println("Running database seeders...")
		if err := seed.Run(database); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		fmt.Println("Seeding completed successfully.")
		return
	}

	if *runRollback {
		fmt.Println("Rolling back database seeders...")
		if err := seed.Rollback(database, 1); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		fmt.Println("Rollback completed successfully.")
		return
	}

	redis := configuration.NewRedis(config)

	appInstance := app.BuildApp(config, database, redis)

	//start app in a goroutine
	go func() {
		err := appInstance.Listen(config.Get("SERVER.PORT"))
		if err != nil {
			common.Logger(context.Background(), "Main").Infof("Server startup error: %v", err)
		}
	}()

	// Listening for OS termination signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	common.Logger(context.Background(), "Main").Info("Shutting down server gracefully...")

	// Attempt graceful shutdown
	if err := appInstance.Shutdown(); err != nil {
		common.Logger(context.Background(), "Main").Errorf("Server forced to shutdown: %v", err)
	}

	common.Logger(context.Background(), "Main").Info("Server exited")
}



func runMigrateCommand(config configuration.Config, action string, args ...string) error {
	username := config.Get("DATASOURCE_USERNAME")
	password := config.Get("DATASOURCE_PASSWORD")
	host := config.Get("DATASOURCE_HOST")
	port := config.Get("DATASOURCE_PORT")
	dbName := config.Get("DATASOURCE_DB_NAME")
	sslMode := config.Get("DATASOURCE_SSL_MODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", username, password, host, port, dbName, sslMode)

	migrateCmd := "migrate"
	if _, err := exec.LookPath(migrateCmd); err != nil {
		userHome, _ := os.UserHomeDir()
		fallbackPath := filepath.Join(userHome, "go", "bin", "migrate")
		if runtime.GOOS == "windows" {
			fallbackPath += ".exe"
		}
		if _, err := os.Stat(fallbackPath); err == nil {
			migrateCmd = fallbackPath
		}
	}

	cmdArgs := []string{"-database", dbURL, "-path", "db/migrations", action}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(migrateCmd, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
