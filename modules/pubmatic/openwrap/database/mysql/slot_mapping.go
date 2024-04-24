package mysql

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

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
		return partnerSlotMappingMap, &models.DBError{Message: err.Error()}
	}
	defer rows.Close()

	for rows.Next() {
		var slotMapping = models.SlotMapping{}
		err := rows.Scan(&slotMapping.PartnerId, &slotMapping.AdapterId, &slotMapping.VersionId, &slotMapping.SlotName, &slotMapping.MappingJson, &slotMapping.OrderID)
		if err != nil {
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
	return fmt.Sprintf(db.cfg.Queries.GetSlotNameHash, pubID)
}
