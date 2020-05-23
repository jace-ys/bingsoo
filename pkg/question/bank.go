package question

import (
	"context"
	"math/rand"

	"github.com/jmoiron/sqlx"

	"github.com/jace-ys/bingsoo/pkg/postgres"
)

type Question struct {
	ID    int
	Value string
}

type Bank struct {
	database *postgres.Client
}

func NewBank(database *postgres.Client) *Bank {
	return &Bank{
		database: database,
	}
}

func (b *Bank) List(ctx context.Context) ([]*Question, error) {
	var questions []*Question
	err := b.database.Transact(ctx, func(tx *sqlx.Tx) error {
		query := `
		SELECT q.id, q.value
		FROM questions AS q
		`
		rows, err := tx.QueryxContext(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var question Question
			if err := rows.StructScan(&question); err != nil {
				return err
			}
			questions = append(questions, &question)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (b *Bank) NewQuestionSet(ctx context.Context, num int) ([]*Question, error) {
	questions, err := b.List(ctx)
	if err != nil {
		return nil, err
	}

	if len(questions) < num {
		num = len(questions)
	}

	unique := make(map[*Question]struct{})
	for len(unique) < num {
		question := questions[rand.Intn(len(questions))]
		unique[question] = struct{}{}
	}

	var set []*Question
	for question := range unique {
		set = append(set, question)
	}

	return set, nil
}
