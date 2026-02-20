package typing

import (
	"encoding/json"
	"net/http"
)

// IndexHandler serves the typing app homepage
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Typing App - Unified Educational Platform</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .placeholder { background: #f0f0f0; padding: 20px; border-radius: 5px; margin-top: 20px; }
    </style>
</head>
<body>
    <h1>Typing App (Go)</h1>
    <div class="placeholder">
        <p><strong>Status:</strong> Placeholder - Ready for Phase 2 migration</p>
        <p>This is the typing practice application. Features will be migrated from Python/Flask in Phase 2.</p>
        <a href="/dashboard">← Back to Dashboard</a>
    </div>
</body>
</html>
	`))
}

// ListLessons returns available typing lessons
func ListLessons(w http.ResponseWriter, r *http.Request) {
	lessons := []map[string]interface{}{
		{"id": 1, "title": "Home Row Keys", "difficulty": "beginner"},
		{"id": 2, "title": "Top Row Keys", "difficulty": "intermediate"},
		{"id": 3, "title": "Bottom Row Keys", "difficulty": "intermediate"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"lessons": lessons,
		"status":  "placeholder",
	})
}

// SaveProgress saves typing progress
func SaveProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Progress saved (placeholder)",
	})
}
