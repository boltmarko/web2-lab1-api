package main

import (
	"api/middleware"

	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/skip2/go-qrcode"
)

type GenerateRequest struct {
	Vatin     string `json:"vatin"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Ticket struct {
	ID        uuid.UUID `json:"id"`
	Vatin     string    `json:"vatin"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	CreatedAt time.Time `json:"createdAt"`
}

type TotalTickets struct {
	TotalTickets int `json:"total_tickets"`
}

func createTicket(ticket Ticket) error {
	db, err := sql.Open("postgres", os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO tickets (id, vatin, first_name, last_name, created_at) VALUES ($1, $2, $3, $4, $5)",
		ticket.ID, ticket.Vatin, ticket.FirstName, ticket.LastName, ticket.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func getTotalTickets() (TotalTickets, error) {
	db, err := sql.Open("postgres", os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		return TotalTickets{}, err
	}

	var totalTickets TotalTickets
	err = db.QueryRow("SELECT COUNT(*) FROM tickets").Scan(&totalTickets.TotalTickets)
	if err != nil {
		return TotalTickets{}, err
	}

	return totalTickets, nil
}

func getTotalTicketsForVatin(vatin string) (int, error) {
	db, err := sql.Open("postgres", os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		return 0, err
	}

	var totalTickets int
	err = db.QueryRow("SELECT COUNT(*) FROM tickets WHERE vatin = $1", vatin).Scan(&totalTickets)
	if err != nil {
		return 0, err
	}

	return totalTickets, nil
}

func getTicket(id uuid.UUID) (Ticket, error) {
	db, err := sql.Open("postgres", os.Getenv("PG_CONNECTION_STRING"))
	if err != nil {
		return Ticket{}, err
	}

	var ticket Ticket
	err = db.QueryRow("SELECT id, vatin, first_name, last_name, created_at FROM tickets WHERE id = $1", id).Scan(&ticket.ID, &ticket.Vatin, &ticket.FirstName, &ticket.LastName, &ticket.CreatedAt)
	if err != nil {
		return Ticket{}, err
	}

	return ticket, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, proceeding...")
	}

	clientURL := os.Getenv("CLIENT_URL")

	router := http.NewServeMux()

	router.Handle("GET /api/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "API is up and running"}`))
	}))

	router.Handle("GET /api/tickets/total", middleware.EnsureValidToken()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			totalTickets, err := getTotalTickets()
			if err != nil {
				log.Println(err)
				http.Error(w, "Error fetching total tickets", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(totalTickets)
		})))

	router.Handle("GET /api/tickets/{id}", middleware.EnsureValidToken()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idStr := r.PathValue("id")
			id, err := uuid.Parse(idStr)
			if err != nil {
				log.Printf("Invalid ticket ID: %s", idStr)
				http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
				return
			}

			ticket, err := getTicket(id)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error fetching ticket", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(ticket)
		})))

	router.Handle("POST /api/generate", middleware.EnsureValidToken()(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			decoder := json.NewDecoder(r.Body)
			var request GenerateRequest
			if err := decoder.Decode(&request); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			totalTickets, err := getTotalTicketsForVatin(request.Vatin)
			if err != nil {
				log.Println(err)
				http.Error(w, "Error checking total tickets", http.StatusInternalServerError)
				return
			}

			if totalTickets >= 3 {
				http.Error(w, "Maximum number of tickets reached", http.StatusBadRequest)
				return
			}

			ticket := Ticket{
				ID:        uuid.New(),
				Vatin:     request.Vatin,
				FirstName: request.FirstName,
				LastName:  request.LastName,
				CreatedAt: time.Now(),
			}

			if err := createTicket(ticket); err != nil {
				http.Error(w, "Error creating ticket", http.StatusInternalServerError)
				return
			}

			url := clientURL + "/" + ticket.ID.String()
			png, err := qrcode.Encode(url, qrcode.Medium, 256)
			if err != nil {
				http.Error(w, "Error generating QR code", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			w.Write(png)
		})))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server started on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
