package team

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/jace-ys/bingsoo/pkg/postgres"
)

var (
	ErrTeamNotFound = errors.New("team not found")
)

type Team struct {
	ID                  uuid.UUID
	CreatedAt           time.Time
	TeamID              string
	TeamDomain          string
	AccessToken         string
	ChannelID           string
	SessionDurationMins int
}

type Registry struct {
	database *postgres.Client
}

func NewRegistry(database *postgres.Client) *Registry {
	return &Registry{
		database: database,
	}
}

func (r *Registry) Get(ctx context.Context, teamID string) (*Team, error) {
	var team Team
	err := r.database.Transact(ctx, func(tx *sqlx.Tx) error {
		query := `
		SELECT t.id, t.created_at, t.team_id, t.team_domain, t.access_token, t.channel_id, t.session_duration_mins
		FROM teams AS t
		WHERE t.team_id=$1
		`
		row := tx.QueryRowxContext(ctx, query, teamID)
		return row.StructScan(&team)
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrTeamNotFound
		default:
			return nil, err
		}
	}
	return &team, nil
}
