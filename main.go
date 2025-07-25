package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"Case_study/models"
	"Case_study/storage"
	"os"
	"database/sql"
	"sort"
)

// Helper struct for JSON output
type MatchJSON struct {
	ID         int    `json:"id"`
	HomeTeamID int    `json:"home_team_id"`
	AwayTeamID int    `json:"away_team_id"`
	HomeGoals  *int   `json:"home_goals"`
	AwayGoals  *int   `json:"away_goals"`
	Week       int    `json:"week"`
	Played     bool   `json:"played"`
}

func matchToJSON(m models.Match) MatchJSON {
	var hg, ag *int
	if m.HomeGoals.Valid {
		v := int(m.HomeGoals.Int64)
		hg = &v
	}
	if m.AwayGoals.Valid {
		v := int(m.AwayGoals.Int64)
		ag = &v
	}
	return MatchJSON{
		ID: m.ID,
		HomeTeamID: m.HomeTeamID,
		AwayTeamID: m.AwayTeamID,
		HomeGoals: hg,
		AwayGoals: ag,
		Week: m.Week,
		Played: m.Played,
	}
}

var (
	leagueRepo models.LeagueRepository = models.SQLiteLeagueRepository{}
	matchSim   models.MatchSimulator = models.BasicMatchSimulator{}
)

func initDBAndData() {
	dbFile := "league.db"
	schemaFile := "sql/schema.sql"
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		storage.InitDB(dbFile, schemaFile)
		// Insert initial teams and matches
		teamRepo := models.SQLiteTeamRepository{}
		teams := []models.Team{
			{ID: 1, Name: "Lions", Strength: 90},
			{ID: 2, Name: "Tigers", Strength: 80},
			{ID: 3, Name: "Bears", Strength: 70},
			{ID: 4, Name: "Wolves", Strength: 60},
		}
		for _, t := range teams {
			_ = teamRepo.CreateTeam(t)
		}
		matchRepo := models.SQLiteMatchRepository{}
		matches := []models.Match{
			{ID: 1, HomeTeamID: 1, AwayTeamID: 2, Week: 1},
			{ID: 2, HomeTeamID: 3, AwayTeamID: 4, Week: 1},
			{ID: 3, HomeTeamID: 1, AwayTeamID: 3, Week: 2},
			{ID: 4, HomeTeamID: 2, AwayTeamID: 4, Week: 2},
			{ID: 5, HomeTeamID: 1, AwayTeamID: 4, Week: 3},
			{ID: 6, HomeTeamID: 2, AwayTeamID: 3, Week: 3},
			{ID: 7, HomeTeamID: 2, AwayTeamID: 1, Week: 4},
			{ID: 8, HomeTeamID: 4, AwayTeamID: 3, Week: 4},
			{ID: 9, HomeTeamID: 3, AwayTeamID: 1, Week: 5},
			{ID: 10, HomeTeamID: 4, AwayTeamID: 2, Week: 5},
			{ID: 11, HomeTeamID: 4, AwayTeamID: 1, Week: 6},
			{ID: 12, HomeTeamID: 3, AwayTeamID: 2, Week: 6},
		}
		for _, m := range matches {
			_ = matchRepo.CreateMatch(m)
		}
	} else {
		storage.InitDB(dbFile, schemaFile)
	}
}

// Add a struct for match results with team names for league table
type TableMatchResult struct {
	Week      int    `json:"week"`
	HomeTeam  string `json:"home_team"`
	AwayTeam  string `json:"away_team"`
	HomeGoals *int   `json:"home_goals"`
	AwayGoals *int   `json:"away_goals"`
}

// Add a struct for league table entry with matches
type LeagueTableWithMatches struct {
	TeamID         int               `json:"TeamID"`
	TeamName       string            `json:"TeamName"`
	Points         int               `json:"Points"`
	GoalsFor       int               `json:"GoalsFor"`
	GoalsAgainst   int               `json:"GoalsAgainst"`
	GoalDifference int               `json:"GoalDifference"`
	MatchesPlayed  int               `json:"MatchesPlayed"`
	MatchResults   []string          `json:"MatchResults"`
	Matches        []TableMatchResult `json:"Matches"`
}

