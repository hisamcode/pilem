package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"pilem/helper"
	"pilem/internal/data"
	"pilem/internal/database"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
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
