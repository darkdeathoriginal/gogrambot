package models

import (
	"os"
	"testing"
)

type DummyModel struct {
	ID uint
}

func TestInitDatabaseAndAddMigrate(t *testing.T) {
	AddModelToMigrate(&DummyModel{})

	db := InitDatabase()

	if db == nil {
		t.Fatalf("Expected db to not be nil")
	}

	if DB == nil {
		t.Fatalf("Expected global DB variable to be set")
	}

	defer os.Remove("bot.db")

	if !db.Migrator().HasTable(&DummyModel{}) {
		t.Errorf("Expected 'dummy_models' table to exist")
	}
}
