package utils

import (
	"strconv"
	"strings"
)

// return int array by tokenizing the input string on
// provided delimiter
func GetIntArrayFromString(str, separtor string) []int {
	intArray := make([]int, 0)

	tokens := strings.Split(str, separtor)
	for _, token := range tokens {
		val, err := strconv.Atoi(token)
		if err == nil {
			intArray = append(intArray, val)
		}
	}
	return intArray
}
