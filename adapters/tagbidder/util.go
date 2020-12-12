package tagbidder

import (
	"bytes"
)

func objectArrayToString(len int, separator string, cb func(i int) string) string {
	if 0 == len {
		return ""
	}

	var out bytes.Buffer
	for i := 0; i < len; i++ {
		if out.Len() > 0 {
			out.WriteString(separator)
		}
		out.WriteString(cb(i))
	}
	return out.String()
}
