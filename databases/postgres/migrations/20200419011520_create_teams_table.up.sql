CREATE TABLE IF NOT EXISTS teams (
  id UUID DEFAULT uuid_generate_v4(),
  team_id TEXT UNIQUE NOT NULL,
  team_domain TEXT UNIQUE NOT NULL,
  access_token TEXT NOT NULL,
  channel_id TEXT NOT NULL,
  session_duration_mins INTEGER NOT NULL,
  participant_quota INTEGER NOT NULL,
  created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);
