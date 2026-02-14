package voucher_test

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"voucher-system/internal/database"
	"voucher-system/internal/voucher"

	_ "github.com/mattn/go-sqlite3"
)

func TestSchemaAndRepository(t *testing.T) {
	// Use a temporary file for the test database
	dbFile := "test_voucher.db"
	os.Remove(dbFile) // clean up before test
	defer os.Remove(dbFile)

	// 1. Initialize DB
	db, err := database.InitDB(dbFile)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// 2. Run Migrations
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// 3. Verify Table Exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='vouchers'").Scan(&tableName)
	if err != nil {
		t.Fatalf("vouchers table not created: %v", err)
	}

	// 4. Test Repository
	repo := voucher.NewSQLiteRepository(db)

	v := &voucher.Voucher{
		ID:        "test-uuid",
		CodeHash:  "test-hash",
		Value:     100,
		Status:    voucher.StatusUnused,
		BatchID:   "batch-1",
		CreatedAt: time.Now(),
	}

	// Test Save
	if err := repo.Save(v); err != nil {
		t.Fatalf("Repository.Save failed: %v", err)
	}

	// Test FindByHash
	found, err := repo.FindByHash("test-hash")
	if err != nil {
		t.Fatalf("Repository.FindByHash failed: %v", err)
	}
	if found == nil {
		t.Fatal("Voucher not found")
	}
	if found.ID != v.ID {
		t.Errorf("Expected ID %s, got %s", v.ID, found.ID)
	}
	if found.Status != voucher.StatusUnused {
		t.Errorf("Expected status %s, got %s", voucher.StatusUnused, found.Status)
	}

	// Test UpdateStatus
	if err := repo.UpdateStatus(v.ID, voucher.StatusUsed); err != nil {
		t.Fatalf("Repository.UpdateStatus failed: %v", err)
	}

	updated, err := repo.FindByHash("test-hash")
	if err != nil {
		t.Fatalf("Repository.FindByHash failed after update: %v", err)
	}
	if updated.Status != voucher.StatusUsed {
		t.Errorf("Expected status %s, got %s", voucher.StatusUsed, updated.Status)
	}
}
