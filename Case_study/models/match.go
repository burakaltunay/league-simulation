package models

import (
	"Case_study/storage"
	"math/rand"
	"database/sql"
)

// Match represents a football match between two teams
 type Match struct {
    ID         int
    HomeTeamID int
    AwayTeamID int
    HomeGoals  sql.NullInt64
    AwayGoals  sql.NullInt64
    Week       int
    Played     bool
}

// MatchSimulator defines logic for simulating a match
 type MatchSimulator interface {
    SimulateMatch(home Team, away Team) (homeGoals int, awayGoals int)
}

// BasicMatchSimulator simulates matches based on team strengths
 type BasicMatchSimulator struct {}

// SimulateMatch returns simulated goals for home and away teams
func (b BasicMatchSimulator) SimulateMatch(home Team, away Team) (int, int) {
    // Simple probabilistic model: higher strength = more likely to score
    // Home advantage: +10% strength
    homeStrength := float64(home.Strength) * 1.1
    awayStrength := float64(away.Strength)
    totalStrength := homeStrength + awayStrength
    
    // Expected goals: scale to 0-3 goals per team
    homeGoals := int((homeStrength/totalStrength)*3 + 0.5) // round
    awayGoals := int((awayStrength/totalStrength)*3 + 0.5)
    
    // Add some randomness
    if homeGoals > 0 && rand.Intn(4) == 0 { homeGoals-- }
    if awayGoals > 0 && rand.Intn(4) == 0 { awayGoals-- }
    if rand.Intn(10) == 0 { homeGoals++ }
    if rand.Intn(10) == 0 { awayGoals++ }
    if homeGoals < 0 { homeGoals = 0 }
    if awayGoals < 0 { awayGoals = 0 }
    return homeGoals, awayGoals
} 

// SQLiteMatchRepository implements DB operations for matches
 type SQLiteMatchRepository struct{}

func (r SQLiteMatchRepository) GetMatchesByWeek(week int) ([]Match, error) {
	db := storage.GetDB()
	rows, err := db.Query("SELECT id, home_team_id, away_team_id, home_goals, away_goals, week, played FROM matches WHERE week = ?", week)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var matches []Match
	for rows.Next() {
		var m Match
		if err := rows.Scan(&m.ID, &m.HomeTeamID, &m.AwayTeamID, &m.HomeGoals, &m.AwayGoals, &m.Week, &m.Played); err != nil {
			return nil, err
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func (r SQLiteMatchRepository) UpdateMatch(m Match) error {
	db := storage.GetDB()
	_, err := db.Exec("UPDATE matches SET home_goals = ?, away_goals = ?, played = ? WHERE id = ?",
		nullableInt(m.HomeGoals), nullableInt(m.AwayGoals), m.Played, m.ID)
	return err
}

func (r SQLiteMatchRepository) CreateMatch(m Match) error {
	db := storage.GetDB()
	_, err := db.Exec("INSERT INTO matches (id, home_team_id, away_team_id, week) VALUES (?, ?, ?, ?)", m.ID, m.HomeTeamID, m.AwayTeamID, m.Week)
	return err
}

// Helper to handle sql.NullInt64 for nullable fields
func nullableInt(n sql.NullInt64) interface{} {
	if n.Valid {
		return n.Int64
	}
	return nil
} 