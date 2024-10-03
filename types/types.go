package types

import (
	"time"

	"github.com/google/uuid"
)

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
