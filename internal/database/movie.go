package database

import (
	"context"
	"database/sql"
	"pilem/internal/data"
	"time"

	"github.com/lib/pq"
)

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *data.Movie) error {
	query := `
	INSERT INTO movies (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version
	`

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}
