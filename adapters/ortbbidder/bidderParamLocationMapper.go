package ortbbidder

import (
	"encoding/json"
	"os"
	"strings"
)

type responseMapper map[string]response

var respMapper responseMapper

func NewResponseMapper() responseMapper {
	if respMapper == nil {
		respMapper = make(responseMapper)
		SetFromFile()
	}

	return respMapper
}

type response struct {
	BidType responseMapper `json:"responseMapper"`
}

func SetFromFile() {
	files, err := os.ReadDir("static/bidder-params")
	if err != nil {
		return
	}
	for _, file := range files {
		if !strings.HasPrefix(file.Name(), "owortb_") {
			continue
		}
		fileBytes, err := os.ReadFile("static/bidder-params/" + file.Name())
		if err != nil {
			return
		}
		resp := response{}
		err = json.Unmarshal(fileBytes, &resp.BidType)
		if err != nil {
			continue
		}
		suff, _ := strings.CutSuffix(file.Name(), ".json")
		respMapper[suff] = resp
	}
}

func (rm responseMapper) Get(key string) response {
	return rm[key]
}
