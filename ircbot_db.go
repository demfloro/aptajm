package main

import (
	"context"
	"database/sql"
)

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
