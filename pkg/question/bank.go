package question

import (
	"math/rand"
	"time"

	"github.com/jace-ys/bingsoo/pkg/postgres"
)

type Question struct {
	ID    int
	Value string
}

type Bank struct {
	database  *postgres.Client
	questions []*Question
}

func NewBank(database *postgres.Client) *Bank {
	return &Bank{
		database: database,
		questions: []*Question{
			&Question{Value: "What's your favourite book?"},
			&Question{Value: "What's your favourite movie?"},
			&Question{Value: "Where's your dream destination?"},
			&Question{Value: "Tabs or spaces?"},
			&Question{Value: "Tell us a fun fact about yourself."},
			&Question{Value: "Favourite ice cream flavour?"},
			&Question{Value: "What's the most recent TV show you've watched?"},
			&Question{Value: "Favourite quote of all time?"},
			&Question{Value: "What languages can you speak?"},
			&Question{Value: "What genre of music do you listen to?"},
		},
	}
}

func (b *Bank) NewQuestionSet(num int) []*Question {
	rand.Seed(time.Now().Unix())
	questions := make([]*Question, num)
	for i := 0; i < num; i++ {
		questions[i] = b.questions[rand.Intn(len(b.questions))]
	}
	return questions
}
