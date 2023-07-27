package mysql

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

// Return the list of Pubmatic slot mappings
func (db *mySqlDB) GetPubmaticSlotMappings(pubID int) (map[string]models.SlotMapping, error) {
	pmSlotMappings := make(map[string]models.SlotMapping, 0)
	rows, err := db.conn.Query(db.cfg.Queries.GetPMSlotToMappings,
		pubID, models.MAX_SLOT_COUNT)
	if nil != err {
		return pmSlotMappings, err
	}

	defer rows.Close()
	for rows.Next() {
		slotInfo := models.SlotInfo{}
		slotMapping := models.SlotMapping{}

		err := rows.Scan(&slotInfo.SlotName, &slotInfo.AdSize, &slotInfo.SiteId,
			&slotInfo.AdTagId, &slotInfo.GId, &slotInfo.Floor)
		if nil != err {
			//continue
		}
		slotMapping.PartnerId = models.PUBMATIC_PARTNER_ID //hardcoding partnerId for pubmatic
		slotMapping.AdapterId = models.PUBMATIC_ADAPTER_ID //hardcoding adapterId for pubmatic
		slotMapping.SlotName = slotInfo.SlotName           //+ "@" + slotInfo.AdSize
		//adtag, site, floor hardcoded as this code is to be removed once pmapi moves to wrapper workflow
		slotMapping.MappingJson =
			"{\"adtag\":\"" + strconv.Itoa(slotInfo.AdTagId) + "\"," +
				"\"site\":\"" + strconv.Itoa(slotInfo.SiteId) + "\"," +
				"\"floor\":\"" + strconv.FormatFloat(slotInfo.Floor, 'f', 2, 64) + "\"," +
				"\"gaid\":\"" + strconv.Itoa(slotInfo.GId) + "\"}"
		var mappingJsonObj map[string]interface{}
		if err := json.Unmarshal([]byte(slotMapping.MappingJson), &mappingJsonObj); err != nil {
			continue
		}

		//Adding slotName from DB in fieldMap for PubMatic, as slotName from DB should be sent to PubMatic instead of slotName from request
		//This is required for case in-sensitive mapping
		mappingJsonObj[models.KEY_OW_SLOT_NAME] = slotMapping.SlotName

		slotMapping.SlotMappings = mappingJsonObj
		pmSlotMappings[strings.ToLower(slotMapping.SlotName)] = slotMapping
	}
	return pmSlotMappings, nil
}

// GetPublisherSlotNameHash Returns a map of all slot names and hashes for a publisher
func (db *mySqlDB) GetPublisherSlotNameHash(pubID int) (map[string]string, error) {
	nameHashMap := make(map[string]string)

	query := db.formSlotNameHashQuery(pubID)
	rows, err := db.conn.Query(query)
	if err != nil {
		return nameHashMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var name, hash string
		if err = rows.Scan(&name, &hash); err != nil {
			continue
		}
		nameHashMap[name] = hash
	}

	//vastTagHookPublisherSlotName(nameHashMap, pubID)
	return nameHashMap, nil
}

// Return the list of wrapper slot mappings
func (db *mySqlDB) GetWrapperSlotMappings(partnerConfigMap map[int]map[string]string, profileID, displayVersion int) (map[int][]models.SlotMapping, error) {
	partnerSlotMappingMap := make(map[int][]models.SlotMapping)

	query := db.formWrapperSlotMappingQuery(profileID, displayVersion, partnerConfigMap)
	rows, err := db.conn.Query(query)
	if err != nil {
		return partnerSlotMappingMap, err
	}
	defer rows.Close()

	for rows.Next() {
		var slotMapping = models.SlotMapping{}
		err := rows.Scan(&slotMapping.PartnerId, &slotMapping.AdapterId, &slotMapping.VersionId, &slotMapping.SlotName, &slotMapping.MappingJson, &slotMapping.OrderID)
		if nil != err {
			continue
		}

		slotMappingList, found := partnerSlotMappingMap[int(slotMapping.PartnerId)]
		if found {
			slotMappingList = append(slotMappingList, slotMapping)
			partnerSlotMappingMap[int(slotMapping.PartnerId)] = slotMappingList
		} else {
			newSlotMappingList := make([]models.SlotMapping, 0)
			newSlotMappingList = append(newSlotMappingList, slotMapping)
			partnerSlotMappingMap[int(slotMapping.PartnerId)] = newSlotMappingList
		}

	}
	//vastTagHookPartnerSlotMapping(partnerSlotMappingMap, profileId, displayVersion)
	return partnerSlotMappingMap, nil
}

// GetMappings will returns slotMapping from map based on slotKey
func (db *mySqlDB) GetMappings(slotKey string, slotMap map[string]models.SlotMapping) (map[string]interface{}, error) {
	slotMappingObj, present := slotMap[strings.ToLower(slotKey)]
	if !present {
		return nil, errors.New("No mapping found for slot:" + slotKey)
	}
	fieldMap := slotMappingObj.SlotMappings
	return fieldMap, nil
}

func (db *mySqlDB) formWrapperSlotMappingQuery(profileID, displayVersion int, partnerConfigMap map[int]map[string]string) string {
	var query string
	var partnerIDStr string
	for partnerID := range partnerConfigMap {
		partnerIDStr = partnerIDStr + strconv.Itoa(partnerID) + ","
	}
	partnerIDStr = strings.TrimSuffix(partnerIDStr, ",")

	if displayVersion != 0 {
		query = strings.Replace(db.cfg.Queries.GetWrapperSlotMappingsQuery, profileIdKey, strconv.Itoa(profileID), -1)
		query = strings.Replace(query, displayVersionKey, strconv.Itoa(displayVersion), -1)
		query = strings.Replace(query, partnerIdKey, partnerIDStr, -1)
	} else {
		query = strings.Replace(db.cfg.Queries.GetWrapperLiveVersionSlotMappings, profileIdKey, strconv.Itoa(profileID), -1)
		query = strings.Replace(query, partnerIdKey, partnerIDStr, -1)
	}
	return query
}

func (db *mySqlDB) formSlotNameHashQuery(pubID int) (query string) {
	//TODO : Remove #PUB_ID from GetSlotNameHash Query
	return fmt.Sprint(db.cfg.Queries.GetSlotNameHash, pubID)
}
