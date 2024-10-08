package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"pilem/internal/database"
)

type Server struct {
	port int

	db database.Models
}

func NewServer(db *sql.DB) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		db: database.NewModels(db),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
