package dashboard

import (
	"encoding/json"
	"net/http"
)

// IndexHandler serves the main dashboard
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Unified Educational Platform - Dashboard</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1200px;
            margin: 50px auto;
            padding: 20px;
            background: #f5f5f5;
        }
        h1 { color: #333; text-align: center; }
        .apps-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }
        .app-card {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            text-align: center;
            transition: transform 0.2s;
        }
        .app-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
        }
        .app-card h2 { color: #2c3e50; margin-bottom: 10px; }
        .app-card p { color: #7f8c8d; margin-bottom: 20px; }
        .app-card a {
            display: inline-block;
            padding: 10px 20px;
            background: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            transition: background 0.2s;
        }
        .app-card a:hover { background: #2980b9; }
        .status {
            background: #e8f5e9;
            color: #2e7d32;
            padding: 15px;
            border-radius: 5px;
            text-align: center;
            margin-bottom: 30px;
        }
    </style>
</head>
<body>
    <h1>Unified Educational Platform</h1>
    <div class="status">
        <strong>Phase 1 Complete:</strong> Go foundation layer is running successfully!
    </div>
    <div class="apps-grid">
        <div class="app-card">
            <h2>📝 Typing</h2>
            <p>Practice typing skills with interactive lessons</p>
            <a href="/typing">Launch App</a>
        </div>
        <div class="app-card">
            <h2>🔢 Math</h2>
            <p>Solve math problems and improve skills</p>
            <a href="/math">Launch App</a>
        </div>
        <div class="app-card">
            <h2>📚 Reading</h2>
            <p>Read books and test comprehension</p>
            <a href="/reading">Launch App</a>
        </div>
        <div class="app-card">
            <h2>🎹 Piano</h2>
            <p>Learn to play piano with guided lessons</p>
            <a href="/piano">Launch App</a>
        </div>
    </div>
    <div style="text-align: center; margin-top: 40px; color: #7f8c8d;">
        <p><a href="/health" style="color: #3498db;">System Health Check</a></p>
    </div>
</body>
</html>
	`))
}

// GetStats returns dashboard statistics
func GetStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"total_users":     0,
		"active_sessions": 0,
		"apps": map[string]interface{}{
			"typing":  map[string]int{"users": 0, "lessons_completed": 0},
			"math":    map[string]int{"users": 0, "problems_solved": 0},
			"reading": map[string]int{"users": 0, "books_read": 0},
			"piano":   map[string]int{"users": 0, "songs_learned": 0},
		},
		"status": "placeholder",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
