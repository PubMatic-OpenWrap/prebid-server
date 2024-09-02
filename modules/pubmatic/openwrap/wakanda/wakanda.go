package wakanda

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

var wakandaRulesMap *rulesMap

// generateKeyFromWakandaRequest returns only one qualifying rule
func generateKeyFromWakandaRequest(pubIDStr string, profIDStr string, filters string) string {
	// create a rule key based on the request params
	// note: for new rule-key extra code will be needed to be added

	// create a rule key based on the request params
	key := ""
	if len(pubIDStr) <= 3 {
		return ""
	}
	if profIDStr == "" {
		profIDStr = "0"
	}
	if filters == "" {
		key = "PUB:" + pubIDStr + "__PROF:" + profIDStr
	} else {
		key = "PUB:" + pubIDStr + "__PROF:" + profIDStr + "__FILTERS:" + filters
	}
	return key
}

// Arguments:
//
//	publisher Id
//	profile Id
//
// Flow
//
//	Return an array of valid keys
func generateKeysFromHBRequest(pubIDStr string, profIDStr string) (generatedKeys []string) {
	generatedKeys = append(
		generatedKeys,
		// key with valid filters conditions
		"PUB:"+pubIDStr+"__PROF:"+profIDStr+"__FILTERS:badrequest",
		"PUB:"+pubIDStr+"__PROF:"+profIDStr+"__FILTERS:winningbidandzerobid",
		"PUB:"+pubIDStr+"__PROF:"+profIDStr,

		"PUB:"+pubIDStr+"__PROF:"+"0"+"__FILTERS:badrequest",
		"PUB:"+pubIDStr+"__PROF:"+"0"+"__FILTERS:winningbidandzerobid",
		"PUB:"+pubIDStr+"__PROF:"+"0",
	)
	return
}

// decodeFilters decodes the Base64 encoded string back to the original pubID, profID, and filter. Currently, we support the "badrequest" filter and the "winningbidandzerobid" filter.
func decodeFilters(encodedFilters string) (string, string, string) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedFilters)
	if err != nil {
		return "", "", ""
	}
	decodedStr := string(decodedBytes)
	parts := strings.Split(decodedStr, ":")
	if len(parts) != 3 {
		return "", "", ""
	}

	return parts[0], parts[1], parts[2]
}

// Handler take the GET request input
func Handler(config Wakanda) http.HandlerFunc {
	return func(httpRespWriter http.ResponseWriter, httpRequest *http.Request) {

		debugLevel, _ := strconv.Atoi(httpRequest.FormValue(cAPIDebugLevel))
		if debugLevel == 0 {
			debugLevel = 1 // default debugLevel
			// 1: pbs debug 1
			// 2: with files
		}

		// if a value more than the known value is set then set to 2
		if debugLevel > 2 {
			debugLevel = 2
		}

		successStatus := "true"
		statusMsg := ""

		encodedFilters := httpRequest.FormValue(cAPIFilters)
		_, _, filters := decodeFilters(encodedFilters)

		key := generateKeyFromWakandaRequest(httpRequest.FormValue(cAPIPublisherID), httpRequest.FormValue(cAPIProfileID), filters)
		if len(key) > 0 {

			if wakandaRulesMap.AddIfNotPresent(key, debugLevel, config.DCName, encodedFilters) {
				statusMsg = "New key generated."
			} else {
				statusMsg = "Key already exists."
			}
		} else {
			// invalid key
			successStatus = "false"
			statusMsg = "No key was generated for the request."
		}

		// return jSON response with status
		httpRespWriter.WriteHeader(http.StatusOK)
		httpRespWriter.Write([]byte(fmt.Sprintf("{success: \"%s\", statusMsg: \"%s\", host: \"%s\"}", successStatus, statusMsg, config.HostName)))
	}
}

func Init(config Wakanda) {
	wakandaRulesMap = getNewRulesMap(config)
	setCommandHandler()
}

func setCommandHandler() {
	commandHandler = &CommandHandler{
		commandExecutor: &CommandHandler{},
	}
}
