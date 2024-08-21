package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"pilem/helper"
	"pilem/internal/data"
	"pilem/internal/database"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/lib/pq"
)

func TestHandler(t *testing.T) {
	s := &Server{}
	server := httptest.NewServer(http.HandlerFunc(s.HelloWorldHandler))
	defer server.Close()
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}
	defer resp.Body.Close()
	// Assertions
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
	expected := "{\"message\":\"Hello World\"}"
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}
	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}

func TestCreateMovieHandler_ReturnMovieWhenPost(t *testing.T) {
	t.Parallel()

	db, _ := helper.NewSQLMock(t, func(mock sqlmock.Sqlmock) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "version"}).AddRow(1, time.Now(), 1)
		mock.ExpectQuery("").WillReturnRows(rows)
	})

	s := &Server{
		db: database.NewModels(db),
	}
	server := httptest.NewServer(http.HandlerFunc(s.CreateMovieHandler))
	defer server.Close()

	movie := data.Movie{
		Title:   "Overlord",
		Year:    2024,
		Runtime: 135,
		Genres:  []string{"Action", "Adventure", "Fantasy"},
	}

	expJson, err := helper.AnyToJSON(movie)
	if err != nil {
		t.Fatalf("error create Expected JSON. Err:%v", err)
	}

	resp, err := http.Post(server.URL, "application/json", strings.NewReader(expJson))
	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}

	movie.ID = 1
	movie.Version = 1
	want, err := helper.AnyToJSON(helper.Envelope{"movie": movie})
	if err != nil {
		t.Fatalf("Cant write want to json(AnyToJSON). Err:%v", err)
	}
	got := string(body)

	// Assertions
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got), "movie not equal")
	}

}

func TestGetMovieHandler_ReturnMovieByID(t *testing.T) {
	t.Parallel()

	now := time.Now().Format(time.RFC3339)
	timeWant, err := time.Parse(time.RFC3339, now)
	if err != nil {
		t.Fatalf("can't parse time. Err:%v", err)
	}

	want := data.Movie{
		ID:        1,
		CreatedAt: timeWant,
		Title:     "overlord",
		Year:      2024,
		Runtime:   135,
		Genres:    []string{"Adventure", "Action"},
		Version:   2,
	}
	db, _ := helper.NewSQLMock(t, func(mock sqlmock.Sqlmock) {
		rows := sqlmock.NewRows([]string{"id", "created_at", "title", "year", "runtime", "genres", "version"}).
			AddRow(want.ID, want.CreatedAt, want.Title, want.Year, want.Runtime, pq.Array(want.Genres), want.Version)
		mock.ExpectQuery("").WillReturnRows(rows)
	})

	s := &Server{
		db: database.NewModels(db),
	}

	r := httptest.NewRequest(http.MethodGet, "/movies/2", nil)
	r.SetPathValue("id", "2")
	w := httptest.NewRecorder()
	s.GetMovieHandler(w, r)

	got, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatalf("cant io.ReadAll. Err: %v", err)
	}

	wantStr, err := helper.AnyToJSON(helper.Envelope{"movie": want})
	if err != nil {
		t.Fatalf("cant convert struct to json. Err:%v", err)
	}

	if !cmp.Equal(wantStr, string(got)) {
		t.Error(cmp.Diff(wantStr, string(got)), "json not equal")
		t.Log(wantStr)
		t.Log(string(got))
	}

}

func TestDeleteMovieHandler_ReturnNothing(t *testing.T) {
	t.Parallel()

	id := int64(5)
	db, _ := helper.NewSQLMock(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec("DELETE FROM movies WHERE id").WithArgs(id).WillReturnResult(sqlmock.NewResult(0, id))
	})

	m := database.NewModels(db)
	err := m.Movies.Delete(id)
	if err != nil {
		t.Fatalf("Can't delete movie id: %d, Err: %v", id, err)
	}

}

func TestUpdateMovieHandler_ReturnMovie(t *testing.T) {
	t.Parallel()

	want := data.Movie{
		ID:      2,
		Title:   "overlord update",
		Year:    2024,
		Runtime: 135,
		Genres:  []string{"Adventure", "Fantasy"},
		Version: 2,
	}

	db, _ := helper.NewSQLMock(t, func(mock sqlmock.Sqlmock) {
		queryUpdate := `
		UPDATE movies
		SET (.+) 
		RETURNING version
		`
		mock.ExpectQuery("").WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "title", "year", "runtime", "genres", "version"}).AddRow(want.ID, want.CreatedAt, want.Title, want.Year, want.Runtime, pq.Array(want.Genres), want.Version))
		rows := sqlmock.NewRows([]string{"version"}).AddRow(want.Version)
		mock.ExpectQuery(queryUpdate).WithArgs(want.Title, want.Year, want.Runtime, pq.Array(want.Genres), want.ID, want.Version).
			WillReturnRows(rows)
	})

	m := database.NewModels(db)
	s := Server{db: m}

	setupJson, err := helper.AnyToJSON(want)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPut, "/movie/2", strings.NewReader(setupJson))
	r.SetPathValue("id", strconv.FormatInt(want.ID, 10))
	w := httptest.NewRecorder()

	s.UpdateMovieHandler(w, r)

	gotByte, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Fatal(err)
	}

	wantStr, err := helper.AnyToJSON(helper.Envelope{"movie": want})
	if err != nil {
		t.Fatal(err)
	}

	if wantStr != string(gotByte) {
		t.Log("want", wantStr)
		t.Log("got", string(gotByte))
		t.Error(cmp.Diff(wantStr, string(gotByte)))
	}

}
