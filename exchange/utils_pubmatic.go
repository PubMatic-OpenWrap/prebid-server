package exchange

import (
	"reflect"
	"strings"

	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
)

// updateContentObjectForBidder updates the content object for each bidder based on content transparency rules
func updateContentObjectForBidder(allBidderRequests []BidderRequest, requestExt *openrtb_ext.ExtRequest) {
	if requestExt == nil || requestExt.Prebid.Transparency == nil || requestExt.Prebid.Transparency.Content == nil {
		return
	}

	var contentObject *openrtb2.Content
	isApp := false
	bidderRequest := allBidderRequests[0]
	if bidderRequest.BidRequest.App != nil && bidderRequest.BidRequest.App.Content != nil {
		contentObject = bidderRequest.BidRequest.App.Content
		isApp = true
	} else if bidderRequest.BidRequest.Site != nil && bidderRequest.BidRequest.Site.Content != nil {
		contentObject = bidderRequest.BidRequest.Site.Content
	} else {
		return
	}

	rules := requestExt.Prebid.Transparency.Content

	// Dont send content object if no rule and default is not present
	var defaultRule = openrtb_ext.TransparencyRule{}
	if rule, ok := rules["default"]; ok {
		defaultRule = rule
	}

	for _, bidderRequest := range allBidderRequests {
		var newContentObject *openrtb2.Content
		if len(rules) != 0 {
			rule, ok := rules[string(bidderRequest.BidderName)]
			if !ok {
				rule = defaultRule
			}

			if len(rule.Keys) != 0 {
				newContentObject = createNewContentObject(newContentObject, contentObject, rule.Include, rule.Keys)
			} else if rule.Include {
				newContentObject = contentObject
			}
		}
		deepCopyContentObj(bidderRequest.BidRequest, newContentObject, isApp)
	}
}

func deepCopyContentObj(request *openrtb2.BidRequest, contentObject *openrtb2.Content, isApp bool) {
	if isApp {
		app := *request.App
		app.Content = contentObject
		request.App = &app
	} else {
		site := *request.Site
		site.Content = contentObject
		request.Site = &site
	}
}

func createNewContentObject(newContentObject, contentObject *openrtb2.Content, include bool, keys []string) *openrtb2.Content {
	if include {
		return includeKeys(contentObject, keys)
	}
	return excludeKeys(contentObject, keys)

}

func excludeKeys(contentObject *openrtb2.Content, keys []string) *openrtb2.Content {
	newContentObject := *contentObject

	keyMap := make(map[string]struct{}, 1)
	for _, key := range keys {
		keyMap[key] = struct{}{}
	}

	rt := reflect.TypeOf(newContentObject)
	for i := 0; i < rt.NumField(); i++ {
		key := strings.Split(rt.Field(i).Tag.Get("json"), ",")[0] // remove omitempty, etc
		if _, ok := keyMap[key]; ok {
			reflect.ValueOf(&newContentObject).Elem().FieldByName(rt.Field(i).Name).Set(reflect.Zero(rt.Field(i).Type))
		}
	}

	return &newContentObject
}

func includeKeys(contentObject *openrtb2.Content, keys []string) *openrtb2.Content {
	newContentObject := openrtb2.Content{}
	v := reflect.ValueOf(contentObject).Elem()
	keyMap := make(map[string]struct{}, 1)
	for _, key := range keys {
		keyMap[key] = struct{}{}
	}

	rt := reflect.TypeOf(newContentObject)
	for i := 0; i < rt.NumField(); i++ {
		key := strings.Split(rt.Field(i).Tag.Get("json"), ",")[0] // remove omitempty, etc
		if _, ok := keyMap[key]; ok {
			name := rt.Field(i).Name
			reflect.ValueOf(&newContentObject).Elem().FieldByName(rt.Field(i).Name).Set(v.FieldByName(name))
		}
	}

	return &newContentObject
}
