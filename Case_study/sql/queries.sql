-- Insert a team
INSERT INTO teams (id, name, strength) VALUES (?, ?, ?);

-- Insert a match
INSERT INTO matches (id, home_team_id, away_team_id, week) VALUES (?, ?, ?, ?);

-- Update match result
UPDATE matches SET home_goals = ?, away_goals = ?, played = 1 WHERE id = ?;

-- Update league table entry
UPDATE league_table SET points = ?, goals_for = ?, goals_against = ?, goal_difference = ?, matches_played = ? WHERE team_id = ?;

-- Select league standings
SELECT t.name, l.points, l.goals_for, l.goals_against, l.goal_difference, l.matches_played
FROM league_table l JOIN teams t ON l.team_id = t.id
ORDER BY l.points DESC, l.goal_difference DESC, l.goals_for DESC;

-- Select matches for a week
SELECT * FROM matches WHERE week = ?; 