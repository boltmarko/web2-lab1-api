CREATE TABLE tickets (
	id UUID PRIMARY KEY,
	vatin TEXT NOT NULL,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
)
