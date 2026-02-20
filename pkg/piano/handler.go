package piano

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var (
	baseTemplate *template.Template
	songsTemplate *template.Template
	practiceTemplate *template.Template
)

// init loads templates at startup
func init() {
	var err error
	templateDir := "pkg/piano/templates"

	// Load base template with all nested templates
	baseTemplate, err = template.ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		log.Printf("Warning: Could not parse templates from %s: %v", templateDir, err)
	}
}

// IndexHandler serves the piano app homepage with song listing
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Redirect to /piano/songs for the main listing
	http.Redirect(w, r, "/piano/songs", http.StatusSeeOther)
}

// SongsHandler displays available piano songs with filtering
func (r *Router) SongsHandler(w http.ResponseWriter, req *http.Request) {
	difficulty := req.URL.Query().Get("difficulty")

	// Get songs from repository
	songs, err := r.service.repo.GetSongs(req.Context(), difficulty, 100, 0)
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Failed to load songs")
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"Title": "Piano Songs",
		"Songs": songs,
		"Difficulty": difficulty,
	}

	if baseTemplate != nil {
		if err := baseTemplate.ExecuteTemplate(w, "base.html", data); err != nil {
			log.Printf("Error rendering songs template: %v", err)
			renderError(w, http.StatusInternalServerError, "Template rendering error")
		}
	} else {
		renderPlaceholder(w, "Songs", "Retrieving your available piano songs...")
	}
}

// PracticeHandler displays the practice interface for a specific song
func (r *Router) PracticeHandler(w http.ResponseWriter, req *http.Request) {
	// Get song ID from URL
	songIDStr := req.PathValue("id")

	var songID int
	_, err := fmt.Sscanf(songIDStr, "%d", &songID)
	if err != nil || songID <= 0 {
		renderError(w, http.StatusBadRequest, "Invalid song ID")
		return
	}

	// Get song details
	song, err := r.service.repo.GetSongByID(req.Context(), uint(songID))
	if err != nil {
		renderError(w, http.StatusNotFound, "Song not found")
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"Title": song.Title + " - Practice",
		"Song": song,
	}

	if baseTemplate != nil {
		if err := baseTemplate.ExecuteTemplate(w, "base.html", data); err != nil {
			log.Printf("Error rendering practice template: %v", err)
			renderError(w, http.StatusInternalServerError, "Template rendering error")
		}
	} else {
		renderPlaceholder(w, song.Title, fmt.Sprintf("Practice session for %s by %s", song.Title, song.Composer))
	}
}

// DashboardHandler shows piano progress statistics
func (r *Router) DashboardHandler(w http.ResponseWriter, req *http.Request) {
	// This would show user's piano progress, lessons completed, etc.
	// For now, render placeholder
	renderPlaceholder(w, "Piano Dashboard", "Your piano learning statistics and progress tracking")
}

// LeaderboardHandler shows piano leaderboard rankings
func (r *Router) LeaderboardHandler(w http.ResponseWriter, req *http.Request) {
	// Get leaderboard data
	leaderboard, err := r.service.repo.GetLeaderboard(req.Context(), 100)
	if err != nil {
		renderError(w, http.StatusInternalServerError, "Failed to load leaderboard")
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"Title": "Piano Leaderboard",
		"Leaderboard": leaderboard,
	}

	if baseTemplate != nil {
		if err := baseTemplate.ExecuteTemplate(w, "base.html", data); err != nil {
			log.Printf("Error rendering leaderboard template: %v", err)
			renderError(w, http.StatusInternalServerError, "Template rendering error")
		}
	} else {
		renderPlaceholder(w, "Piano Leaderboard", "Top piano players on the platform")
	}
}

// renderError displays an error page
func renderError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Error - Piano App</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        .error { background: #f8d7da; color: #721c24; padding: 20px; border-radius: 5px; border: 1px solid #f5c6cb; }
        a { color: #667eea; text-decoration: none; }
    </style>
</head>
<body>
    <div class="error">
        <h2>Error</h2>
        <p>%s</p>
        <a href="/piano/songs">← Back to Songs</a>
    </div>
</body>
</html>
	`, message)
}

// renderPlaceholder displays a placeholder page
func renderPlaceholder(w http.ResponseWriter, title string, description string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>%s - Piano App</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); }
        .placeholder { background: white; padding: 40px; border-radius: 10px; box-shadow: 0 10px 30px rgba(0,0,0,0.2); }
        h1 { color: #667eea; }
        p { color: #666; line-height: 1.6; }
        a { color: #667eea; text-decoration: none; font-weight: 600; }
    </style>
</head>
<body>
    <div class="placeholder">
        <h1>🎹 %s</h1>
        <p>%s</p>
        <a href="/piano/songs">← Back to Songs</a>
    </div>
</body>
</html>
	`, title, title, description)
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error JSON response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]interface{}{
		"error": message,
	})
}
