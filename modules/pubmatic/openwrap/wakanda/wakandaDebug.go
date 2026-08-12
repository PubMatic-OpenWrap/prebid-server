package wakanda

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"

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

type WakandaDebug interface {
	IsEnable() bool
	SetHTTPRequestData(HTTPRequest *http.Request, HTTPRequestBody json.RawMessage)
	SetHTTPResponseWriter(HTTPResponse http.ResponseWriter)
	SetHTTPResponseBodyWriter(HTTPResponseBody string)
	SetOpenRTB(OpenRTB *openrtb2.BidRequest)
	SetLogger(Logger json.RawMessage)
	SetWinningBid(WinningBid bool)
	SetHttpCalls(HttpCalls json.RawMessage)
	EnableIfRequired(pubIDStr string, profIDStr string)
	WriteLogToFiles()
}

func (wD *Debug) IsEnable() bool {
	return wD.Enabled
}

func (wD *Debug) SetHTTPRequestData(HTTPRequest *http.Request, HTTPRequestBody json.RawMessage) {
	wD.DebugData.HTTPRequest = HTTPRequest
	wD.DebugData.HTTPRequestBody = HTTPRequestBody
}

func (wD *Debug) SetHTTPResponseWriter(HTTPResponse http.ResponseWriter) {
	wD.DebugData.HTTPResponse = HTTPResponse
}

func (wD *Debug) SetHTTPResponseBodyWriter(HTTPResponseBody string) {
	wD.DebugData.HTTPResponseBody = HTTPResponseBody
}

func (wD *Debug) SetOpenRTB(OpenRTB *openrtb2.BidRequest) {
	wD.DebugData.OpenRTB = OpenRTB
}

func (wD *Debug) SetLogger(Logger json.RawMessage) {
	wD.DebugData.Logger = Logger
}

func (wD *Debug) SetWinningBid(WinningBid bool) {
	wD.DebugData.WinningBid = WinningBid
}

func (wD *Debug) SetHttpCalls(HTTPCalls json.RawMessage) {
	wD.DebugData.HTTPCalls = HTTPCalls
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
