package math

import (
	"encoding/json"
	"net/http"
)

// IndexHandler serves the math app homepage
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Math App - Unified Educational Platform</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .placeholder { background: #f0f0f0; padding: 20px; border-radius: 5px; margin-top: 20px; }
    </style>
</head>
<body>
    <h1>Math App (Go)</h1>
    <div class="placeholder">
        <p><strong>Status:</strong> Placeholder - Ready for Phase 2 migration</p>
        <p>This is the math practice application. Features will be migrated from Python/Flask in Phase 2.</p>
        <a href="/dashboard">← Back to Dashboard</a>
    </div>
</body>
</html>
	`))
}

// ListProblems returns available math problems
func ListProblems(w http.ResponseWriter, r *http.Request) {
	problems := []map[string]interface{}{
		{"id": 1, "type": "addition", "difficulty": "easy"},
		{"id": 2, "type": "subtraction", "difficulty": "easy"},
		{"id": 3, "type": "multiplication", "difficulty": "medium"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"problems": problems,
		"status":   "placeholder",
	})
}

// SaveProgress saves math progress
func SaveProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Progress saved (placeholder)",
	})
}
