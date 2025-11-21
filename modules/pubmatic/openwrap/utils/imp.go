package utils

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	V25_IMP_ID_FORMAT     = "%s::%s::%d"
	ImpressionIDSeparator = `::`
)

func PopulateV25ImpID(podId string, impId string, seq int) string {
	return fmt.Sprintf(V25_IMP_ID_FORMAT, podId, impId, seq)
}

func DecodeV25ImpID(id string) (podId string, impId string, seq int) {
	str := strings.Split(id, ImpressionIDSeparator)
	if len(str) != 3 {
		return "", str[0], 0
	}

	podId = str[0]
	impId = str[1]
	seq, _ = strconv.Atoi(str[2])
	return
}
