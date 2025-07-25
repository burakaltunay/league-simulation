# Football League Simulation (Go)

## Quick Start (Docker)

1. **Build the Docker image:**
   ```sh
   docker build -t league-sim .
   ```
2. **Run the container:**
   ```sh
   docker run -p 8080:8080 league-sim
   ```

### Get League Table
```sh
http://localhost:8080/league/table
```

### Play Next Week
```sh
http://localhost:8080/league/next-week
```

### Play All Matches
```sh
http://localhost:8080/league/play-all
```

### Edit a Match Result
```sh
http://localhost:8080/match/1 -H "Content-Type: application/json" -d '{"home_goals":2,"away_goals":2}'
```

### Results by Week
```sh
http://localhost:8080/league/results-by-week
```

### Champion Probability (after week 4)
```sh
http://localhost:8080/league/champion-estimation
```

### Reset League
```sh
http://localhost:8080/league/reset
```
