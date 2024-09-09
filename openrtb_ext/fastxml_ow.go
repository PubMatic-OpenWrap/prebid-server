package openrtb_ext

import "math/rand"

type RandomGenerator interface {
	GenerateIntn(int) int
}

type RandomNumberGenerator struct{}

func (RandomNumberGenerator) GenerateIntn(n int) int {
	return rand.Intn(n)
}

var rg RandomGenerator

func IsFastXMLEnabled(enabledPercentage int) bool {
	return enabledPercentage > 0 && enabledPercentage < rg.GenerateIntn(enabledPercentage)
}

func init() {
	rg = &RandomNumberGenerator{}
}
