package math

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PhonicsEngine handles math fact family classification and pattern detection
// Equivalent to pattern detection in reading app, but for math fact families
type PhonicsEngine struct {
	repo *Repository
}

// NewPhonicsEngine creates a new phonics (fact families) engine
func NewPhonicsEngine(repo *Repository) *PhonicsEngine {
	return &PhonicsEngine{repo: repo}
}

// FactFamilyPattern represents a mathematical fact family with regex pattern
type FactFamilyPattern struct {
	Family *FactFamily
	Pattern string // Regex pattern
	Difficulty int  // 1-15
	Regex    *regexp.Regexp
}

// MathFactFamilies defines 20+ fact family patterns used in math practice
var MathFactFamilies = []FactFamilyPattern{
	{
		Family: &FactFamily{
			Name:     "doubles",
			Category: "addition",
			Examples: []string{"2+2", "5+5", "10+10"},
			Hint:     "Adding a number to itself",
			Strategy: "Think of the number and add it to itself",
		},
		Pattern:    `^(\d+)\s*\+\s*\1$`,
		Difficulty: 2,
	},
	{
		Family: &FactFamily{
			Name:     "make_ten",
			Category: "addition",
			Examples: []string{"3+7", "4+6", "5+5"},
			Hint:     "Adding numbers that sum to 10",
			Strategy: "Find the missing number to make 10",
		},
		Pattern:    `^([1-9])\s*\+\s*([1-9])$`,
		Difficulty: 3,
	},
	{
		Family: &FactFamily{
			Name:     "plus_one",
			Category: "addition",
			Examples: []string{"5+1", "9+1", "15+1"},
			Hint:     "Adding 1 to a number",
			Strategy: "Count up by one",
		},
		Pattern:    `^(\d+)\s*\+\s*1$`,
		Difficulty: 1,
	},
	{
		Family: &FactFamily{
			Name:     "plus_two",
			Category: "addition",
			Examples: []string{"5+2", "8+2", "12+2"},
			Hint:     "Adding 2 to a number",
			Strategy: "Count up by two",
		},
		Pattern:    `^(\d+)\s*\+\s*2$`,
		Difficulty: 1,
	},
	{
		Family: &FactFamily{
			Name:     "plus_nine",
			Category: "addition",
			Examples: []string{"5+9", "8+9", "12+9"},
			Hint:     "Adding 9 to a number",
			Strategy: "Add 10 then subtract 1",
		},
		Pattern:    `^(\d+)\s*\+\s*9$`,
		Difficulty: 4,
	},
	{
		Family: &FactFamily{
			Name:     "plus_ten",
			Category: "addition",
			Examples: []string{"5+10", "15+10", "23+10"},
			Hint:     "Adding 10 to a number",
			Strategy: "Just increase the tens place by 1",
		},
		Pattern:    `^(\d+)\s*\+\s*10$`,
		Difficulty: 2,
	},
	{
		Family: &FactFamily{
			Name:     "teen_addition",
			Category: "addition",
			Examples: []string{"12+5", "15+7", "18+3"},
			Hint:     "Adding with teen numbers",
			Strategy: "Break apart the teen number",
		},
		Pattern:    `^(1[0-9])\s*\+\s*([0-9]+)$`,
		Difficulty: 5,
	},
	{
		Family: &FactFamily{
			Name:     "two_digit_addition",
			Category: "addition",
			Examples: []string{"25+30", "45+12", "67+23"},
			Hint:     "Adding two 2-digit numbers",
			Strategy: "Add tens first, then ones",
		},
		Pattern:    `^([1-9][0-9])\s*\+\s*([1-9][0-9])$`,
		Difficulty: 8,
	},
	{
		Family: &FactFamily{
			Name:     "minus_one",
			Category: "subtraction",
			Examples: []string{"5-1", "10-1", "15-1"},
			Hint:     "Subtracting 1",
			Strategy: "Count back by one",
		},
		Pattern:    `^(\d+)\s*-\s*1$`,
		Difficulty: 1,
	},
	{
		Family: &FactFamily{
			Name:     "minus_two",
			Category: "subtraction",
			Examples: []string{"5-2", "8-2", "12-2"},
			Hint:     "Subtracting 2",
			Strategy: "Count back by two",
		},
		Pattern:    `^(\d+)\s*-\s*2$`,
		Difficulty: 2,
	},
	{
		Family: &FactFamily{
			Name:     "minus_ten",
			Category: "subtraction",
			Examples: []string{"15-10", "20-10", "35-10"},
			Hint:     "Subtracting 10",
			Strategy: "Decrease the tens place by 1",
		},
		Pattern:    `^(1[0-9]|[2-9][0-9])\s*-\s*10$`,
		Difficulty: 3,
	},
	{
		Family: &FactFamily{
			Name:     "count_back",
			Category: "subtraction",
			Examples: []string{"20-5", "50-8", "100-7"},
			Hint:     "Counting back",
			Strategy: "Count back on your fingers or number line",
		},
		Pattern:    `^(\d{2,})\s*-\s*([1-9])$`,
		Difficulty: 6,
	},
	{
		Family: &FactFamily{
			Name:     "related_facts",
			Category: "mixed",
			Examples: []string{"5+3", "8-3", "6+4", "10-4"},
			Hint:     "Related addition and subtraction",
			Strategy: "If 5+3=8, then 8-3=5",
		},
		Pattern:    `^(\d+)\s*[+-]\s*(\d+)$`,
		Difficulty: 4,
	},
	{
		Family: &FactFamily{
			Name:     "two_factor",
			Category: "multiplication",
			Examples: []string{"5*2", "10*2", "15*2"},
			Hint:     "Multiplication by 2",
			Strategy: "Double the number",
		},
		Pattern:    `^(\d+)\s*\*\s*2$`,
		Difficulty: 3,
	},
	{
		Family: &FactFamily{
			Name:     "five_factor",
			Category: "multiplication",
			Examples: []string{"2*5", "7*5", "10*5"},
			Hint:     "Multiplication by 5",
			Strategy: "Skip count by 5s",
		},
		Pattern:    `^(\d+)\s*\*\s*5$`,
		Difficulty: 4,
	},
	{
		Family: &FactFamily{
			Name:     "ten_factor",
			Category: "multiplication",
			Examples: []string{"3*10", "7*10", "12*10"},
			Hint:     "Multiplication by 10",
			Strategy: "Add a zero to the number",
		},
		Pattern:    `^(\d+)\s*\*\s*10$`,
		Difficulty: 2,
	},
	{
		Family: &FactFamily{
			Name:     "square_numbers",
			Category: "multiplication",
			Examples: []string{"6*6", "7*7", "8*8", "9*9"},
			Hint:     "Squaring numbers",
			Strategy: "Memorize: 6*6=36, 7*7=49, 8*8=64, 9*9=81",
		},
		Pattern:    `^([6-9])\s*\*\s*\1$`,
		Difficulty: 5,
	},
	{
		Family: &FactFamily{
			Name:     "single_digit_multiply",
			Category: "multiplication",
			Examples: []string{"3*4", "7*6", "8*9"},
			Hint:     "Single digit multiplication",
			Strategy: "Use multiplication table facts",
		},
		Pattern:    `^([1-9])\s*\*\s*([1-9])$`,
		Difficulty: 6,
	},
	{
		Family: &FactFamily{
			Name:     "divide_by_two",
			Category: "division",
			Examples: []string{"10/2", "14/2", "20/2"},
			Hint:     "Division by 2",
			Strategy: "Halve the number",
		},
		Pattern:    `^(\d+)\s*\/\s*2$`,
		Difficulty: 4,
	},
	{
		Family: &FactFamily{
			Name:     "divide_by_five",
			Category: "division",
			Examples: []string{"10/5", "15/5", "25/5"},
			Hint:     "Division by 5",
			Strategy: "Skip count backwards by 5s",
		},
		Pattern:    `^(\d+)\s*\/\s*5$`,
		Difficulty: 5,
	},
	{
		Family: &FactFamily{
			Name:     "divide_by_ten",
			Category: "division",
			Examples: []string{"10/10", "20/10", "100/10"},
			Hint:     "Division by 10",
			Strategy: "Remove a zero from the number",
		},
		Pattern:    `^(\d+)\s*\/\s*10$`,
		Difficulty: 3,
	},
	{
		Family: &FactFamily{
			Name:     "single_digit_divide",
			Category: "division",
			Examples: []string{"12/3", "20/4", "56/7"},
			Hint:     "Single digit division",
			Strategy: "Use multiplication facts: if 3*4=12, then 12/3=4",
		},
		Pattern:    `^(\d+)\s*\/\s*([1-9])$`,
		Difficulty: 7,
	},
}

