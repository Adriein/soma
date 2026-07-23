package coach

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rotisserie/eris"
)

var ErrAssessmentNotFound = eris.New("Assessment not found")

type AssessmentRepository interface {
	Save(ctx context.Context, assessment *Assessment) error
	GetByID(ctx context.Context, ID int) (*Assessment, error)
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

func (r *PgAssessmentRepository) GetByID(ctx context.Context, ID int) (*Assessment, error) {
	query := `
		SELECT
			soa_id,
			soa_text,
			soa_date_add
		FROM so_assessment
		WHERE soa_id = $1
	`

	var assessment Assessment

	if err := r.connection.QueryRowContext(ctx, query, ID).Scan(
		&assessment.ID,
		&assessment.Content,
		&assessment.DateAdd,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, eris.Wrapf(ErrAssessmentNotFound, "No assessment found with id: %d", ID)
		}

		return nil, eris.Wrap(err, "Error querying assessment by id")
	}

	return &assessment, nil

}
