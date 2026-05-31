package seed

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

//go:embed *.sql
var seedFiles embed.FS

// SeederHistory maps to the database table recording applied seed versions.
type SeederHistory struct {
	Version string    `gorm:"primaryKey;column:version"`
	RunAt   time.Time `gorm:"column:run_at;default:CURRENT_TIMESTAMP"`
}

// TableName returns the custom table name for seeder tracking.
func (SeederHistory) TableName() string {
	return "seeder_history"
}

// Create generates a pair of empty .up.sql and .down.sql files in the db/seed directory using a timestamp prefix.
func Create(name string) error {
	// Clean the name of spaces/special characters
	cleanedName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		if r == ' ' {
			return '_'
		}
		return -1
	}, name)

	timestamp := time.Now().Format("20060102150405")
	baseName := fmt.Sprintf("%s_%s", timestamp, cleanedName)

	upPath := filepath.Join("db", "seed", baseName+".up.sql")
	downPath := filepath.Join("db", "seed", baseName+".down.sql")

	// Ensure seed directory exists
	if err := os.MkdirAll(filepath.Join("db", "seed"), 0755); err != nil {
		return fmt.Errorf("failed to create db/seed directory: %w", err)
	}

	// Create up file
	upFile, err := os.Create(upPath)
	if err != nil {
		return fmt.Errorf("failed to create up-migration seeder file: %w", err)
	}
	defer upFile.Close()
	_, _ = upFile.WriteString(fmt.Sprintf("-- SQL Seed up-migration: %s\n\n", cleanedName))

	// Create down file
	downFile, err := os.Create(downPath)
	if err != nil {
		return fmt.Errorf("failed to create down-migration seeder file: %w", err)
	}
	defer downFile.Close()
	_, _ = downFile.WriteString(fmt.Sprintf("-- SQL Seed down-migration: %s\n\n", cleanedName))

	fmt.Printf("Created seeder files:\n  - %s\n  - %s\n", upPath, downPath)
	return nil
}

// Run executes all pending .up.sql seeders in alphanumeric order.
func Run(db *gorm.DB) error {
	// 1. Ensure the seeder history table exists
	if err := db.AutoMigrate(&SeederHistory{}); err != nil {
		return fmt.Errorf("failed to migrate seeder_history table: %w", err)
	}

	// 2. Read all embedded seed files
	entries, err := seedFiles.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read embedded seed directory: %w", err)
	}

	// 3. Filter and sort up-migration files
	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}
	sort.Strings(upFiles)

	// 4. Apply pending seeders
	for _, fileName := range upFiles {
		version := strings.TrimSuffix(fileName, ".up.sql")

		var exists bool
		err := db.Model(&SeederHistory{}).
			Select("count(*) > 0").
			Where("version = ?", version).
			Find(&exists).Error
		if err != nil {
			return fmt.Errorf("failed to query seeder history for %s: %w", version, err)
		}

		if exists {
			fmt.Printf("Seeder %s already applied, skipping.\n", version)
			continue
		}

		fmt.Printf("Applying database seed: %s...\n", version)

		content, err := seedFiles.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("failed to read seeder content %s: %w", fileName, err)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			sqlQueries := string(content)
			if strings.TrimSpace(sqlQueries) != "" {
				if err := tx.Exec(sqlQueries).Error; err != nil {
					return err
				}
			}

			history := SeederHistory{
				Version: version,
				RunAt:   time.Now(),
			}
			return tx.Create(&history).Error
		})

		if err != nil {
			return fmt.Errorf("failed executing seeder %s: %w", version, err)
		}

		fmt.Printf("Successfully applied seed: %s\n", version)
	}

	return nil
}

// Rollback rolls back seeders (reverting down migrations) in reverse order.
func Rollback(db *gorm.DB, steps int) error {
	if err := db.AutoMigrate(&SeederHistory{}); err != nil {
		return fmt.Errorf("failed to migrate seeder_history table: %w", err)
	}

	var history []SeederHistory
	if err := db.Order("run_at desc, version desc").Find(&history).Error; err != nil {
		return fmt.Errorf("failed to fetch seeder history: %w", err)
	}

	if len(history) == 0 {
		fmt.Println("No applied seeders found to roll back.")
		return nil
	}

	limit := len(history)
	if steps > 0 && steps < limit {
		limit = steps
	}

	for i := 0; i < limit; i++ {
		version := history[i].Version
		downFileName := version + ".down.sql"

		fmt.Printf("Rolling back database seed: %s...\n", version)

		content, err := seedFiles.ReadFile(downFileName)
		if err != nil {
			return fmt.Errorf("failed to read rollback seeder content %s: %w", downFileName, err)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			sqlQueries := string(content)
			if strings.TrimSpace(sqlQueries) != "" {
				if err := tx.Exec(sqlQueries).Error; err != nil {
					return err
				}
			}

			return tx.Where("version = ?", version).Delete(&SeederHistory{}).Error
		})

		if err != nil {
			return fmt.Errorf("failed executing rollback for %s: %w", version, err)
		}

		fmt.Printf("Successfully rolled back seed: %s\n", version)
	}

	return nil
}
