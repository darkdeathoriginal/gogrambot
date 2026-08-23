package models

import (
	"log"
	"strings"

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

	switch {
	case dbURL == "":
		dbFile := "bot.db?_journal_mode=WAL&_busy_timeout=5000"
		log.Printf("DATABASE_URL not set, using SQLite fallback: %s\n", dbFile)
		dialector = sqlite.Open(dbFile)

	case strings.HasPrefix(dbURL, "postgres://"),
		strings.HasPrefix(dbURL, "postgresql://"):
		log.Println("PostgreSQL DATABASE_URL found, connecting...")
		dialector = postgres.New(postgres.Config{
			DSN:                  dbURL,
			PreferSimpleProtocol: true,
		})

	default:
		log.Printf("Using SQLite database: %s\n", dbURL)
		dialector = sqlite.Open(dbURL)
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
