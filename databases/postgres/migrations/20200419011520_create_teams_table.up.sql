CREATE TABLE IF NOT EXISTS teams (
  id UUID DEFAULT uuid_generate_v4(),
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  team_id TEXT UNIQUE NOT NULL,
  team_domain TEXT UNIQUE NOT NULL,
  channel_id TEXT NOT NULL,
  session_duration_mins INTEGER,
  PRIMARY KEY (id)
);
