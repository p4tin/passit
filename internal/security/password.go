package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"strings"

	"golang.org/x/crypto/argon2"
)

var dictionaryWords = []string{
	"ache", "acorn", "adept", "aero", "age", "airy", "alert", "amber",
	"angle", "arch", "arid", "ash", "atom", "audio", "aura", "awe",
	"bake", "bark", "beam", "beep", "belt", "bend", "best", "bias",
	"bike", "bill", "bind", "bird", "bite", "blue", "bold", "bolt",
	"book", "brim", "brisk", "broom", "buck", "byte", "cable", "calm",
	"camp", "cap", "card", "cart", "case", "cash", "cell", "charm",
	"chess", "chime", "chip", "civic", "clay", "clip", "clue", "coal",
	"coin", "comb", "cord", "core", "crisp", "crow", "cube", "curl",
	"dash", "data", "dawn", "deck", "dent", "dial", "dice", "dime",
	"disk", "dove", "drip", "drum", "dusk", "echo", "edge", "ember",
	"envy", "epoch", "faint", "fern", "fife", "file", "film", "fine",
	"firm", "flame", "flint", "flip", "flux", "foam", "fold", "fork",
	"frost", "gain", "gale", "game", "gear", "gem", "gift", "gilt",
	"glow", "gnat", "golf", "grip", "gush", "hail", "halo", "hand",
	"harp", "haze", "heft", "hint", "hive", "hope", "icon", "idle",
	"inch", "ink", "jazz", "jolt", "june", "keen", "keep", "kite",
	"lace", "lamp", "leaf", "lean", "lens", "lift", "lilt", "link",
	"lisp", "lumen", "lurk", "mace", "made", "mail", "main", "mask",
	"melt", "memo", "mesh", "mild", "mint", "mode", "mole", "myth",
	"navy", "neon", "nest", "node", "note", "nova", "oak", "odds",
	"opal", "orbit", "pace", "palm", "park", "pearl", "perk", "piano",
	"pike", "pink", "pipe", "pixel", "plum", "pond", "pouch", "purr",
	"quill", "quiz", "raft", "rail", "rain", "ramp", "rank", "raven",
	"reef", "ring", "rivet", "road", "roam", "robe", "rock", "root",
	"ruby", "sage", "sail", "scan", "scar", "seal", "seed", "silk",
	"sing", "sink", "slim", "slug", "smok", "snap", "snow", "soda",
	"solo", "soup", "spin", "spur", "star", "stem", "step", "storm",
	"suit", "surf", "sway", "tact", "tape", "task", "tide", "tier",
	"tile", "tilt", "tint", "tone", "twin", "urn", "veil", "vent",
	"vest", "vibe", "view", "vine", "visa", "void", "vow", "wave",
	"weld", "whim", "wisp", "wolf", "yarn", "yoke", "zinc", "zone",
}

type StrengthFeedback struct {
	Valid                 bool
	Length                int
	EntropyBits           float64
	OnlineCrackTimeYears  float64
	OfflineCrackTimeYears float64
	Issues                []string
}

func GeneratePassword(length int) string {
	const targetLen = 12

	for i := 0; i < 200; i++ {
		passphrase := generatePronounceable(targetLen)
		passphrase = applySubstitutions(passphrase)
		passphrase = ensureDigit(passphrase)
		passphrase = ensureSymbol(passphrase)
		passphrase = ensureCapital(passphrase)
		if ValidatePasswordStrength(passphrase).Valid {
			return passphrase
		}
	}

	passphrase := generatePronounceable(targetLen)
	passphrase = applySubstitutions(passphrase)
	passphrase = ensureDigit(passphrase)
	passphrase = ensureSymbol(passphrase)
	passphrase = ensureCapital(passphrase)
	return passphrase
}

