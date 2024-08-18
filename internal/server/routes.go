package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"pilem/helper"
	"pilem/internal/data"
	"pilem/internal/database"
	"strconv"
)

func (s *Server) RegisterRoutes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.HelloWorldHandler)

	mux.HandleFunc("/health", s.healthHandler)

	return mux
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	// jsonResp, err := json.Marshal(s.db.Health())

	// if err != nil {
	// 	log.Fatalf("error handling JSON marshal. Err: %v", err)
	// }

	// _, _ = w.Write(jsonResp)
}

func (s *Server) CreateMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := helper.ReadJSON(w, r, &input)
	if err != nil {
		helper.BadRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	err = s.db.Movies.Insert(movie)
	if err != nil {
		helper.ServerErrorResponse(w, r, err)
		return
	}

	err = helper.WriteJSON(w, http.StatusCreated, helper.Envelope{"movie": movie}, nil)
	if err != nil {
		helper.ServerErrorResponse(w, r, err)
	}

}

func (s *Server) GetMovieHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		helper.NotFoundResponse(w, r, err)
		return
	}

	movie, err := s.db.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			helper.NotFoundResponse(w, r, err)
		default:
			helper.ServerErrorResponse(w, r, err)
		}
		return
	}

	err = helper.WriteJSON(w, http.StatusOK, helper.Envelope{"movie": movie}, nil)
	if err != nil {
		helper.ServerErrorResponse(w, r, err)
	}

}
