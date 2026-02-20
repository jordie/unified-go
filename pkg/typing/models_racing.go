package typing

// TextSamples contains categorized text for typing tests
var TextSamples = map[string][]string{
	"common_words": {
		"the quick brown fox jumps over the lazy dog",
		"pack my box with five dozen liquor jugs",
		"how vexingly quick daft zebras jump",
		"the five boxing wizards jump quickly",
		"sphinx of black quartz judge my vow",
		"two driven jocks help fax my big quiz",
		"five quacking zephyrs jolt my wax bed",
		"the jay pig fox zebra and my wolves quack",
		"a wizard's job is to vex chumps quickly in fog",
		"watch jeopardy alex trebek's fun tv quiz game",
	},
	"programming": {
		"function calculateSum(a, b) { return a + b; }",
		"const array = [1, 2, 3, 4, 5].map(x => x * 2);",
		"if (condition) { doSomething(); } else { doSomethingElse(); }",
		"class MyClass extends BaseClass { constructor() { super(); } }",
		"try { await fetch(url); } catch (error) { console.log(error); }",
		"import React from 'react'; export default App;",
		"def fibonacci(n): return n if n <= 1 else fibonacci(n-1) + fibonacci(n-2)",
		"SELECT * FROM users WHERE age > 18 ORDER BY name ASC;",
		"git commit -m 'Initial commit' && git push origin main",
		"docker run -it --rm -p 8080:80 nginx:latest",
	},
	"quotes": {
		"The only way to do great work is to love what you do. - Steve Jobs",
		"Innovation distinguishes between a leader and a follower. - Steve Jobs",
		"Life is what happens when you're busy making other plans. - John Lennon",
		"The future belongs to those who believe in the beauty of their dreams. - Eleanor Roosevelt",
		"It is during our darkest moments that we must focus to see the light. - Aristotle",
		"The best way to predict the future is to create it. - Peter Drucker",
		"Success is not final, failure is not fatal: it is the courage to continue that counts. - Winston Churchill",
		"The only impossible thing is that which you don't attempt. - Unknown",
		"Your time is limited, don't waste it living someone else's life. - Steve Jobs",
		"The greatest glory in living lies not in never falling, but in rising every time we fall. - Nelson Mandela",
	},
	"paragraphs": {
		"In the heart of the bustling city, where skyscrapers touched the clouds and streets hummed with endless activity, there lived a small community of artists who found beauty in the chaos. They gathered each evening in a forgotten courtyard, sharing stories and creating masterpieces that captured the soul of urban life.",
		"Technology has revolutionized the way we communicate, work, and live. From smartphones that connect us instantly to anyone around the world, to artificial intelligence that helps us solve complex problems, we are living in an age of unprecedented innovation. Yet with these advances come new challenges that we must navigate carefully.",
		"The ocean stretched endlessly before her, its waves dancing in the golden light of sunset. She had always found peace by the water, where the rhythmic sound of the tide seemed to wash away the worries of the world. This evening was no different, as she sat on the weathered dock, contemplating the journey that had brought her here.",
		"Learning to code is like learning a new language. At first, the syntax seems foreign and the concepts abstract. But with practice and patience, patterns begin to emerge. Soon, you're not just writing code; you're crafting solutions, building applications, and bringing ideas to life through the power of programming.",
		"The art of cooking is more than just following recipes. It's about understanding flavors, techniques, and the science behind how ingredients interact. A great chef doesn't just cook food; they create experiences, tell stories, and bring people together through the universal language of cuisine.",
	},
	"numbers": {
		"123 456 789 012 345 678 901 234 567 890",
		"3.14159 2.71828 1.41421 1.73205 2.23606",
		"2024 2025 2026 2027 2028 2029 2030 2031",
		"100% 75% 50% 25% 0% -25% -50% -75% -100%",
		"$1,234.56 €987.65 £456.78 ¥123,456 ₹78,901",
		"192.168.1.1 255.255.255.0 127.0.0.1 8.8.8.8",
		"1st 2nd 3rd 4th 5th 6th 7th 8th 9th 10th",
		"1/2 1/3 1/4 2/3 3/4 1/5 2/5 3/5 4/5 1/8",
		"+1 (555) 123-4567 ext. 890 PIN: 1234",
		"10:30 AM 2:45 PM 18:00 23:59 00:00 12:00",
	},
	"special_characters": {
		"!@#$%^&*()_+-=[]{}|;':\",./<>?",
		"email@example.com user_name@domain.co.uk",
		"https://www.example.com/path?param=value&other=123",
		"C:\\Users\\Name\\Documents\\file.txt",
		"/home/user/documents/project/src/main.py",
		"~!@#$%^&*()_+{}|:\"<>? all special chars",
		"Brackets: [] {} () <> all types here",
		"Symbols: * & ^ % $ # @ ! ~ ` | \\ / ?",
		"Math: + - * / = % ^ < > <= >= != == ** //",
		"Punctuation: . , ; : ! ? - -- ... \" ' « »",
	},
}

