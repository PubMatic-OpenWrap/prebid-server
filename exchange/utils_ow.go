package exchange

import (
	"encoding/json"
	"slices"

	"github.com/golang/glog"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
)

func JLogf(msg string, obj interface{}) {
	if glog.V(3) {
		data, _ := json.Marshal(obj)
		glog.Infof("[OPENWRAP] %v:%v", msg, string(data))
	}
}

// updateContentObjectForBidder updates the content object for each bidder based on content transparency rules
func updateContentObjectForBidder(allBidderRequests []BidderRequest, requestExt *openrtb_ext.ExtRequest) {
	if requestExt == nil || requestExt.Prebid.Transparency == nil || requestExt.Prebid.Transparency.Content == nil {
		return
	}

	rules := requestExt.Prebid.Transparency.Content
	if len(rules) == 0 {
		return
	}

	var content *openrtb2.Content
	isApp := false
	bidderRequest := allBidderRequests[0]

	if bidderRequest.BidRequest.App != nil && bidderRequest.BidRequest.App.Content != nil {
		content = bidderRequest.BidRequest.App.Content
		isApp = true
	} else if bidderRequest.BidRequest.Site != nil && bidderRequest.BidRequest.Site.Content != nil {
		content = bidderRequest.BidRequest.Site.Content
	} else {
		return
	}

	// Dont send content object if no rule and default is not present
	var defaultRule = openrtb_ext.TransparencyRule{}
	if rule, ok := rules["default"]; ok {
		defaultRule = rule
	}

	for _, bidderRequest := range allBidderRequests {
		rule, ok := rules[string(bidderRequest.BidderName)]
		if !ok {
			rule = defaultRule
		}

		newContentObject := deepCopyContentObj(content, rule.Include, rule.Keys)
		updateContentObj(bidderRequest.BidRequest, newContentObject, isApp)
	}
}

