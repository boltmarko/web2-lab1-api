package db

import (
	"api/types"
	"database/sql"

	"github.com/google/uuid"
)

type PostgresDatabase struct {
	*sql.DB
}

func (db PostgresDatabase) CreateTicket(ticket types.Ticket) error {
	_, err := db.Exec("INSERT INTO tickets (id, vatin, first_name, last_name, created_at) VALUES ($1, $2, $3, $4, $5)",
		ticket.ID, ticket.Vatin, ticket.FirstName, ticket.LastName, ticket.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (db PostgresDatabase) GetTotalTickets() (types.TotalTickets, error) {
	var totalTickets types.TotalTickets
	err := db.QueryRow("SELECT COUNT(*) FROM tickets").Scan(&totalTickets.TotalTickets)
	if err != nil {
		return types.TotalTickets{}, err
	}

	return totalTickets, nil
}

func (db PostgresDatabase) GetTotalTicketsForVatin(vatin string) (int, error) {
	var totalTickets int
	err := db.QueryRow("SELECT COUNT(*) FROM tickets WHERE vatin = $1", vatin).Scan(&totalTickets)
	if err != nil {
		return 0, err
	}

	return totalTickets, nil
}

func (db PostgresDatabase) GetTicket(id uuid.UUID) (types.Ticket, error) {
	var ticket types.Ticket
	err := db.QueryRow("SELECT id, vatin, first_name, last_name, created_at FROM tickets WHERE id = $1", id).Scan(&ticket.ID, &ticket.Vatin, &ticket.FirstName, &ticket.LastName, &ticket.CreatedAt)
	if err != nil {
		return types.Ticket{}, err
	}

	return ticket, nil
}

func NewPostgresDatabase(connectionString string) (PostgresDatabase, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return PostgresDatabase{}, err
	}

	return PostgresDatabase{db}, nil
}
