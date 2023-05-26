package stats

// stat represents a single stat along with its value
type stat struct {
	Key   string
	Value int
}

// validateStatKey will be used to check the characters that are in the stat key.
// Unwanted characters can removed or replaced with different characters.
func (s *stat) validateStatKey() string {
	//TODO: Add some validation on key. Invalid keys can be changed or can be logged via sc.logger
	return s.Key
}
