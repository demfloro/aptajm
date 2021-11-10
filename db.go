package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func initDB(ctx context.Context, filename string, logger *log.Logger) (*sql.DB, context.CancelFunc, error) {
	ctx, backupCancel := context.WithCancel(ctx)
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, func() {}, err
	}
	cancel := func() { backupCancel(); db.Close() }
	go backupLoop(ctx, db, filename+".backup")
	return db, cancel, nil
}

func backupLoop(ctx context.Context, db *sql.DB, filename string) {
	backup, err := db.PrepareContext(ctx, `VACUUM INTO ?`)
	if err != nil {
		log.Printf("Failed to prepare backup query: %q", err)
		return
	}
	ticker := time.NewTicker(Config.BackupTimeout)

	for {
		select {
		case <-ctx.Done():
			backup.Close()
			ticker.Stop()
			return
		case <-ticker.C:
		}
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		_, err = backup.ExecContext(ctx, "."+filename)
		cancel()
		if err != nil {
			log.Printf("Failed to backup DB: %q", err)
			continue
		}
		if err = os.Rename("."+filename, filename); err != nil {
			log.Printf("Failed to rename: %q", err)
			if err = os.Remove("." + filename); err != nil {
				log.Printf("Failed to remove: %q", err)
			}
		}

	}
}
