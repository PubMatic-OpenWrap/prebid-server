package wakanda

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	//CMaxTraceCount maximum trace request can be logged
	CMaxTraceCount = 20
	//CAPIDebugLevel debug level parameter of wakanda handler
	CAPIDebugLevel = "debugLevel"
	//CAPIPublisherID publisher id paramater of wakanda handler
	CAPIPublisherID = "pubId"
	//CAPIProfileID profile id parameter of wakanda handler
	CAPIProfileID = "profId"
	//CRuleKeyPubProfile rule format ,same is used for folder name with "__DC"
	CRuleKeyPubProfile = "PUB:%s__PROF:%s"
)

var wakandaRulesMap *rulesMap

// generateKeyFromWakandaRequest returns only one qualifying rule
func generateKeyFromWakandaRequest(pubIDStr string, profIDStr string) string {
	// create a rule key based on the request params
	// note: for new rule-key extra code will be needed to be added

	if len(pubIDStr) > 3 && len(profIDStr) >= 1 {
		return fmt.Sprintf(CRuleKeyPubProfile, pubIDStr, profIDStr)
	}

	if len(pubIDStr) > 3 {
		return fmt.Sprintf(CRuleKeyPubProfile, pubIDStr, "0") // setting profile id as 0 for all profile ids
	}

	return ""
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
		fmt.Sprintf(CRuleKeyPubProfile, pubIDStr, profIDStr),
		fmt.Sprintf(CRuleKeyPubProfile, pubIDStr, "0"), // setting profile id as 0 for all profile ids
	)
	return
}

// Handler take the GET request input
func Handler(config Wakanda) http.HandlerFunc {
	return func(httpRespWriter http.ResponseWriter, httpRequest *http.Request) {

		debugLevel, _ := strconv.Atoi(httpRequest.FormValue(CAPIDebugLevel))
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

		key := generateKeyFromWakandaRequest(httpRequest.FormValue(CAPIPublisherID), httpRequest.FormValue(CAPIProfileID))
		if len(key) > 0 {

			if wakandaRulesMap.AddIfNotPresent(key, debugLevel, config.DCName) {
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
