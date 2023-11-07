package tracker

import (
	"errors"
	"fmt"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// Inject TrackerCall in Native Adm
func injectNativeCreativeTrackers(native *openrtb2.Native, adm string, tracker models.OWTracker) (string, error) {
	if native == nil {
		return adm, errors.New("native object is missing")
	}
	if len(native.Request) == 0 {
		return adm, errors.New("native request is empty")
	}
	setTrackerURL := false
	callback := func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if err != nil {
			return
		}
		if setTrackerURL {
			return
		}
		adm, setTrackerURL = injectNativeEventTracker(&adm, value, tracker)
	}
	jsonparser.ArrayEach([]byte(native.Request), callback, models.EventTrackers)

	if setTrackerURL {
		return adm, nil
	}
	return injectNativeImpressionTracker(&adm, tracker)
}

// inject tracker in EventTracker Object
func injectNativeEventTracker(adm *string, value []byte, trackerParam models.OWTracker) (string, bool) {
	//Check for event=1
	event, _, _, err := jsonparser.Get(value, models.Event)
	if err != nil || string(event) != models.EventValue {
		return *adm, false
	}
	//Check for method=1
	methodsArray, _, _, err := jsonparser.Get(value, models.Methods) // "[1]","[2]","[1,2]", "[2,1]"
	if err != nil || !strings.Contains(string(methodsArray), models.MethodValue) {
		return *adm, false
	}

	nativeEventTracker := strings.Replace(models.NativeTrackerMacro, "${trackerUrl}", trackerParam.TrackerURL, 1)
	newAdm, err := jsonparser.Set([]byte(*adm), []byte(nativeEventTracker), models.EventTrackers, "[]")
	if err != nil {
		return *adm, false
	}
	*adm = string(newAdm)
	return *adm, true
}

// inject tracker in ImpTracker Object
func injectNativeImpressionTracker(adm *string, tracker models.OWTracker) (string, error) {
	impTrackers := []string{}
	callback := func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		impTrackers = append(impTrackers, string(value))
	}
	jsonparser.ArrayEach([]byte(*adm), callback, models.ImpTrackers)
	//append trackerUrl
	impTrackers = append(impTrackers, tracker.TrackerURL)
	allImpTrackers := fmt.Sprintf(`["%s"]`, strings.Join(impTrackers, `","`))
	newAdm, err := jsonparser.Set([]byte(*adm), []byte(allImpTrackers), models.ImpTrackers)
	if err != nil {
		return *adm, errors.New("error setting imptrackers in native adm")
	}
	*adm = string(newAdm)
	return *adm, nil
}
