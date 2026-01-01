package models

import (
	"log"

	"github.com/darkdeathoriginal/gogrambot/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDatabase connects to the database, runs migrations, and returns the configured object.
// It is designed to be called ONCE at application startup.
var DB *gorm.DB
var ModelsToMigrate = []interface{}{}

func InitDatabase() *gorm.DB {
	var dialector gorm.Dialector
	dbURL := config.Getenv("DATABASE_URL", "")

	if dbURL != "" {
		log.Println("DATABASE_URL found, connecting to PostgreSQL...")
		dialector = postgres.New(postgres.Config{
			DSN:                  dbURL,
			PreferSimpleProtocol: true,
		})
	} else {
		dbFile := "bot.db"
		log.Printf("DATABASE_URL not set, using SQLite fallback: %s\n", dbFile)
		dialector = sqlite.Open(dbFile)
	}

	// Connect to the database, making sure to disable the prepared statement cache.
	db, err := gorm.Open(dialector, &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to the database: %v", err)
	}
	log.Println("Database connection successful.")

	log.Println("Running database migrations...")

	for _, model := range ModelsToMigrate {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("WARNING: Migration for %T failed (this is often safe if tables already exist): %v", model, err)
		} else {
			log.Printf("%T migration completed successfully.\n", model)
		}
	}

	DB = db
	return db
}

func AddModelToMigrate(model interface{}) {
	ModelsToMigrate = append(ModelsToMigrate, model)
}
