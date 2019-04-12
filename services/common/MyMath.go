package common

type (
	// SeededRandom struct
	SeededRandom struct {
		RandomSeed int32
	}
)

// SeededRandom return a mock random value
func (s *SeededRandom) SeededRandom(min, max float64) float64 {
	s.RandomSeed = int32((int64(s.RandomSeed)*9301 + 49297) % 233280)
	random := float64(s.RandomSeed) / float64(233280)

	return min + random*(max-min)
}
