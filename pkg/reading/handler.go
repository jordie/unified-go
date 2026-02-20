package reading

import (
	"encoding/json"
	"net/http"
)

// IndexHandler serves the reading app homepage
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Reading App - Unified Educational Platform</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .placeholder { background: #f0f0f0; padding: 20px; border-radius: 5px; margin-top: 20px; }
    </style>
</head>
<body>
    <h1>Reading App (Go)</h1>
    <div class="placeholder">
        <p><strong>Status:</strong> Placeholder - Ready for Phase 2 migration</p>
        <p>This is the reading comprehension application. Features will be migrated from Python/Flask in Phase 2.</p>
        <a href="/dashboard">← Back to Dashboard</a>
    </div>
</body>
</html>
	`))
}

// ListBooks returns available reading books
func ListBooks(w http.ResponseWriter, r *http.Request) {
	books := []map[string]interface{}{
		{"id": 1, "title": "The Cat in the Hat", "level": "beginner"},
		{"id": 2, "title": "Charlotte's Web", "level": "intermediate"},
		{"id": 3, "title": "Harry Potter", "level": "advanced"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"books":  books,
		"status": "placeholder",
	})
}

// SaveProgress saves reading progress
func SaveProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Progress saved (placeholder)",
	})
}
