package coach

import (
	"context"
	"database/sql"

	"github.com/rotisserie/eris"
)

type AssessmentRepository interface {
	Save(ctx context.Context, assessment *Assessment) error
}

type PgAssessmentRepository struct {
	connection *sql.DB
}

func NewPgAssessmentRepository(c *sql.DB) *PgAssessmentRepository {
	return &PgAssessmentRepository{connection: c}
}

func (r *PgAssessmentRepository) Save(ctx context.Context, assessment *Assessment) error {
	query := `
		INSERT INTO so_assessment (
			soa_text,
			soa_date_add,
			soa_date_upd
		)
		VALUES ($1, TIMEZONE('UTC', NOW()), TIMEZONE('UTC', NOW()));
	`

	_, err := r.connection.ExecContext(
		ctx,
		query,
		assessment.Content,
	)

	if err != nil {
		return eris.Wrap(err, "Error saving assessment")
	}

	return nil
}
