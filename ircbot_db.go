package main

import (
	"context"
	"database/sql"
)

func initStmts(ctx context.Context, db *sql.DB) (map[dbStmt]*sql.Stmt, error) {
	fetchQ, err := db.PrepareContext(ctx, `SELECT id, date, rating, text FROM quotes WHERE id=?`)
	if err != nil {
		return nil, err
	}
	fetchRandom, err := db.PrepareContext(ctx, `SELECT id, date, rating, text FROM quotes LIMIT 1 OFFSET ABS(RANDOM()) % MAX((SELECT count(id) FROM quotes), 1)`)
	if err != nil {
		return nil, err
	}
	fetchC, err := db.PrepareContext(ctx, `SELECT city, country FROM cities WHERE alias=?`)
	if err != nil {
		return nil, err
	}
	ignDomain, err := db.PrepareContext(ctx, `SELECT domain from ignored_domains where domain=?`)
	if err != nil {
		return nil, err
	}
	return map[dbStmt]*sql.Stmt{
		fetchQuote:       fetchQ,
		fetchRandomQuote: fetchRandom,
		fetchCity:        fetchC,
		ignoredDomain:    ignDomain,
	}, nil
}
