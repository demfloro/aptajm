package main

import (
	"context"
	"database/sql"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/tomb.v2"
)

func (b *ircbot) initDB() error {
	db, err := sql.Open("sqlite3", b.config.dbname)
	if err != nil {
		return err
	}
	ctx := b.tomb.Context(nil)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	stmts, err := initStmts(ctx, db)
	cancel()
	if err != nil {
		db.Close()
		return err
	}
	b.db = db
	b.stmts = stmts
	b.tomb.Go(b.backupLoop)
	return nil
}

func (b *ircbot) backupLoop() error {
	ctx := b.tomb.Context(nil)
	backup, err := b.db.PrepareContext(ctx, `VACUUM INTO ?`)
	if err != nil {
		b.Debug("Failed to prepare backup query: %q", err)
		return err
	}
	ticker := time.NewTicker(time.Hour / 2)
	filename := b.config.dbname

	for {
		select {
		case <-b.tomb.Dying():
			backup.Close()
			ticker.Stop()
			return tomb.ErrDying
		case <-ticker.C:
		}
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		_, err = backup.ExecContext(ctx, "."+filename)
		cancel()
		if err != nil {
			b.Logf("Failed to backup DB: %q", err)
			continue
		}
		if err = os.Rename("."+filename, filename); err != nil {
			b.Logf("Failed to rename: %q", err)
			if err = os.Remove("." + filename); err != nil {
				b.Logf("Failed to remove: %q", err)
			}
		}

	}
}

func initStmts(ctx context.Context, db *sql.DB) (map[dbStmt]*sql.Stmt, error) {
	prepQ, err := db.PrepareContext(ctx,
		`SELECT id, date, rating, text FROM quotes WHERE id=?`)
	if err != nil {
		return nil, err
	}
	prepRandom, err := db.PrepareContext(ctx,
		`SELECT id, date, rating, text FROM quotes LIMIT 1 OFFSET ABS(RANDOM()) % MAX((SELECT count(id) FROM quotes), 1)`)
	if err != nil {
		return nil, err
	}
	prepRatingRandom, err := db.PrepareContext(ctx,
		`SELECT id, date, rating, text FROM quotes WHERE rating>=? LIMIT 1 OFFSET ABS(RANDOM()) % MAX((SELECT count(id) FROM quotes WHERE rating>=?), 1)`)
	if err != nil {
		return nil, err
	}
	prepCities, err := db.PrepareContext(ctx,
		`SELECT city, country FROM cities WHERE alias=?`)
	if err != nil {
		return nil, err
	}
	prepIgnDomain, err := db.PrepareContext(ctx,
		`SELECT domain from ignored_domains where domain=?`)
	if err != nil {
		return nil, err
	}
	return map[dbStmt]*sql.Stmt{
		fetchQuote:        prepQ,
		fetchRandomQuote:  prepRandom,
		fetchRandomRating: prepRatingRandom,
		fetchCity:         prepCities,
		ignoredDomain:     prepIgnDomain,
	}, nil
}
