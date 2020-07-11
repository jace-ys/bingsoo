CREATE TABLE IF NOT EXISTS sessions (
  id UUID NOT NULL,
  team_id TEXT NOT NULL,
  question_votes JSONB NOT NULL,
  selected_question TEXT NOT NULL,
  responses JSONB NOT NULL,
  PRIMARY KEY (id),
  FOREIGN KEY (team_id) REFERENCES teams (team_id)
);