// InitializePatterns initializes regex patterns for all families
func InitializePatterns() {
	for i := range MathFactFamilies {
		MathFactFamilies[i].Regex = regexp.MustCompile(MathFactFamilies[i].Pattern)
	}
}

// DetectFactFamily identifies which fact family a problem belongs to
func (p *PhonicsEngine) DetectFactFamily(question string) string {
	// Normalize the question
	normalized := normalizeQuestion(question)

	// Try each pattern
	for _, familyPattern := range MathFactFamilies {
		if familyPattern.Regex != nil && familyPattern.Regex.MatchString(normalized) {
			return familyPattern.Family.Name
		}
	}

	// Check basic operations
	if strings.Contains(normalized, "+") {
		parts := strings.Split(normalized, "+")
		if len(parts) == 2 {
			a, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
			b, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
			if a+b == 10 {
				return "make_ten"
			}
			return "addition_basic"
		}
	}
	if strings.Contains(normalized, "-") {
		return "subtraction_basic"
	}
	if strings.Contains(normalized, "*") {
		return "multiplication_basic"
	}
	if strings.Contains(normalized, "/") {
		return "division_basic"
	}

	return "mixed_basic"
}

// GetFactFamilyInfo returns detailed information about a fact family
func (p *PhonicsEngine) GetFactFamilyInfo(familyName string) *FactFamily {
	for _, familyPattern := range MathFactFamilies {
		if familyPattern.Family.Name == familyName {
			return familyPattern.Family
		}
	}
	return nil
}

