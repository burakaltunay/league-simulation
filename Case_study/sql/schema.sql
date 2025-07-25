-- Teams table
CREATE TABLE teams (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    strength INTEGER NOT NULL
);

-- Matches table
CREATE TABLE matches (
    id INTEGER PRIMARY KEY,
    home_team_id INTEGER NOT NULL,
    away_team_id INTEGER NOT NULL,
    home_goals INTEGER,
    away_goals INTEGER,
    week INTEGER NOT NULL,
    played BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY(home_team_id) REFERENCES teams(id),
    FOREIGN KEY(away_team_id) REFERENCES teams(id)
);

-- League table (standings)
CREATE TABLE league_table (
    team_id INTEGER PRIMARY KEY,
    points INTEGER NOT NULL,
    goals_for INTEGER NOT NULL,
    goals_against INTEGER NOT NULL,
    goal_difference INTEGER NOT NULL,
    matches_played INTEGER NOT NULL,
    FOREIGN KEY(team_id) REFERENCES teams(id)
); 