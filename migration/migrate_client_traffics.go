package main

import (
	"fmt"
	"os"

	"x-ui/database"
	"x-ui/logger"
	"x-ui/xray"

	"github.com/op/go-logging"
)

// This migration script updates the client_traffics table to use a composite unique key
// on (inbound_id, email) instead of just email, allowing the same email to exist
// in different inbounds (useful for v2board sync across multiple nodes)

func main() {
	logger.InitLogger(logging.INFO)

	fmt.Println("Starting database migration...")
	fmt.Println("This will update the client_traffics table to support duplicate emails across different inbounds.")

	db := database.GetDB()
	if db == nil {
		logger.Error("Failed to get database connection")
		os.Exit(1)
	}

	// Check if migration is needed
	var count int64
	err := db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_inbound_email'").Scan(&count).Error
	if err != nil {
		logger.Error("Failed to check migration status:", err)
		os.Exit(1)
	}

	if count > 0 {
		fmt.Println("Migration has already been applied. Skipping...")
		return
	}

	fmt.Println("Backing up current client_traffics table...")

	// Create backup table
	err = db.Exec("CREATE TABLE IF NOT EXISTS client_traffics_backup AS SELECT * FROM client_traffics").Error
	if err != nil {
		logger.Error("Failed to create backup:", err)
		os.Exit(1)
	}

	fmt.Println("Applying migration...")

	// Start transaction
	tx := db.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Error("Migration failed, rolling back:", r)
			os.Exit(1)
		}
	}()

	// Use GORM's AutoMigrate to recreate the table with new structure
	// This will drop and recreate the table based on the updated model
	err = tx.Migrator().DropTable(&xray.ClientTraffic{})
	if err != nil {
		tx.Rollback()
		logger.Error("Failed to drop old table:", err)
		os.Exit(1)
	}

	err = tx.AutoMigrate(&xray.ClientTraffic{})
	if err != nil {
		tx.Rollback()
		logger.Error("Failed to create new table:", err)
		os.Exit(1)
	}

	// Restore data from backup
	// Use COALESCE to set inbound_id to 1 if it's NULL or 0
	err = tx.Exec(`
		INSERT INTO client_traffics (id, inbound_id, enable, email, up, down, all_time, expiry_time, total, reset, last_online)
		SELECT id, COALESCE(NULLIF(inbound_id, 0), 1), enable, email, up, down, all_time, expiry_time, total, reset, last_online
		FROM client_traffics_backup
	`).Error

	if err != nil {
		tx.Rollback()
		logger.Error("Failed to restore data:", err)
		fmt.Println("\nData is safe in client_traffics_backup table. You can manually restore it.")
		os.Exit(1)
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		logger.Error("Failed to commit transaction:", err)
		os.Exit(1)
	}

	fmt.Println("Migration completed successfully!")
	fmt.Println("You can now drop the backup table: DROP TABLE client_traffics_backup;")
	fmt.Println("Please restart the X-Panel application for changes to take effect.")
}