// Helper to get match results for a given week
func getMatchResultsForWeek(matches []models.Match, teamNames map[int]string, week int) []string {
	var results []string
	for _, m := range matches {
		if m.Played && m.Week == week && m.HomeGoals.Valid && m.AwayGoals.Valid {
			res := teamNames[m.HomeTeamID] + " " + strconv.Itoa(int(m.HomeGoals.Int64)) + " - " + strconv.Itoa(int(m.AwayGoals.Int64)) + " " + teamNames[m.AwayTeamID]
			results = append(results, res)
		}
	}
	return results
}

// Helper to get the latest week with played matches
func getLatestPlayedWeek(matches []models.Match) int {
	maxWeek := 0
	for _, m := range matches {
		if m.Played && m.Week > maxWeek {
			maxWeek = m.Week
		}
	}
	return maxWeek
}

// Helper to build standings with win/draw/loss
func buildStandings(table []models.LeagueTableEntry, matches []models.Match) []map[string]interface{} {
	teamStats := make(map[int]struct{W, D, L int})
	for _, entry := range table {
		teamStats[entry.TeamID] = struct{W, D, L int}{0, 0, 0}
	}
	for _, m := range matches {
		if !m.Played || !m.HomeGoals.Valid || !m.AwayGoals.Valid {
			continue
		}
		hg := int(m.HomeGoals.Int64)
		ag := int(m.AwayGoals.Int64)
		if hg > ag {
			st := teamStats[m.HomeTeamID]
			st.W++
			teamStats[m.HomeTeamID] = st
			st = teamStats[m.AwayTeamID]
			st.L++
			teamStats[m.AwayTeamID] = st
		} else if hg < ag {
			st := teamStats[m.AwayTeamID]
			st.W++
			teamStats[m.AwayTeamID] = st
			st = teamStats[m.HomeTeamID]
			st.L++
			teamStats[m.HomeTeamID] = st
		} else {
			st := teamStats[m.HomeTeamID]
			st.D++
			teamStats[m.HomeTeamID] = st
			st = teamStats[m.AwayTeamID]
			st.D++
			teamStats[m.AwayTeamID] = st
		}
	}
	var standings []map[string]interface{}
	for _, entry := range table {
		st := teamStats[entry.TeamID]
		standings = append(standings, map[string]interface{}{
			"TeamID": entry.TeamID,
			"TeamName": entry.TeamName,
			"Points": entry.Points,
			"GoalsFor": entry.GoalsFor,
			"GoalsAgainst": entry.GoalsAgainst,
			"GoalDifference": entry.GoalDifference,
			"MatchesPlayed": entry.MatchesPlayed,
			"Wins": st.W,
			"Draws": st.D,
			"Losses": st.L,
		})
	}
	// Sort standings: Points desc, GoalDifference desc, GoalsFor desc, TeamName asc
	sort.SliceStable(standings, func(i, j int) bool {
		si, sj := standings[i], standings[j]
		if si["Points"].(int) != sj["Points"].(int) {
			return si["Points"].(int) > sj["Points"].(int)
		}
		if si["GoalDifference"].(int) != sj["GoalDifference"].(int) {
			return si["GoalDifference"].(int) > sj["GoalDifference"].(int)
		}
		if si["GoalsFor"].(int) != sj["GoalsFor"].(int) {
			return si["GoalsFor"].(int) > sj["GoalsFor"].(int)
		}
		return si["TeamName"].(string) < sj["TeamName"].(string)
	})
	return standings
}

