package models

import (
	"encoding/json"
	"fmt"
)

func GetRequestExt(ext []byte) (*RequestExt, error) {
	extRequest := &RequestExt{}

	err := json.Unmarshal(ext, extRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to decode request.ext : %v", err)
	}

	return extRequest, nil
}
