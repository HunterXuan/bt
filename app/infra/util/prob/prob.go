package prob

import (
	"math/rand"
	"time"
)

func IfProbGreaterThan(p float32) bool {
	rand.Seed(time.Now().Unix())

	maxVal := 1000
	dividingNum := 1000 * p

	return float32(rand.Intn(maxVal)) > dividingNum
}
