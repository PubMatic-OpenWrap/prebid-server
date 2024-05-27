package wakanda

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/glog"

	"git.pubmatic.com/PubMatic/go-common/logger"
)

// Debug debug structure hold data for each request
type Debug struct {
	Enabled     bool
	FolderPaths []string
	DebugLevel  int
	DebugData   DebugData
	Config      Wakanda
}

// EnableIfRequired will check if rule is applicable or not
// Arguments:
//
//	Debug
//	array of generated keys
//
// Flow
//
//	For each passed keys
//		if entry is present in wakandaRulesMap
//			set the waknada data in HB request
//			increment the count wakandaRulesMap entry; consider maxTraceCount
func (wD *Debug) EnableIfRequired(pubIDStr string, profIDStr string) {
	if !wakandaRulesMap.IsEmpty() {
		for _, key := range generateKeysFromHBRequest(pubIDStr, profIDStr) {
			if wakandaRulesMap.IsRulePresent(key) {
				aWakandaRule := wakandaRulesMap.Incr(key)
				if aWakandaRule != nil {
					// enable wakanda
					logger.Info("Wakanda is enabled for %s %s", pubIDStr, profIDStr)
					wD.Enabled = true
					wD.FolderPaths = append(wD.FolderPaths, aWakandaRule.FolderPath)
					wD.DebugLevel = aWakandaRule.DebugLevel
				}
			}
		}
	}
}

// WriteLogToFiles writes log to file
func (wD *Debug) WriteLogToFiles() {
	if wD.DebugLevel == 2 { // todo remove hard-coding
		record := NewLogRecord(&wD.DebugData)
		recordBytes, err := json.Marshal(record)
		if err != nil {
			for _, logDir := range wD.FolderPaths {
				logger.Error("WAKANDA-ERROR: WAKANDA-HB-S2S-%s, %v", logDir, err)
			}
		}

		for _, logDir := range wD.FolderPaths {
			logger.Info("WAKANDA-HB-S2S-%s, %s", logDir, string(recordBytes))
			glog.Flush()
			// <POD_NAME>-<UNIX_TIME_NANO_SEC>.json
			var sftpDestinationFile string
			if wD.DebugData.WinningBid {
				sftpDestinationFile = fmt.Sprintf("%s-%d.json.winningbid", wD.Config.PodName, time.Now().UnixNano())
			} else {
				sftpDestinationFile = fmt.Sprintf("%s-%d.json", wD.Config.PodName, time.Now().UnixNano())
			}
			if err := send(sftpDestinationFile, logDir, recordBytes, wD.Config.SFTP); err != nil {
				logger.Error("Wakanda '%s' SFTP Error : %s", sftpDestinationFile, err.Error())
			}
		}
	}
}
