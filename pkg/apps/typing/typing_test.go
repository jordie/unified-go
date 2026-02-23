package typing

import (
	"testing"
)

// ============================================================================
// CALCULATION TESTS
// ============================================================================

func TestCalculateWPM(t *testing.T) {
	tests := []struct {
		name             string
		totalCharacters  int
		timeTakenSeconds float64
		expected         int
	}{
		{
			name:             "basic calculation",
			totalCharacters:  250, // 50 words
			timeTakenSeconds: 60,
			expected:         50,
		},
		{
			name:             "half minute",
			totalCharacters:  125, // 25 words
			timeTakenSeconds: 30,
			expected:         50,
		},
		{
			name:             "zero time",
			totalCharacters:  100,
			timeTakenSeconds: 0,
			expected:         0,
		},
		{
			name:             "high speed",
			totalCharacters:  500, // 100 words
			timeTakenSeconds: 60,
			expected:         100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateWPM(tt.totalCharacters, tt.timeTakenSeconds)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCalculateAccuracy(t *testing.T) {
	tests := []struct {
		name              string
		correctCharacters int
		totalCharacters   int
		expected          float64
	}{
		{
			name:              "perfect accuracy",
			correctCharacters: 100,
			totalCharacters:   100,
			expected:          100.0,
		},
		{
			name:              "half accuracy",
			correctCharacters: 50,
			totalCharacters:   100,
			expected:          50.0,
		},
		{
			name:              "zero total",
			correctCharacters: 0,
			totalCharacters:   0,
			expected:          0,
		},
		{
			name:              "95 percent",
			correctCharacters: 95,
			totalCharacters:   100,
			expected:          95.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateAccuracy(tt.correctCharacters, tt.totalCharacters)
			if result != tt.expected {
				t.Errorf("expected %.1f, got %.1f", tt.expected, result)
			}
		})
	}
}

// ============================================================================
// TEXT GENERATION TESTS
// ============================================================================

func TestGenerateWordText(t *testing.T) {
	app := &TypingApp{
		commonWordsList: initCommonWords(),
	}

	text := app.generateWordText(5)
	words := len(split(text))
	if words != 5 {
		t.Errorf("expected 5 words, got %d", words)
	}
}

func TestGenerateRaceText(t *testing.T) {
	app := &TypingApp{
		commonWordsList: initCommonWords(),
	}

	text := app.generateRaceText(30)
	words := len(split(text))
	if words != 30 {
		t.Errorf("expected 30 words, got %d", words)
	}
}

func TestGetTextFromCategory(t *testing.T) {
	app := &TypingApp{
		commonWordsList: initCommonWords(),
		textSamples:     initTextSamples(),
	}

	tests := []string{"common_words", "programming", "quotes", "paragraphs", "numbers", "special_characters"}
	for _, category := range tests {
		t.Run(category, func(t *testing.T) {
			text := app.getRandomTextFromCategory(category)
			if text == "" {
				t.Errorf("expected non-empty text for category %s", category)
			}
		})
	}
}

// ============================================================================
// RACING TESTS
// ============================================================================

func TestGenerateRaceOpponents(t *testing.T) {
	app := &TypingApp{
		commonWordsList: initCommonWords(),
	}

	difficulties := []string{"easy", "medium", "hard"}
	for _, difficulty := range difficulties {
		t.Run(difficulty, func(t *testing.T) {
			text, opponents := app.GenerateRaceOpponents(difficulty, 30)

			if text == "" {
				t.Error("expected non-empty race text")
			}

			if len(opponents) != 3 {
				t.Errorf("expected 3 opponents, got %d", len(opponents))
			}

			for _, opponent := range opponents {
				if opponent.Name == "" {
					t.Error("expected non-empty opponent name")
				}
				if opponent.WPM == 0 {
					t.Error("expected non-zero WPM")
				}
				if opponent.Accuracy == 0 {
					t.Error("expected non-zero accuracy")
				}
				if opponent.Car == "" {
					t.Error("expected non-empty car emoji")
				}
			}
		})
	}
}

// ============================================================================
// XP CALCULATION TESTS
// ============================================================================

func TestCalculateRaceXP(t *testing.T) {
	tests := []struct {
		name        string
		wpm         int
		accuracy    float64
		placement   int
		difficulty  string
		minExpected int
	}{
		{
			name:        "first place easy",
			wpm:         40,
			accuracy:    100,
			placement:   1,
			difficulty: "easy",
			minExpected: 80, // base 10 + placement 50 + accuracy 25 + speed 10 = 95
		},
		{
			name:        "first place hard",
			wpm:         80,
			accuracy:    99,
			placement:   1,
			difficulty: "hard",
			minExpected: 100, // base 10 + placement 50 + accuracy 25 + speed 20 * 1.5 = 142
		},
		{
			name:        "last place",
			wpm:         10,
			accuracy:    50,
			placement:   4,
			difficulty: "medium",
			minExpected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xp := CalculateRaceXP(tt.wpm, tt.accuracy, tt.placement, tt.difficulty)
			if xp < tt.minExpected {
				t.Errorf("expected at least %d XP, got %d", tt.minExpected, xp)
			}
		})
	}
}

// ============================================================================
// CAR PROGRESSION TESTS
// ============================================================================

func TestGetCarForXP(t *testing.T) {
	tests := []struct {
		xp       int
		expected string
	}{
		{0, "🚗"},
		{50, "🚗"},
		{100, "🚙"},
		{250, "🚕"},
		{500, "🏎️"},
		{1000, "🚀"},
		{2000, "🚀"},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.xp)), func(t *testing.T) {
			car := GetCarForXP(tt.xp)
			if car != tt.expected {
				t.Errorf("for %d XP, expected %s, got %s", tt.xp, tt.expected, car)
			}
		})
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func split(s string) []string {
	var words []string
	var word string
	for _, ch := range s {
		if ch == ' ' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(ch)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}