// CommonWordsList represents 200 most common English words
var CommonWordsList = []string{
	"the", "be", "to", "of", "and", "a", "in", "that", "have", "i",
	"it", "for", "not", "on", "with", "he", "as", "you", "do", "at",
	"this", "but", "his", "by", "from", "they", "we", "say", "her", "she",
	"or", "an", "will", "my", "one", "all", "would", "there", "their", "what",
	"so", "up", "out", "if", "about", "who", "get", "which", "go", "me",
	"when", "make", "can", "like", "time", "no", "just", "him", "know", "take",
	"people", "into", "year", "your", "good", "some", "could", "them", "see", "other",
	"than", "then", "now", "look", "only", "come", "its", "over", "think", "also",
	"back", "after", "use", "two", "how", "our", "work", "first", "well", "way",
	"even", "new", "want", "because", "any", "these", "give", "day", "most", "us",
	"is", "was", "are", "been", "has", "had", "were", "said", "did", "get",
	"may", "part", "made", "find", "where", "much", "too", "very", "still", "being",
	"going", "why", "before", "never", "here", "more", "out", "do", "like", "just",
	"should", "over", "such", "great", "think", "say", "help", "low", "line", "differ",
	"turn", "cause", "much", "mean", "before", "move", "right", "boy", "old", "too",
	"same", "tell", "does", "set", "three", "want", "air", "well", "also", "play",
	"small", "end", "put", "home", "read", "hand", "port", "large", "spell", "add",
	"land", "here", "must", "big", "high", "such", "follow", "act", "why", "ask",
	"men", "change", "went", "light", "kind", "off", "need", "house", "picture", "try",
	"ask", "again", "no", "problem", "man", "day", "thing", "old", "see", "get",
}

// RaceDifficulties contains difficulty levels for racing
type RaceDifficulty string

const (
	RaceDifficultyEasy   RaceDifficulty = "easy"
	RaceDifficultyMedium RaceDifficulty = "medium"
	RaceDifficultyHard   RaceDifficulty = "hard"
)

// AIOpponentGenerationParams defines WPM and accuracy ranges for AI opponents
var AIOpponentParams = map[RaceDifficulty]struct {
	MinWPM           float64
	MaxWPM           float64
	MinAccuracy      float64
	MaxAccuracy      float64
	DifficultyMultiplier float64
}{
	RaceDifficultyEasy: {
		MinWPM:           30,
		MaxWPM:           60,
		MinAccuracy:      85,
		MaxAccuracy:      95,
		DifficultyMultiplier: 1.0,
	},
	RaceDifficultyMedium: {
		MinWPM:           60,
		MaxWPM:           100,
		MinAccuracy:      90,
		MaxAccuracy:      98,
		DifficultyMultiplier: 1.2,
	},
	RaceDifficultyHard: {
		MinWPM:           100,
		MaxWPM:           150,
		MinAccuracy:      95,
		MaxAccuracy:      99,
		DifficultyMultiplier: 1.5,
	},
}

// AIOpponentNames provides names for AI opponents
var AIOpponentNames = []string{
	"Speed Racer", "Flash Gordon", "Lightning Bolt", "Swift Arrow", "Quick Silver",
	"Dash Master", "Speedy Gonzales", "Zoom Zoom", "Turbo Tom", "Rocket Ron",
	"Rapid Ray", "Velocity Vince", "Fast Frank", "Sprint Steve", "Quick Quinn",
	"Swift Sally", "Dash Diana", "Speedy Susan", "Rapid Ruby", "Quick Quinn",
}

// AIOpponentCars provides car emojis for AI opponents
var AIOpponentCars = []string{
	"🚗", "🏎️", "🚕", "🚙", "🚓", "🚑", "🚒", "🚐", "🛻", "🚛",
}