// AnalyzeUserPatternMastery analyzes which fact families a user struggles with
func (p *PhonicsEngine) AnalyzeUserPatternMastery(ctx context.Context, userID uint) (*PatternMasteryAnalysis, error) {
	analysis := &PatternMasteryAnalysis{
		UserID:       userID,
		Families:     make(map[string]*FamilyMastery),
		StrongestFam: "",
		WeakestFam:   "",
	}

	// Get all mistakes grouped by fact family
	mistakes, err := p.repo.GetMistakesByUser(ctx, userID, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to get mistakes: %w", err)
	}

	// Track family stats
	familyStats := make(map[string]*FamilyMastery)

	for _, mistake := range mistakes {
		famName := mistake.FactFamily
		if _, exists := familyStats[famName]; !exists {
			familyStats[famName] = &FamilyMastery{
				FamilyName:  famName,
				ErrorCount:  0,
				Practice:    0,
				Mastery:     0,
			}
		}
		familyStats[famName].ErrorCount += mistake.ErrorCount
	}

	// Get all history to count total attempts per family
	histories, err := p.repo.GetHistoryByUser(ctx, userID, 1000, 0)
	if err == nil {
		for _, history := range histories {
			famName := history.FactFamily
			if _, exists := familyStats[famName]; !exists {
				familyStats[famName] = &FamilyMastery{
					FamilyName:  famName,
					ErrorCount:  0,
					Practice:    0,
					Mastery:     0,
				}
			}
			familyStats[famName].Practice++
		}
	}

	// Calculate mastery for each family
	maxMastery := 0
	minMastery := 100
	for famName, stats := range familyStats {
		if stats.Practice > 0 {
			accuracy := float64(stats.Practice-stats.ErrorCount) / float64(stats.Practice)
			stats.Mastery = int(accuracy * 100)
		}

		analysis.Families[famName] = stats

		if stats.Mastery > maxMastery {
			maxMastery = stats.Mastery
			analysis.StrongestFam = famName
		}
		if stats.Mastery < minMastery {
			minMastery = stats.Mastery
			analysis.WeakestFam = famName
		}
	}

	analysis.AverageMastery = (maxMastery + minMastery) / 2

	return analysis, nil
}

