# Reading App Algorithms

Detailed documentation of core algorithms and formulas used in the reading app.

## 1. Words Per Minute (WPM) Calculation

### Formula
```
WPM = (character_count / 5) / reading_time_minutes
```

### Explanation

- **Character Count**: Total characters in the content (whitespace trimmed)
- **Divide by 5**: Standard assumption that 1 word = 5 characters
- **Reading Time**: Time spent reading in minutes

### Implementation (Go)

```go
func (s *Service) CalculateWPM(content string, timeSpentSeconds float64) float64 {
    if timeSpentSeconds <= 0 {
        return 0
    }
    minutes := timeSpentSeconds / 60.0
    charCount := float64(len(strings.TrimSpace(content)))
    wpm := (charCount / 5.0) / minutes
    return math.Round(wpm*100) / 100
}
```

### Examples

**Example 1**: 250 characters in 2 minutes
- WPM = (250 / 5) / 2 = 50 / 2 = **25 WPM**

**Example 2**: 600 characters in 3 minutes
- WPM = (600 / 5) / 3 = 120 / 3 = **40 WPM**

**Example 3**: 3000 characters in 10 minutes
- WPM = (3000 / 5) / 10 = 600 / 10 = **60 WPM**

### Reading Speed Levels

- **< 100 WPM**: Beginner (elementary school)
- **100-200 WPM**: Intermediate (typical adult)
- **200-300 WPM**: Advanced (skilled reader)
- **> 300 WPM**: Expert (speed reader)

## 2. Accuracy Calculation

### Formula
```
Accuracy = ((total_characters - errors) / total_characters) × 100
```

### Explanation

- **Total Characters**: Length of original content
- **Errors**: Number of character-level mistakes
- **Percentage**: Converts to 0-100 scale

### Implementation (Go)

```go
func (s *Service) CalculateAccuracy(totalChars, errorCount int) float64 {
    if totalChars <= 0 {
        return 0
    }
    accuracy := (1.0 - float64(errorCount)/float64(totalChars)) * 100.0
    if accuracy < 0 {
        accuracy = 0
    }
    return math.Round(accuracy*100) / 100
}
```

### Examples

**Example 1**: 100 characters, 5 errors
- Accuracy = ((100 - 5) / 100) × 100 = 95%

**Example 2**: 250 characters, 10 errors
- Accuracy = ((250 - 10) / 250) × 100 = 96%

**Example 3**: 500 characters, 50 errors
- Accuracy = ((500 - 50) / 500) × 100 = 90%

### Accuracy Quality Levels

- **95-100%**: Excellent (professional quality)
- **90-94%**: Good (acceptable)
- **85-89%**: Fair (needs improvement)
- **< 85%**: Poor (requires practice)

## 3. Comprehension Score

### Formula
```
Comprehension = (correct_answers / total_questions) × 100
```

### Explanation

- **Correct Answers**: Number of assessment questions answered correctly
- **Total Questions**: Total questions in the comprehension test
- **Percentage**: Converts to 0-100 scale

### Implementation (Go)

```go
func (s *Service) CalculateComprehensionScore(correct, total int) float64 {
    if total <= 0 {
        return 0
    }
    score := (float64(correct) / float64(total)) * 100.0
    return math.Round(score*100) / 100
}
```

### Examples

**Example 1**: 4 correct out of 5 questions
- Comprehension = (4 / 5) × 100 = 80%

**Example 2**: 18 correct out of 20 questions
- Comprehension = (18 / 20) × 100 = 90%

**Example 3**: 25 correct out of 25 questions
- Comprehension = (25 / 25) × 100 = 100%

### Comprehension Quality Levels

- **90-100%**: Excellent (complete understanding)
- **80-89%**: Good (strong understanding)
- **70-79%**: Fair (basic understanding)
- **< 70%**: Poor (limited understanding)

## 4. Reading Level Estimation

### Algorithm

```go
func (s *Service) EstimateUserLevel(wpm float64) string {
    switch {
    case wpm < 100:
        return "beginner"
    case wpm < 200:
        return "intermediate"
    case wpm < 300:
        return "advanced"
    default:
        return "expert"
    }
}
```

### Level Characteristics

**Beginner (< 100 WPM)**
- Focuses on basic comprehension
- Needs vocabulary support
- Benefits from shorter passages
- Prioritize accuracy over speed

**Intermediate (100-200 WPM)**
- Balanced speed and comprehension
- Handles diverse text types
- Learns vocabulary in context
- Ready for varied difficulty

**Advanced (200-300 WPM)**
- High speed with good comprehension
- Reads for efficiency
- Handles complex texts
- Ready for challenging materials

**Expert (> 300 WPM)**
- Expert-level speed reading
- Excellent comprehension
- Handles any text type
- Focus on specialized materials

## 5. Performance Trend Analysis

### Trend Calculation

```go
func (s *Service) CalculateTrend(previous, current float64) string {
    if previous == 0 {
        return "neutral"
    }
    
    change := ((current - previous) / previous) * 100
    
    if change > 2 {
        return "improving"
    } else if change < -2 {
        return "declining"
    }
    return "neutral"
}
```

### Trend Thresholds

- **Improving**: Change > 2%
- **Declining**: Change < -2%
- **Neutral**: Change within ±2%

### Example

**Session 1**: 45 WPM
**Session 2**: 48 WPM
**Change**: (48 - 45) / 45 × 100 = 6.67% **→ Improving** ✓

## 6. Book Recommendation Algorithm

### Scoring Components

Books are scored on:
1. **Level Match** (40%): Does difficulty match user level?
2. **Performance** (30%): How is user performing recently?
3. **Genre** (20%): Matches user reading history?
4. **Popularity** (10%): Community engagement

### Selection Logic

1. Top 3 highest-scoring books recommended
2. One book at current level
3. One above current level (challenge)
4. One below mastered level (reinforcement)

## 7. Floating Point Precision

### Rounding

All calculations rounded to 2 decimal places:
```go
func round(value float64) float64 {
    return math.Round(value*100) / 100
}
```

### Edge Cases

- Division by zero: Return 0
- Negative values: Clamp to 0
- Empty content: Return 0
- Invalid input: Return error

## Performance Characteristics

All algorithms use:
- **O(n)** complexity (content length)
- In-memory processing
- < 1ms per calculation
- No external dependencies

## Testing

Comprehensive test coverage:
- Unit tests for formulas
- Edge case validation
- Integration with real data
- Performance benchmarks

Run: `go test ./pkg/reading -v`
