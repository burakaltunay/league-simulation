package models

import (
	"Case_study/storage"
	"database/sql"
)

// Team represents a football team in the league
 type Team struct {
    ID            int
    Name          string
    Strength      int
    Points        int
    GoalsFor      int
    GoalsAgainst  int
    GoalDifference int
    MatchesPlayed int
}

// TeamRepository defines DB operations for teams
 type TeamRepository interface {
    GetAllTeams() ([]Team, error)
    GetTeamByID(id int) (Team, error)
    UpdateTeam(team Team) error
    CreateTeam(team Team) error
}

// SQLiteTeamRepository implements TeamRepository using SQLite
 type SQLiteTeamRepository struct{}

func (r SQLiteTeamRepository) GetAllTeams() ([]Team, error) {
	db := storage.GetDB()
	rows, err := db.Query("SELECT id, name, strength FROM teams")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var teams []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name, &t.Strength); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}

func (r SQLiteTeamRepository) GetTeamByID(id int) (Team, error) {
	db := storage.GetDB()
	row := db.QueryRow("SELECT id, name, strength FROM teams WHERE id = ?", id)
	var t Team
	if err := row.Scan(&t.ID, &t.Name, &t.Strength); err != nil {
		if err == sql.ErrNoRows {
			return t, nil
		}
		return t, err
	}
	return t, nil
}

func (r SQLiteTeamRepository) UpdateTeam(team Team) error {
	db := storage.GetDB()
	_, err := db.Exec("UPDATE teams SET name = ?, strength = ? WHERE id = ?", team.Name, team.Strength, team.ID)
	return err
}

func (r SQLiteTeamRepository) CreateTeam(team Team) error {
	db := storage.GetDB()
	_, err := db.Exec("INSERT INTO teams (id, name, strength) VALUES (?, ?, ?)", team.ID, team.Name, team.Strength)
	return err
} 