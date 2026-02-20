package piano

import (
	"encoding/json"
	"net/http"
)

// IndexHandler serves the piano app homepage
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Piano App - Unified Educational Platform</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .placeholder { background: #f0f0f0; padding: 20px; border-radius: 5px; margin-top: 20px; }
    </style>
</head>
<body>
    <h1>Piano App (Go)</h1>
    <div class="placeholder">
        <p><strong>Status:</strong> Placeholder - Ready for Phase 2 migration</p>
        <p>This is the piano learning application. Features will be migrated from Python/Flask in Phase 2.</p>
        <a href="/dashboard">← Back to Dashboard</a>
    </div>
</body>
</html>
	`))
}

// ListSongs returns available piano songs
func ListSongs(w http.ResponseWriter, r *http.Request) {
	songs := []map[string]interface{}{
		{"id": 1, "title": "Twinkle Twinkle Little Star", "difficulty": "beginner"},
		{"id": 2, "title": "Mary Had a Little Lamb", "difficulty": "beginner"},
		{"id": 3, "title": "Ode to Joy", "difficulty": "intermediate"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"songs":  songs,
		"status": "placeholder",
	})
}

// SaveProgress saves piano progress
func SaveProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Progress saved (placeholder)",
	})
}
