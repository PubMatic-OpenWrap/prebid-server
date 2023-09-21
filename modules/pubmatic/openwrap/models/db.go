package models

import "strings"

// VASTTag contains tag details of VASTBidders
type VASTTag struct {
	ID        int     `json:"id,omitempty"`
	PartnerID int     `json:"partnerId,omitempty"`
	URL       string  `json:"url,omitempty"`
	Duration  int     `json:"dur,omitempty"`
	Price     float64 `json:"price,omitempty"`
}

// PublisherVASTTags holds publisher level vast tag entries
type PublisherVASTTags = map[int]*VASTTag

/*SlotMappingInfo contains the ordered list of slot names and a map of slot names to their hash values*/
type SlotMappingInfo struct {
	OrderedSlotList []string
	HashValueMap    map[string]string
}

type SlotInfo struct {
	SlotName string
	AdSize   string
	AdWidth  int
	AdHeight int
	SiteId   int
	AdTagId  int
	GId      int // Gauanteed Id
	Floor    float64
}

/*SlotMapping object contains information for a given slot*/
type SlotMapping struct {
	PartnerId    int64
	AdapterId    int64
	VersionId    int64
	SlotName     string
	MappingJson  string
	SlotMappings map[string]interface{}
	Hash         string
	OrderID      int64
}

type BySlotName []*SlotInfo

func (t BySlotName) Len() int { return len(t) }
func (t BySlotName) Compare(i int, element interface{}) int {
	slotname := element.(string)
	return strings.Compare(t[i].SlotName, slotname)
}

// AdUnitConfig type definition for Ad Unit config parsed from stored config JSON
type AdUnitConfig map[string]interface{}