// Update getLeagueTable to use the helper
func getLeagueTable(w http.ResponseWriter, r *http.Request) {
	league, err := leagueRepo.GetLeague()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	table := league.CalculateTable()
	teamNames := make(map[int]string)
	for _, t := range league.Teams {
		teamNames[t.ID] = t.Name
	}
	// Compose standings with win/draw/loss
	type StandingsEntry struct {
		TeamID         int    `json:"TeamID"`
		TeamName       string `json:"TeamName"`
		Points         int    `json:"Points"`
		GoalsFor       int    `json:"GoalsFor"`
		GoalsAgainst   int    `json:"GoalsAgainst"`
		GoalDifference int    `json:"GoalDifference"`
		MatchesPlayed  int    `json:"MatchesPlayed"`
		Wins           int    `json:"Wins"`
		Draws          int    `json:"Draws"`
		Losses         int    `json:"Losses"`
	}
	// Calculate win/draw/loss for each team
	teamStats := make(map[int]struct{W, D, L int})
	for _, t := range league.Teams {
		teamStats[t.ID] = struct{W, D, L int}{0, 0, 0}
	}
	for _, m := range league.Matches {
		if !m.Played || !m.HomeGoals.Valid || !m.AwayGoals.Valid {
			continue
		}
		hg := int(m.HomeGoals.Int64)
		ag := int(m.AwayGoals.Int64)
		if hg > ag {
			st := teamStats[m.HomeTeamID]
			st.W++
			teamStats[m.HomeTeamID] = st
			st = teamStats[m.AwayTeamID]
			st.L++
			teamStats[m.AwayTeamID] = st
		} else if hg < ag {
			st := teamStats[m.AwayTeamID]
			st.W++
			teamStats[m.AwayTeamID] = st
			st = teamStats[m.HomeTeamID]
			st.L++
			teamStats[m.HomeTeamID] = st
		} else {
			st := teamStats[m.HomeTeamID]
			st.D++
			teamStats[m.HomeTeamID] = st
			st = teamStats[m.AwayTeamID]
			st.D++
			teamStats[m.AwayTeamID] = st
		}
	}
	var standings []StandingsEntry
	for _, entry := range table {
		st := teamStats[entry.TeamID]
		standings = append(standings, StandingsEntry{
			TeamID: entry.TeamID,
			TeamName: entry.TeamName,
			Points: entry.Points,
			GoalsFor: entry.GoalsFor,
			GoalsAgainst: entry.GoalsAgainst,
			GoalDifference: entry.GoalDifference,
			MatchesPlayed: entry.MatchesPlayed,
			Wins: st.W,
			Draws: st.D,
			Losses: st.L,
		})
	}
	// Get latest week and match results
	latestWeek := getLatestPlayedWeek(league.Matches)
	matchResults := getMatchResultsForWeek(league.Matches, teamNames, latestWeek)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"standings": standings,
		"match_results": matchResults,
	})
}

func playNextWeek(w http.ResponseWriter, r *http.Request) {
	league, err := leagueRepo.GetLeague()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	week := 1
	for _, m := range league.Matches {
		if m.Week > week && !m.Played {
			week = m.Week
		}
	}
	for i := range league.Matches {
		m := &league.Matches[i]
		if m.Week == week && !m.Played {
			home, away := getTeamByID(league.Teams, m.HomeTeamID), getTeamByID(league.Teams, m.AwayTeamID)
			hg, ag := matchSim.SimulateMatch(home, away)
			m.HomeGoals = sql.NullInt64{Int64: int64(hg), Valid: true}
			m.AwayGoals = sql.NullInt64{Int64: int64(ag), Valid: true}
			m.Played = true
			repo := models.SQLiteMatchRepository{}
			repo.UpdateMatch(*m)
		}
	}
	table := league.CalculateTable()
	teamNames := make(map[int]string)
	for _, t := range league.Teams {
		teamNames[t.ID] = t.Name
	}
	standings := buildStandings(table, league.Matches)
	matchResults := getMatchResultsForWeek(league.Matches, teamNames, week)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"standings": standings,
		"match_results": matchResults,
	})
}

func playAll(w http.ResponseWriter, r *http.Request) {
	league, err := leagueRepo.GetLeague()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for i := range league.Matches {
		m := &league.Matches[i]
		if !m.Played {
			home, away := getTeamByID(league.Teams, m.HomeTeamID), getTeamByID(league.Teams, m.AwayTeamID)
			hg, ag := matchSim.SimulateMatch(home, away)
			m.HomeGoals = sql.NullInt64{Int64: int64(hg), Valid: true}
			m.AwayGoals = sql.NullInt64{Int64: int64(ag), Valid: true}
			m.Played = true
			repo := models.SQLiteMatchRepository{}
			repo.UpdateMatch(*m)
		}
	}
	table := league.CalculateTable()
	teamNames := make(map[int]string)
	for _, t := range league.Teams {
		teamNames[t.ID] = t.Name
	}
	standings := buildStandings(table, league.Matches)
	latestWeek := getLatestPlayedWeek(league.Matches)
	matchResults := getMatchResultsForWeek(league.Matches, teamNames, latestWeek)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"standings": standings,
		"match_results": matchResults,
	})
}

func editMatchResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Path[len("/match/") : ]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid match ID", 400)
		return
	}
	var req struct {
		HomeGoals int `json:"home_goals"`
		AwayGoals int `json:"away_goals"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	league, err := leagueRepo.GetLeague()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	found := false
	for i := range league.Matches {
		if league.Matches[i].ID == id {
			league.Matches[i].HomeGoals = sql.NullInt64{Int64: int64(req.HomeGoals), Valid: true}
			league.Matches[i].AwayGoals = sql.NullInt64{Int64: int64(req.AwayGoals), Valid: true}
			league.Matches[i].Played = true
			repo := models.SQLiteMatchRepository{}
			repo.UpdateMatch(league.Matches[i])
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Match not found", 404)
		return
	}
	table := league.CalculateTable()
	json.NewEncoder(w).Encode(table)
}

func estimateFinalTable(w http.ResponseWriter, r *http.Request) {
	league, err := leagueRepo.GetLeague()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	copyLeague := league
	for i := range copyLeague.Matches {
		m := &copyLeague.Matches[i]
		if !m.Played && m.Week > 4 {
			home, away := getTeamByID(copyLeague.Teams, m.HomeTeamID), getTeamByID(copyLeague.Teams, m.AwayTeamID)
			hg, ag := matchSim.SimulateMatch(home, away)
			m.HomeGoals = sql.NullInt64{Int64: int64(hg), Valid: true}
			m.AwayGoals = sql.NullInt64{Int64: int64(ag), Valid: true}
			m.Played = true
		}
	}
	table := copyLeague.CalculateTable()
	teamNames := make(map[int]string)
	for _, t := range league.Teams {
		teamNames[t.ID] = t.Name
	}
	standings := buildStandings(table, league.Matches)
	latestWeek := getLatestPlayedWeek(league.Matches)
	matchResults := getMatchResultsForWeek(league.Matches, teamNames, latestWeek)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"standings": standings,
		"match_results": matchResults,
	})
}

func resultsByWeek(w http.ResponseWriter, r *http.Request) {
	league, err := leagueRepo.GetLeague()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	teamNames := make(map[int]string)
	for _, t := range league.Teams {
		teamNames[t.ID] = t.Name
	}
	results := make(map[int][]string)
	for _, m := range league.Matches {
		if m.Played && m.HomeGoals.Valid && m.AwayGoals.Valid {
			res := teamNames[m.HomeTeamID] + " " + strconv.Itoa(int(m.HomeGoals.Int64)) + " - " + strconv.Itoa(int(m.AwayGoals.Int64)) + " " + teamNames[m.AwayTeamID]
			results[m.Week] = append(results[m.Week], res)
		}
	}
	latestWeek := getLatestPlayedWeek(league.Matches)
	matchResults := getMatchResultsForWeek(league.Matches, teamNames, latestWeek)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results_by_week": results,
		"match_results": matchResults,
	})
}

func afterWeek4Estimate(w http.ResponseWriter, r *http.Request) {
	league, err := leagueRepo.GetLeague()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	played := []MatchJSON{}
	for _, m := range league.Matches {
		if m.Week <= 4 && m.Played {
			played = append(played, matchToJSON(m))
		}
	}
	copyLeague := league
	for i := range copyLeague.Matches {
		m := &copyLeague.Matches[i]
		if !m.Played && m.Week > 4 {
			home, away := getTeamByID(copyLeague.Teams, m.HomeTeamID), getTeamByID(copyLeague.Teams, m.AwayTeamID)
			hg, ag := matchSim.SimulateMatch(home, away)
			m.HomeGoals = sql.NullInt64{Int64: int64(hg), Valid: true}
			m.AwayGoals = sql.NullInt64{Int64: int64(ag), Valid: true}
			m.Played = true
		}
	}
	table := copyLeague.CalculateTable()
	teamNames := make(map[int]string)
	for _, t := range league.Teams {
		teamNames[t.ID] = t.Name
	}
	standings := buildStandings(table, league.Matches)
	latestWeek := getLatestPlayedWeek(league.Matches)
	matchResults := getMatchResultsForWeek(league.Matches, teamNames, latestWeek)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"played_matches_up_to_week4": played,
		"estimated_final_table": standings,
		"match_results": matchResults,
	})
}

// Deep copy helper for League
func deepCopyLeague(league models.League) models.League {
	newLeague := league
	newLeague.Teams = make([]models.Team, len(league.Teams))
	copy(newLeague.Teams, league.Teams)
	newLeague.Matches = make([]models.Match, len(league.Matches))
	for i, m := range league.Matches {
		newLeague.Matches[i] = m
	}
	return newLeague
}

func championEstimation(w http.ResponseWriter, r *http.Request) {
	const simulations = 1000
	league, err := leagueRepo.GetLeague()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	teamChampions := make(map[int]int)
	teamNames := make(map[int]string)
	for _, t := range league.Teams {
		teamNames[t.ID] = t.Name
	}
	for sim := 0; sim < simulations; sim++ {
		copyLeague := deepCopyLeague(league)
		r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(sim*1000)))
		for i := range copyLeague.Matches {
			m := &copyLeague.Matches[i]
			if !m.Played && m.Week > 4 {
				home, away := getTeamByID(copyLeague.Teams, m.HomeTeamID), getTeamByID(copyLeague.Teams, m.AwayTeamID)
				hg, ag := simulateWithRand(home, away, r)
				m.HomeGoals = sql.NullInt64{Int64: int64(hg), Valid: true}
				m.AwayGoals = sql.NullInt64{Int64: int64(ag), Valid: true}
				m.Played = true
			}
		}
		table := copyLeague.CalculateTable()
		if len(table) > 0 {
			teamChampions[table[0].TeamID]++
		}
	}
	result := make(map[string]float64)
	for id, count := range teamChampions {
		result[teamNames[id]] = float64(count) * 100.0 / float64(simulations)
	}
	json.NewEncoder(w).Encode(result)
}

// simulateWithRand is like SimulateMatch but uses a custom rand.Rand
func simulateWithRand(home models.Team, away models.Team, r *rand.Rand) (int, int) {
	homeStrength := float64(home.Strength) * 1.1
	awayStrength := float64(away.Strength)
	totalStrength := homeStrength + awayStrength
	homeGoals := int((homeStrength/totalStrength)*3 + 0.5)
	awayGoals := int((awayStrength/totalStrength)*3 + 0.5)
	if homeGoals > 0 && r.Intn(4) == 0 { homeGoals-- }
	if awayGoals > 0 && r.Intn(4) == 0 { awayGoals-- }
	if r.Intn(10) == 0 { homeGoals++ }
	if r.Intn(10) == 0 { awayGoals++ }
	if homeGoals < 0 { homeGoals = 0 }
	if awayGoals < 0 { awayGoals = 0 }
	return homeGoals, awayGoals
}

func resetLeague(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	db := storage.GetDB()
	_, err := db.Exec("UPDATE matches SET home_goals = NULL, away_goals = NULL, played = 0")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reset successful"})
}

func getTeamByID(teams []models.Team, id int) models.Team {
	for _, t := range teams {
		if t.ID == id {
			return t
		}
	}
	return models.Team{}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	initDBAndData()
	http.HandleFunc("/league/table", getLeagueTable)
	http.HandleFunc("/league/next-week", playNextWeek)
	http.HandleFunc("/league/play-all", playAll)
	http.HandleFunc("/league/estimate", estimateFinalTable)
	http.HandleFunc("/league/results-by-week", resultsByWeek)
	http.HandleFunc("/league/after-week4-estimate", afterWeek4Estimate)
	http.HandleFunc("/league/champion-estimation", championEstimation)
	http.HandleFunc("/league/reset", resetLeague)
	http.HandleFunc("/match/", editMatchResult)
	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
} 