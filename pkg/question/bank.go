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

type QuestionSet map[string]int

func (b *Bank) NewQuestionSet(ctx context.Context, num int) (QuestionSet, error) {
	questions, err := b.List(ctx)
	if err != nil {
		return nil, err
	}

	if len(questions) < num {
		num = len(questions)
	}

	set := make(QuestionSet, num)
	for i := 0; i < num; i++ {
		question := questions[rand.Intn(len(questions))]
		set[question.Value] = 0
	}

	return set, nil
}
