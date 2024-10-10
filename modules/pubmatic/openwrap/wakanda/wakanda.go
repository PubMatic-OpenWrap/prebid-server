package wakanda

import (
	"fmt"
	"net/http"
	"strconv"
)

var wakandaRulesMap *rulesMap

// generateKeyFromWakandaRequest returns only one qualifying rule
func generateKeyFromWakandaRequest(pubIDStr string, profIDStr string) string {
	// create a rule key based on the request params
	// note: for new rule-key extra code will be needed to be added

	if len(pubIDStr) > 3 && len(profIDStr) >= 1 {
		return fmt.Sprintf(cRuleKeyPubProfile, pubIDStr, profIDStr)
	}

	if len(pubIDStr) > 3 {
		return fmt.Sprintf(cRuleKeyPubProfile, pubIDStr, "0") // setting profile id as 0 for all profile ids
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
		fmt.Sprintf(cRuleKeyPubProfile, pubIDStr, profIDStr),
		fmt.Sprintf(cRuleKeyPubProfile, pubIDStr, "0"), // setting profile id as 0 for all profile ids
	)
	return
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

		key := generateKeyFromWakandaRequest(httpRequest.FormValue(cAPIPublisherID), httpRequest.FormValue(cAPIProfileID))
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

func TestInstance() func() {
	wakandaRulesMap = &rulesMap{
		rules: make(map[string]*wakandaRule),
	}

	key := "PUB:111__PROF:222"
	wakandaRulesMap.AddIfNotPresent(key, 2, "DC1")
	return func() {
		wakandaRulesMap = nil
	}
}
