package wakanda

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	FilePrefix  map[string]string // This will contain encodedFilters value received in the wakanda request
}

type WakandaDebug interface {
	IsEnable() bool
	SetHTTPRequestData(HTTPRequest *http.Request, HTTPRequestBody json.RawMessage)
	SetHTTPResponseWriter(HTTPResponse http.ResponseWriter)
	SetHTTPResponseBodyWriter(HTTPResponseBody string)
	SetOpenRTB(OpenRTB *openrtb2.BidRequest)
	SetLogger(Logger json.RawMessage)
	SetWinningBid(WinningBid bool)
	EnableIfRequired(pubIDStr string, profIDStr string)
	WriteLogToFiles()
	SetBadRequest()
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

func (wD *Debug) SetBadRequest() {
	wD.DebugData.isBadRequest = true
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
		wD.FilePrefix = make(map[string]string)
		for _, key := range generateKeysFromHBRequest(pubIDStr, profIDStr) {
			if wakandaRulesMap.IsRulePresent(key) {
				aWakandaRule := wakandaRulesMap.Incr(key)
				if aWakandaRule != nil {
					// enable wakanda
					logger.Info("Wakanda is enabled for %s %s", pubIDStr, profIDStr)
					wD.Enabled = true
					wD.FolderPaths = append(wD.FolderPaths, aWakandaRule.FolderPath)
					wD.DebugLevel = aWakandaRule.DebugLevel
					wD.FilePrefix[aWakandaRule.FolderPath] = aWakandaRule.Filters
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

			if filePrefix, ok := wD.FilePrefix[logDir]; ok {
				sftpDestinationFile = generateDestinationFile(filePrefix, wD.DebugData, logDir)
			}
			if sftpDestinationFile != "" {
				if err := send(sftpDestinationFile, logDir, recordBytes, wD.Config.SFTP); err != nil {
					logger.Error("Wakanda '%s' SFTP Error : %s", sftpDestinationFile, err.Error())
				}
			}

		}
	}
}

func generateDestinationFile(filePrefix string, debugData DebugData, logDir string) string {
	if debugData.isBadRequest {
		if strings.Contains(logDir, "FILTERS:badrequest") {
			return filePrefix + "_" + fmt.Sprint(time.Now().UnixNano())
		}
		return ""
	}
	if strings.Contains(logDir, "FILTERS:winningbidandzerobid") {
		if debugData.WinningBid {
			return filePrefix + "_" + fmt.Sprint(time.Now().UnixNano()) + ".winningbid"
		}
		return filePrefix + "_" + fmt.Sprint(time.Now().UnixNano())
	}
	return ""
}

func (wD *Debug) SetBadRequestFlag() {
	wD.DebugData.isBadRequest = true
}

// TrimFilters trims the "__FILTERS" substring from the given filter string
func TrimFilters(filter string) string {
	index := strings.Index(filter, "__FILTERS")
	if index != -1 {
		return filter[:index]
	}
	return ""
}
