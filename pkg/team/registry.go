package team

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jace-ys/go-library/postgres"
	"github.com/jmoiron/sqlx"
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
	ParticipantQuota    int
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
		SELECT id, created_at, team_id, team_domain, access_token, channel_id, session_duration_mins, participant_quota
		FROM teams
		WHERE team_id=$1
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
