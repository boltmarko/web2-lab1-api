package main

import (
	"api/db"
	"api/middleware"
	"api/types"

	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type Config struct {
	ClientURL string
}

type Database interface {
	CreateTicket(ticket types.Ticket) error
	GetTotalTickets() (types.TotalTickets, error)
	GetTotalTicketsForVatin(vatin string) (int, error)
	GetTicket(id uuid.UUID) (types.Ticket, error)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, proceeding...")
	}

	config := Config{ClientURL: os.Getenv("CLIENT_URL")}

	db, err := db.NewPostgresDatabase(os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()

	router.Handle("GET /api/tickets/total", middleware.EnsureValidToken()(handleGetTotalTickets(db)))
	router.Handle("GET /api/tickets/{id}", middleware.EnsureValidToken()(handleGetTicketById(db)))
	router.Handle("POST /api/generate", middleware.EnsureValidToken()(handleGenerateTicket(db, &config)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server started on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
