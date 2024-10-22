package gocache

import (
	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// PopulatePublisherVASTTags will put publisher level VAST Tag details into cache
func (c *cache) populatePublisherVASTTags(pubID int) error {
	cacheKey := key(PubVASTTags, pubID)

	//get publisher level vast tag details from DB
	publisherVASTTags, err := c.db.GetPublisherVASTTags(pubID)
	if err != nil {
		glog.Errorf("[QueryFailure] for PublisherVASTTags with err: %v", err)
		return err
	}

	if publisherVASTTags == nil {
		publisherVASTTags = models.PublisherVASTTags{}
	}

	c.cache.Set(cacheKey, publisherVASTTags, getSeconds(c.cfg.VASTTagCacheExpiry))
	return nil
}

// GetPublisherVASTTagsFromCache read publisher level vast tag details from cache
func (c *cache) GetPublisherVASTTagsFromCache(pubID int) models.PublisherVASTTags {
	cacheKey := key(PubVASTTags, pubID)
	if value, ok := c.cache.Get(cacheKey); ok && value != nil {
		return value.(models.PublisherVASTTags)
	}
	//if found then return actual value or else return empty
	return models.PublisherVASTTags{}
}
