package typing

import (
	"context"
	"fmt"
	"math"
	"math/rand"
)

// ProcessRaceResult processes a racing session and calculates XP
func (s *Service) ProcessRaceResult(ctx context.Context, userID uint, wpm, accuracy float64, raceTime float64, placement int) (*Race, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	if placement < 1 || placement > 4 {
		return nil, fmt.Errorf("placement must be between 1 and 4")
	}

	if raceTime <= 0 {
		return nil, fmt.Errorf("race_time must be positive")
	}

	// Calculate XP
	xpEarned := s.CalculateRaceXP(wpm, accuracy, placement)

	race := &Race{
		UserID:    userID,
		Mode:      "standard",
		Placement: placement,
		WPM:       wpm,
		Accuracy:  accuracy,
		RaceTime:  raceTime,
		XPEarned:  xpEarned,
	}

	// Validate the race
	if err := race.Validate(); err != nil {
		return nil, fmt.Errorf("invalid race: %w", err)
	}

	// Save to repository
	id, err := s.repo.SaveRace(ctx, race)
	if err != nil {
		return nil, fmt.Errorf("failed to save race: %w", err)
	}

	race.ID = id
	return race, nil
}

// CalculateRaceXP calculates XP earned in a race using complex formula with bonuses
// Formula: Base (10) + PlacementBonus (0-50) + AccuracyBonus (0-25) + SpeedBonus (0-20) × DifficultyMultiplier (1.0-1.5)
func (s *Service) CalculateRaceXP(wpm, accuracy float64, placement int) int {
	breakdown := s.calculateXPBreakdown(wpm, accuracy, placement)
	return breakdown.Total
}

// calculateXPBreakdown provides detailed XP calculation breakdown
func (s *Service) calculateXPBreakdown(wpm, accuracy float64, placement int) XPBreakdown {
	// Base XP
	base := 10

	// Placement bonus: 1st=50, 2nd=30, 3rd=15, 4th=0
	placementBonus := 0
	switch placement {
	case 1:
		placementBonus = 50
	case 2:
		placementBonus = 30
	case 3:
		placementBonus = 15
	case 4:
		placementBonus = 0
	}

	// Accuracy bonus: 0-25 points based on accuracy percentage
	// 100% accuracy = 25 points, scales linearly down from there
	accuracyBonus := int(math.Round(accuracy / 4.0))
	if accuracyBonus > 25 {
		accuracyBonus = 25
	}

	// Speed bonus: 0-20 points based on WPM
	// Base threshold: 40 WPM = 0 bonus, 100+ WPM = 20 bonus
	speedBonus := 0
	if wpm >= 40 {
		speedBonus = int(math.Round((wpm - 40) / 3.0))
		if speedBonus > 20 {
			speedBonus = 20
		}
	}

	// Difficulty multiplier (1.0 by default for standard mode)
	difficultyMultiplier := 1.0

	// Total XP = (base + bonuses) × multiplier, rounded to int
	total := int(math.Round(float64(base+placementBonus+accuracyBonus+speedBonus) * difficultyMultiplier))

	return XPBreakdown{
		Base:                  base,
		PlacementBonus:        placementBonus,
		AccuracyBonus:         accuracyBonus,
		SpeedBonus:            speedBonus,
		DifficultyMultiplier:  difficultyMultiplier,
		Total:                 total,
	}
}

// GenerateAIOpponent creates an AI opponent for a race
func (s *Service) GenerateAIOpponent(difficulty string) *AIOpponent {
	params, exists := AIOpponentParams[RaceDifficulty(difficulty)]
	if !exists {
		params = AIOpponentParams[RaceDifficultyMedium]
	}

	// Random WPM in range
	wpm := params.MinWPM + (rand.Float64() * (params.MaxWPM - params.MinWPM))
	wpm = math.Round(wpm*10) / 10

	// Random accuracy in range
	accuracy := params.MinAccuracy + (rand.Float64() * (params.MaxAccuracy - params.MinAccuracy))
	accuracy = math.Round(accuracy*10) / 10

	// Random name and car
	nameIdx := rand.Intn(len(AIOpponentNames))
	carIdx := rand.Intn(len(AIOpponentCars))

	opponent := &AIOpponent{
		ID:         uint(rand.Intn(1000) + 1),
		Name:       AIOpponentNames[nameIdx],
		Difficulty: difficulty,
		WPM:        wpm,
		Accuracy:   accuracy,
		Car:        AIOpponentCars[carIdx],
	}

	return opponent
}