// GetRemediationPlan generates practice recommendations for weak families
func (p *PhonicsEngine) GetRemediationPlan(ctx context.Context, userID uint, limit int) ([]*RemediationItem, error) {
	analysis, err := p.AnalyzeUserPatternMastery(ctx, userID)
	if err != nil {
		return nil, err
	}

	var items []*RemediationItem
	count := 0

	// Sort families by mastery (ascending) to get weakest first
	// For simplicity, iterate and collect low-mastery families
	for familyName, mastery := range analysis.Families {
		if count >= limit {
			break
		}

		if mastery.Mastery < 80 { // Focus on families below 80% mastery
			// Find difficulty level from MathFactFamilies
			difficulty := 5 // Default
			for _, fp := range MathFactFamilies {
				if fp.Family.Name == familyName {
					difficulty = fp.Difficulty
					break
				}
			}

			item := &RemediationItem{
				FamilyName:    familyName,
				CurrentMastery: mastery.Mastery,
				TargetMastery:  90,
				Priority:       calculatePriority(mastery.Mastery, mastery.ErrorCount),
				Difficulty:     difficulty,
				Recommendation: fmt.Sprintf("Practice %s facts until you reach 90%% accuracy", familyName),
			}
			items = append(items, item)
			count++
		}
	}

	return items, nil
}

// PatternMasteryAnalysis represents analysis of user's fact family mastery
type PatternMasteryAnalysis struct {
	UserID          uint
	Families        map[string]*FamilyMastery
	AverageMastery  int
	StrongestFam    string
	WeakestFam      string
	ReadyForAdvance bool
}

// FamilyMastery represents mastery level for one fact family
type FamilyMastery struct {
	FamilyName  string
	ErrorCount  int
	Practice    int
	Mastery     int // 0-100
}

// RemediationItem represents a remediation recommendation
type RemediationItem struct {
	FamilyName       string
	CurrentMastery   int
	TargetMastery    int
	Priority         string // high, medium, low
	Difficulty       int    // 1-15
	Recommendation   string
}

// Helper function to normalize questions for pattern matching
func normalizeQuestion(question string) string {
	// Remove spaces
	normalized := strings.ReplaceAll(question, " ", "")
	// Convert to lowercase
	normalized = strings.ToLower(normalized)
	// Remove leading/trailing whitespace
	normalized = strings.TrimSpace(normalized)
	return normalized
}

// Helper function to calculate priority
func calculatePriority(mastery int, errorCount int) string {
	if mastery < 30 || errorCount > 5 {
		return "high"
	}
	if mastery < 60 || errorCount > 2 {
		return "medium"
	}
	return "low"
}

// GetFactFamiliesByDifficulty returns all fact families at a specific difficulty level
func (p *PhonicsEngine) GetFactFamiliesByDifficulty(difficulty int) []*FactFamily {
	var families []*FactFamily
	for _, familyPattern := range MathFactFamilies {
		if familyPattern.Difficulty == difficulty {
			families = append(families, familyPattern.Family)
		}
	}
	return families
}

// GetFactFamiliesByCategory returns all fact families in a category
func (p *PhonicsEngine) GetFactFamiliesByCategory(category string) []*FactFamily {
	var families []*FactFamily
	for _, familyPattern := range MathFactFamilies {
		if familyPattern.Family.Category == category {
			families = append(families, familyPattern.Family)
		}
	}
	return families
}

// GenerateFactFamilyPractice generates a practice problem from a fact family
func (p *PhonicsEngine) GenerateFactFamilyPractice(familyName string) string {
	family := p.GetFactFamilyInfo(familyName)
	if family != nil && len(family.Examples) > 0 {
		// Return first example (would randomize in real implementation)
		return family.Examples[0]
	}
	return ""
}

// GetHistoryByUser returns all history records for a user
func (p *PhonicsEngine) GetHistoryByUser(ctx context.Context, userID uint) ([]*QuestionHistory, error) {
	return p.repo.GetHistoryByUser(ctx, userID, 1000, 0)
}

// GetMistakesByUser returns all mistake records for a user
func (p *PhonicsEngine) GetMistakesByUser(ctx context.Context, userID uint) ([]*Mistake, error) {
	return p.repo.GetMistakesByUser(ctx, userID, 1000)
}
