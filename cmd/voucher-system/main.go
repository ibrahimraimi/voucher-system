package main

import (
	"fmt"
	"log"
	"time"

	"voucher-system/internal/database"
	"voucher-system/internal/observability"
	"voucher-system/internal/ratelimit"
	"voucher-system/internal/voucher"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Initialize Database
	db, err := database.InitDB("vouchers.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatal(err)
	}

	// Initialize Service
	repo := voucher.NewSQLiteRepository(db)
	limiter := ratelimit.NewLimiter(3, time.Minute)
	service := voucher.NewService(repo, limiter)

	// Start Dashboard Loop
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			metrics := observability.GetMetrics()
			fmt.Println("\n--- System Dashboard ---")
			fmt.Printf("Total Generated: %d\n", metrics.TotalGenerated)
			fmt.Printf("Total Redeemed:  %d\n", metrics.TotalRedeemed)
			fmt.Printf("Failed Attempts: %d\n", metrics.FailedAttempts)
			fmt.Println("------------------------")
		}
	}()

	// Example Usage / Simulation
	fmt.Println("System Started. Generating a batch of vouchers...")
	pins, _, err := service.CreateBatch(100, 5)
	if err != nil {
		log.Printf("Failed to create batch: %v", err)
	} else {
		log.Printf("generated %d vouchers", len(pins))
		// Log the first PIN for manual testing if needed
		log.Printf("Example PIN: %s", pins[0])
	}

	// Keep main alive
	select {}
}
