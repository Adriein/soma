package customer

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rotisserie/eris"
)

var ErrCustomerNotFound = eris.New("Customer not found")

type CustomerRepository interface {
	Save(ctx context.Context, customer *Customer) error
	GetByTelegramChatID(ctx context.Context, chatID int64) (*Customer, error)
}

type PgCustomerRepsitory struct {
	connection *sql.DB
}

func NewPgCustomerRepository(c *sql.DB) *PgCustomerRepsitory {
	return &PgCustomerRepsitory{connection: c}
}

func (r *PgCustomerRepsitory) Save(ctx context.Context, customer *Customer) error {
	query := `
		INSERT INTO so_users (
			sou_telegram_chat_id,
			sou_name,
			sou_token,
			sou_token_secret,
			sou_token_verifier,
			sou_date_add,
			sou_date_upd
		)
		VALUES ($1, $2, $3, $4, $5, TIMEZONE('UTC', NOW()), TIMEZONE('UTC', NOW()))
	`

	_, err := r.connection.ExecContext(
		ctx,
		query,
		customer.TelegramChatID,
		customer.Name,
		customer.Token,
		customer.TokenSecret,
		customer.TokenVerifier,
	)

	if err != nil {
		return eris.Wrap(err, "Error saving customer")
	}

	return nil
}

func (r *PgCustomerRepsitory) GetByTelegramChatID(ctx context.Context, chatID int64) (*Customer, error) {
	query := `
		SELECT
			sou_id,
			sou_name,
			sou_telegram_chat_id,
			sou_token,
			sou_token_secret,
			sou_token_verifier
		FROM so_users
		WHERE sou_telegram_chat_id = $1
	`

	var customer Customer

	if err := r.connection.QueryRowContext(ctx, query, chatID).Scan(
		&customer.ID,
		&customer.Name,
		&customer.TelegramChatID,
		&customer.Token,
		&customer.TokenSecret,
		&customer.TokenVerifier,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, eris.Wrapf(ErrCustomerNotFound, "No customer found with the chat id: %d", chatID)
		}

		return nil, eris.Wrap(err, "Error querying customer by telegram chat id")
	}

	return &customer, nil
}