func applySubstitutions(input string) string {
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		switch r {
		case 'a', 'A':
			b.WriteRune('@')
		case 'i', 'I':
			b.WriteRune('1')
		case 'b', 'B':
			b.WriteRune('8')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func generatePronounceable(length int) string {
	const consonants = "bcdfghjkmnpqrstvwxyz"
	const vowels = "aeiou"

	if length <= 0 {
		return ""
	}

	var b strings.Builder
	b.Grow(length)
	for i := 0; i < length; i++ {
		if i%2 == 0 {
			b.WriteByte(randomByte(consonants))
		} else {
			b.WriteByte(randomByte(vowels))
		}
	}
	return b.String()
}

func randomByte(charset string) byte {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return charset[0]
	}
	return charset[index.Int64()]
}

func ValidatePasswordStrength(password string) StrengthFeedback {
	feedback := StrengthFeedback{
		Length:      len(password),
		EntropyBits: estimateEntropyBits(password),
	}

	feedback.OnlineCrackTimeYears = estimateCrackTimeYears(feedback.EntropyBits, 100)
	feedback.OfflineCrackTimeYears = estimateCrackTimeYears(feedback.EntropyBits, 1e10)

	if feedback.Length < 12 {
		feedback.Issues = append(feedback.Issues, "use at least 12 characters")
	}
	if feedback.EntropyBits < 60 {
		feedback.Issues = append(feedback.Issues, "increase entropy to at least 60 bits")
	}
	if isDictionaryBased(password) {
		feedback.Issues = append(feedback.Issues, "avoid single dictionary-word passwords")
	}
	if hasRepeatedPattern(password) {
		feedback.Issues = append(feedback.Issues, "avoid repeated patterns")
	}

	feedback.Valid = len(feedback.Issues) == 0
	return feedback
}

func estimateEntropyBits(password string) float64 {
	charsetSize := characterSetSize(password)
	if charsetSize == 0 || len(password) == 0 {
		return 0
	}
	return math.Log2(float64(charsetSize)) * float64(len(password))
}

func characterSetSize(password string) int {
	const lower = 26
	const upper = 26
	const digits = 10
	const symbols = 32

	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSymbol := false

	for _, r := range password {
		switch {
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= '0' && r <= '9':
			hasDigit = true
		default:
			hasSymbol = true
		}
	}

	size := 0
	if hasLower {
		size += lower
	}
	if hasUpper {
		size += upper
	}
	if hasDigit {
		size += digits
	}
	if hasSymbol {
		size += symbols
	}
	return size
}

func estimateCrackTimeYears(entropyBits float64, guessesPerSecond float64) float64 {
	if entropyBits <= 0 || guessesPerSecond <= 0 {
		return 0
	}
	guesses := math.Exp2(entropyBits) / 2
	seconds := guesses / guessesPerSecond
	return seconds / (60 * 60 * 24 * 365)
}

func isDictionaryBased(password string) bool {
	words := extractWords(password)
	if len(words) != 1 {
		return false
	}

	word := words[0]
	for _, dictWord := range dictionaryWords {
		if word == dictWord {
			return true
		}
	}
	return false
}

func extractWords(password string) []string {
	var words []string
	var b strings.Builder

	for _, r := range password {
		if r >= 'a' && r <= 'z' {
			b.WriteRune(r)
			continue
		}
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		if b.Len() > 0 {
			words = append(words, b.String())
			b.Reset()
		}
	}
	if b.Len() > 0 {
		words = append(words, b.String())
	}
	return words
}

func hasRepeatedPattern(password string) bool {
	length := len(password)
	if length < 2 {
		return false
	}

	for size := 1; size <= length/2; size++ {
		if length%size != 0 {
			continue
		}
		chunk := password[:size]
		if strings.Repeat(chunk, length/size) == password {
			return true
		}
	}
	return false
}

func ensureDigit(input string) string {
	for i := 0; i < len(input); i++ {
		if input[i] >= '0' && input[i] <= '9' {
			return input
		}
	}

	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(input))))
	if err != nil {
		index = big.NewInt(0)
	}
	pos := int(index.Int64())

	digitIndex, err := rand.Int(rand.Reader, big.NewInt(10))
	if err != nil {
		digitIndex = big.NewInt(0)
	}
	digit := byte('0' + digitIndex.Int64())

	bytes := []byte(input)
	bytes[pos] = digit
	return string(bytes)
}

func ensureSymbol(input string) string {
	const symbols = "!@#$%^&*()-_=+[]{}:,.?"
	for i := 0; i < len(input); i++ {
		switch input[i] {
		case '!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '_', '=', '+', '[', ']', '{', '}', ':', ',', '.', '?':
			return input
		}
	}

	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(input))))
	if err != nil {
		index = big.NewInt(0)
	}
	pos := int(index.Int64())

	symbolIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(symbols))))
	if err != nil {
		symbolIndex = big.NewInt(0)
	}
	symbol := symbols[symbolIndex.Int64()]

	bytes := []byte(input)
	bytes[pos] = symbol
	return string(bytes)
}

func ensureCapital(input string) string {
	for i, r := range input {
		if r >= 'a' && r <= 'z' {
			return input[:i] + strings.ToUpper(string(r)) + input[i+1:]
		}
		if r >= 'A' && r <= 'Z' {
			return input
		}
	}
	if len(input) == 0 {
		return "A"
	}
	return "A" + input[1:]
}

func HashPasswordArgon2id(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	timeCost := uint32(3)
	memoryCost := uint32(64 * 1024)
	threads := uint8(2)
	keyLen := uint32(32)

	hash := argon2.IDKey([]byte(password), salt, timeCost, memoryCost, threads, keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memoryCost, timeCost, threads, b64Salt, b64Hash)

	return encoded, nil
}
