package models

import (
	"Case_study/storage"
	"sort"
)

// LeagueTableEntry represents a row in the league table
 type LeagueTableEntry struct {
    TeamID         int
    TeamName       string
    Points         int
    GoalsFor       int
    GoalsAgainst   int
    GoalDifference int
    MatchesPlayed  int
}

// League represents the league state
 type League struct {
    Teams   []Team
    Matches []Match
    Week    int
}

// LeagueRepository defines DB operations for the league
 type LeagueRepository interface {
    GetLeague() (League, error)
    UpdateLeague(league League) error
}

// SQLiteLeagueRepository implements LeagueRepository using SQLite
 type SQLiteLeagueRepository struct{}

func (r SQLiteLeagueRepository) GetLeague() (League, error) {
	db := storage.GetDB()
	var league League
	// Get teams
	teamRepo := SQLiteTeamRepository{}
	teams, err := teamRepo.GetAllTeams()
	if err != nil {
		return league, err
	}
	league.Teams = teams
	// Get matches
	rows, err := db.Query("SELECT id, home_team_id, away_team_id, home_goals, away_goals, week, played FROM matches")
	if err != nil {
		return league, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Match
		if err := rows.Scan(&m.ID, &m.HomeTeamID, &m.AwayTeamID, &m.HomeGoals, &m.AwayGoals, &m.Week, &m.Played); err != nil {
			return league, err
		}
		league.Matches = append(league.Matches, m)
	}
	return league, nil
}

func (r SQLiteLeagueRepository) UpdateLeague(league League) error {
	// Not implemented: would update all teams and matches
	return nil
}

// CalculateTable updates the league table based on played matches
func (l *League) CalculateTable() []LeagueTableEntry {
    // Reset stats
    stats := make(map[int]*LeagueTableEntry)
    for _, t := range l.Teams {
        stats[t.ID] = &LeagueTableEntry{
            TeamID:   t.ID,
            TeamName: t.Name,
        }
    }
    for _, m := range l.Matches {
        if !m.Played {
            continue
        }
        home := stats[m.HomeTeamID]
        away := stats[m.AwayTeamID]
        homeGoals := 0
        awayGoals := 0
        if m.HomeGoals.Valid {
            homeGoals = int(m.HomeGoals.Int64)
        }
        if m.AwayGoals.Valid {
            awayGoals = int(m.AwayGoals.Int64)
        }
        home.GoalsFor += homeGoals
        home.GoalsAgainst += awayGoals
        home.MatchesPlayed++
        away.GoalsFor += awayGoals
        away.GoalsAgainst += homeGoals
        away.MatchesPlayed++
        if homeGoals > awayGoals {
            home.Points += 3
        } else if homeGoals < awayGoals {
            away.Points += 3
        } else {
            home.Points++
            away.Points++
        }
    }
    // Calculate goal difference
    table := make([]LeagueTableEntry, 0, len(stats))
    for _, entry := range stats {
        entry.GoalDifference = entry.GoalsFor - entry.GoalsAgainst
        table = append(table, *entry)
    }
    // Sort by points, then goal difference, then goals for
    sort.Slice(table, func(i, j int) bool {
        if table[i].Points != table[j].Points {
            return table[i].Points > table[j].Points
        }
        if table[i].GoalDifference != table[j].GoalDifference {
            return table[i].GoalDifference > table[j].GoalDifference
        }
        return table[i].GoalsFor > table[j].GoalsFor
    })
    return table
} 