func updateContentObj(request *openrtb2.BidRequest, contentObject *openrtb2.Content, isApp bool) {
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

func deepCopyContentNetworkObj(network *openrtb2.Network) *openrtb2.Network {
	if network == nil {
		return nil
	}

	return &openrtb2.Network{
		ID:     network.ID,
		Name:   network.Name,
		Domain: network.Domain,
		Ext:    slices.Clone(network.Ext),
	}
}

func deepCopyContentChannelObj(channel *openrtb2.Channel) *openrtb2.Channel {
	if channel == nil {
		return nil
	}

	return &openrtb2.Channel{
		ID:     channel.ID,
		Name:   channel.Name,
		Domain: channel.Domain,
		Ext:    slices.Clone(channel.Ext),
	}
}

func deepCopyContentProducer(producer *openrtb2.Producer) *openrtb2.Producer {
	if producer == nil {
		return nil
	}

	return &openrtb2.Producer{
		ID:     producer.ID,
		Name:   producer.Name,
		Cat:    slices.Clone(producer.Cat),
		Domain: producer.Domain,
		Ext:    slices.Clone(producer.Ext),
	}
}

func deepCopyContentObj(contentObject *openrtb2.Content, include bool, keys []string) *openrtb2.Content {
	if contentObject == nil {
		return nil
	}

	if !include && len(keys) == 0 {
		return nil
	}

	// Create a deep copy of the content object first
	newContentObject := &openrtb2.Content{
		ID:             contentObject.ID,
		Episode:        contentObject.Episode,
		Title:          contentObject.Title,
		Series:         contentObject.Series,
		Season:         contentObject.Season,
		Artist:         contentObject.Artist,
		Genre:          contentObject.Genre,
		Album:          contentObject.Album,
		ISRC:           contentObject.ISRC,
		URL:            contentObject.URL,
		Cat:            slices.Clone(contentObject.Cat),
		Context:        contentObject.Context,
		ContentRating:  contentObject.ContentRating,
		UserRating:     contentObject.UserRating,
		QAGMediaRating: contentObject.QAGMediaRating,
		Keywords:       contentObject.Keywords,
		Len:            contentObject.Len,
		Language:       contentObject.Language,
	}

	// Deep copy pointer fields
	if contentObject.Producer != nil {
		newContentObject.Producer = deepCopyContentProducer(contentObject.Producer)
	}
	if contentObject.ProdQ != nil {
		prodQ := *contentObject.ProdQ
		newContentObject.ProdQ = &prodQ
	}
	if contentObject.VideoQuality != nil {
		videoQuality := *contentObject.VideoQuality
		newContentObject.VideoQuality = &videoQuality
	}
	if contentObject.LiveStream != nil {
		liveStream := *contentObject.LiveStream
		newContentObject.LiveStream = &liveStream
	}
	if contentObject.SourceRelationship != nil {
		sourceRel := *contentObject.SourceRelationship
		newContentObject.SourceRelationship = &sourceRel
	}
	if contentObject.Embeddable != nil {
		embeddable := *contentObject.Embeddable
		newContentObject.Embeddable = &embeddable
	}
	if contentObject.Data != nil {
		newContentObject.Data = slices.Clone(contentObject.Data)
	}
	if contentObject.Network != nil {
		newContentObject.Network = deepCopyContentNetworkObj(contentObject.Network)
	}
	if contentObject.Channel != nil {
		newContentObject.Channel = deepCopyContentChannelObj(contentObject.Channel)
	}
	newContentObject.Ext = slices.Clone(contentObject.Ext)

	if include && len(keys) == 0 {
		return newContentObject
	}

	// Create a map for O(1) key lookups
	keyMap := make(map[string]bool, len(keys))
	for _, key := range keys {
		keyMap[key] = true
	}

	// Function to clear a field if it's in the keys
	clearField := func(key string) bool {
		return (!include && keyMap[key]) || (include && !keyMap[key])
	}

	// Clear or keep fields based on include flag and keys
	if clearField("id") {
		newContentObject.ID = ""
	}
	if clearField("episode") {
		newContentObject.Episode = 0
	}
	if clearField("title") {
		newContentObject.Title = ""
	}
	if clearField("series") {
		newContentObject.Series = ""
	}
	if clearField("season") {
		newContentObject.Season = ""
	}
	if clearField("artist") {
		newContentObject.Artist = ""
	}
	if clearField("genre") {
		newContentObject.Genre = ""
	}
	if clearField("album") {
		newContentObject.Album = ""
	}
	if clearField("isrc") {
		newContentObject.ISRC = ""
	}
	if clearField("producer") {
		newContentObject.Producer = nil
	}
	if clearField("url") {
		newContentObject.URL = ""
	}
	if clearField("cat") {
		newContentObject.Cat = nil
	}
	if clearField("prodq") {
		newContentObject.ProdQ = nil
	}
	if clearField("videoquality") {
		newContentObject.VideoQuality = nil
	}
	if clearField("context") {
		newContentObject.Context = 0
	}
	if clearField("contentrating") {
		newContentObject.ContentRating = ""
	}
	if clearField("userrating") {
		newContentObject.UserRating = ""
	}
	if clearField("qagmediarating") {
		newContentObject.QAGMediaRating = 0
	}
	if clearField("keywords") {
		newContentObject.Keywords = ""
	}
	if clearField("livestream") {
		newContentObject.LiveStream = nil
	}
	if clearField("sourcerelationship") {
		newContentObject.SourceRelationship = nil
	}
	if clearField("len") {
		newContentObject.Len = 0
	}
	if clearField("language") {
		newContentObject.Language = ""
	}
	if clearField("embeddable") {
		newContentObject.Embeddable = nil
	}
	if clearField("data") {
		newContentObject.Data = nil
	}
	if clearField("network") {
		newContentObject.Network = nil
	}
	if clearField("channel") {
		newContentObject.Channel = nil
	}
	if clearField("ext") {
		newContentObject.Ext = nil
	}

	return newContentObject
}
