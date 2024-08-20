package database_test

import (
	"pilem/helper"
	"pilem/internal/data"
	"pilem/internal/database"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
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

func TestMovieGet_GetMovie(t *testing.T) {
	t.Parallel()

	now, err := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("cant parse time. Err:%v", err)
	}

	want := data.Movie{
		ID:        2,
		CreatedAt: now,
		Title:     "overlord",
		Year:      2024,
		Runtime:   135,
		Genres:    []string{"Adventure", "Fantasy"},
		Version:   4,
	}

	db, mock := helper.NewSQLMock(t, func(mock sqlmock.Sqlmock) {
		query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id 
		`
		rows := sqlmock.NewRows([]string{"id", "created_at", "title", "year", "runtime", "genres", "version"}).
			AddRow(want.ID, want.CreatedAt, want.Title, want.Year, want.Runtime, pq.Array(want.Genres), want.Version)
		mock.ExpectQuery(query).WithArgs(want.ID).WillReturnRows(rows)
	})

	m := database.NewModels(db)
	got, err := m.Movies.Get(want.ID)
	if err != nil {
		t.Fatalf("can't get movie. Err:%v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("expectation query not met")
	}

	if want.ID != got.ID {
		t.Errorf("want id %d got %d", want.ID, got.ID)
	}

	if want.CreatedAt != got.CreatedAt {
		t.Errorf("want created at %v got %v", want.CreatedAt, got.CreatedAt)
	}

	if want.Title != got.Title {
		t.Errorf("want title %s got %s", want.Title, got.Title)
	}

	if want.Year != got.Year {
		t.Errorf("want year %d got %d", want.Year, got.Year)
	}

	if want.Runtime != got.Runtime {
		t.Errorf("want title %d got %d", want.Runtime, got.Runtime)
	}

	if !cmp.Equal(want.Genres, got.Genres) {
		t.Errorf("want genres %v got %v", want.Genres, got.Genres)
	}

	if want.Version != got.Version {
		t.Errorf("want version %d got %d", want.Version, got.Version)
	}
}

func TestMovieDelete_DeleteMovie(t *testing.T) {
	t.Parallel()

	id := int64(2)
	db, mock := helper.NewSQLMock(t, func(mock sqlmock.Sqlmock) {
		query := `DELETE FROM movies WHERE id `
		mock.ExpectExec(query).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, id))
	})

	m := database.NewModels(db)

	err := m.Movies.Delete(id)
	if err != nil {
		t.Fatalf("Can't delete movie id: %d, Err: %v", id, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal("query not as expected", err)
	}

}
