package bingsoo

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type team struct {
	ID                  *uuid.UUID
	CreatedAt           *time.Time
	TeamID              string
	TeamDomain          string
	BotToken            string
	ChannelID           string
	SessionDurationMins int
}

func (bot *BingsooBot) getTeam(ctx context.Context, teamID string) (*team, error) {
	var team team
	err := bot.database.Transact(ctx, func(tx *sqlx.Tx) error {
		query := `
		SELECT t.id, t.team_id, t.team_domain, t.channel_id, t.created_at, t.session_duration_mins
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
