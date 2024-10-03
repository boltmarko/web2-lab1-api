package main

import (
	"api/types"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

type GenerateRequest struct {
	Vatin     string `json:"vatin"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func handleGetTotalTickets(db Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		totalTickets, err := db.GetTotalTickets()
		if err != nil {
			log.Println(err)
			http.Error(w, "Error fetching total tickets", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(totalTickets)
	}
}

func handleGetTicketById(db Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Printf("Invalid ticket ID: %s", idStr)
			http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
			return
		}

		ticket, err := db.GetTicket(id)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error fetching ticket", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ticket)
	}
}

func handleGenerateTicket(db Database, c *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var request GenerateRequest
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		totalTickets, err := db.GetTotalTicketsForVatin(request.Vatin)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error checking total tickets", http.StatusInternalServerError)
			return
		}

		if totalTickets >= 3 {
			http.Error(w, "Maximum number of tickets reached", http.StatusBadRequest)
			return
		}

		ticket := types.Ticket{
			ID:        uuid.New(),
			Vatin:     request.Vatin,
			FirstName: request.FirstName,
			LastName:  request.LastName,
			CreatedAt: time.Now(),
		}

		if err := db.CreateTicket(ticket); err != nil {
			http.Error(w, "Error creating ticket", http.StatusInternalServerError)
			return
		}

		url := c.ClientURL + "/" + ticket.ID.String()
		png, err := qrcode.Encode(url, qrcode.Medium, 256)
		if err != nil {
			http.Error(w, "Error generating QR code", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(png)
	}
}
