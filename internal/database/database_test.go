package database_test

import (
	"pilem/helper"
	"pilem/internal/data"
	"pilem/internal/database"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func TestMovieInsert_GetIDCreatedAtVersionWhenInsertingNewMovie(t *testing.T) {
	t.Parallel()

	d := data.Movie{
		Title:   "Overlord",
		Year:    2024,
		Runtime: 135,
		Genres:  []string{"Adventure", "Action", "Fantasy"},
	}

	want := data.Movie{
		ID:        1,
		CreatedAt: time.Now(),
		Version:   1,
	}

	db, mock := helper.NewSQLMock(t, func(mock sqlmock.Sqlmock) {

		query := `INSERT INTO movies (.+) RETURNING id, created_at, version`
		rows := sqlmock.NewRows([]string{"id", "created_at", "version"}).
			AddRow(want.ID, want.CreatedAt, want.Version)

		mock.ExpectQuery(query).
			WithArgs(d.Title, d.Year, d.Runtime, pq.Array(d.Genres)).
			WillReturnRows(rows)
	})

	m := database.MovieModel{DB: db}
	movie := data.Movie{
		Title:   "Overlord",
		Year:    2024,
		Runtime: 135,
		Genres:  []string{"Adventure", "Action", "Fantasy"},
	}

	err := m.Insert(&movie)
	if err != nil {
		t.Fatalf("can't insert movie. Err: %s", err)
	}

	got := movie

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled query expectations: %s\n", err)
	}

	if want.ID != movie.ID {
		t.Errorf("want ID %d got %d\n", want.ID, movie.ID)
	}

	if want.CreatedAt != movie.CreatedAt {
		t.Errorf("want created at %s got %s\n", want.CreatedAt.Format(time.RFC3339), movie.CreatedAt.Format(time.RFC3339))
	}

	if want.Version != movie.Version {
		t.Errorf("want version %d got %d\n", want.ID, got.ID)
	}

}