// GetRacingStats retrieves racing statistics for a user
func (s *Service) GetRacingStats(ctx context.Context, userID uint) (*UserRacingStats, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	stats, err := s.repo.GetUserRacingStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get racing stats: %w", err)
	}

	return stats, nil
}

// GetRacingLeaderboard retrieves top racers
func (s *Service) GetRacingLeaderboard(ctx context.Context, metric string, limit int) ([]UserRacingStats, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	leaderboard, err := s.repo.GetRacingLeaderboard(ctx, metric, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get racing leaderboard: %w", err)
	}

	return leaderboard, nil
}

// GetUnlockedCars returns list of cars unlocked by user
func (s *Service) GetUnlockedCars(ctx context.Context, userID uint) ([]CarProgression, error) {
	stats, err := s.GetRacingStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get racing stats: %w", err)
	}

	var unlockedCars []CarProgression
	for _, carProg := range CarProgressions {
		if stats.TotalXP >= carProg.XPRequired {
			carProg.Unlocked = true
			unlockedCars = append(unlockedCars, carProg)
		}
	}

	return unlockedCars, nil
}

// GetNextCarUnlock returns the next car to be unlocked and XP needed
func (s *Service) GetNextCarUnlock(ctx context.Context, userID uint) (*CarProgression, int, error) {
	stats, err := s.GetRacingStats(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get racing stats: %w", err)
	}

	// Find next locked car
	for i, carProg := range CarProgressions {
		if stats.TotalXP < carProg.XPRequired {
			xpNeeded := carProg.XPRequired - stats.TotalXP
			return &CarProgressions[i], xpNeeded, nil
		}
	}

	// All cars unlocked
	return nil, 0, nil
}

// GetRaceHistory retrieves user's race history with pagination
func (s *Service) GetRaceHistory(ctx context.Context, userID uint, limit, offset int) ([]Race, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	races, err := s.repo.GetRacesByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get race history: %w", err)
	}

	return races, nil
}

// SimulateAIRace simulates a race against an AI opponent
// Returns placement of human player (1-4)
func (s *Service) SimulateAIRace(playerWPM, playerAccuracy float64, aiOpponent *AIOpponent) int {
	// Simple simulation: compare WPM and accuracy to determine winner
	// Factor in accuracy more heavily than speed
	playerScore := playerWPM * (playerAccuracy / 100.0)
	aiScore := aiOpponent.WPM * (aiOpponent.Accuracy / 100.0)

	// Add some randomness to make races less predictable
	randomFactor := 0.85 + (rand.Float64() * 0.30) // 0.85-1.15
	playerScore *= randomFactor

	if playerScore > aiScore {
		return 1 // Player wins
	}
	return 2 // AI wins
}

// GetSelectedText returns a random text sample for the given category
func (s *Service) GetSelectedText(category string) string {
	samples, exists := TextSamples[category]
	if !exists || len(samples) == 0 {
		// Default to common words if category not found
		samples = TextSamples["common_words"]
	}

	if len(samples) == 0 {
		return "the quick brown fox jumps over the lazy dog"
	}

	idx := rand.Intn(len(samples))
	return samples[idx]
}

// CalculateRaceLevel estimates user's racing skill level
func (s *Service) CalculateRaceLevel(ctx context.Context, userID uint) (string, error) {
	stats, err := s.GetRacingStats(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get racing stats: %w", err)
	}

	// Estimate based on win rate
	if stats.TotalRaces == 0 {
		return "novice", nil
	}

	winRate := float64(stats.Wins) / float64(stats.TotalRaces)
	level := "novice"
	switch {
	case winRate < 0.1:
		level = "novice"
	case winRate < 0.25:
		level = "beginner"
	case winRate < 0.40:
		level = "intermediate"
	case winRate < 0.60:
		level = "advanced"
	default:
		level = "expert"
	}
	return level, nil
}
