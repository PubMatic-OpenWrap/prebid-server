package adapters

import (
	"encoding/base64"

	"github.com/google/uuid"
	"github.com/mxmCherry/openrtb/v16/openrtb2"
)

const (
	USER_AGE               = "target_age"
	GENDER_MALE            = "Male"
	GENDER_FEMALE          = "Female"
	GENDER_OTHER           = "Others"
	DEVICE_COMPUTER        = "Personal Computer"
	DEVICE_PHONE           = "Phone"
	DEVICE_TABLET          = "Tablet"
	DEVICE_CONNECTEDDEVICE = "Connected Devices"
	USER_GENDER            = "target_gender"
	COUNTRY                = "target_country"
	REGION                 = "target_region"
	CITY                   = "target_city"
	DEVICE                 = "target_device"
	STRING_TRUE            = "true"
	STRING_FALSE           = "false"
)

func AddDefaultFieldsComm(bid *openrtb2.Bid) {
	if bid != nil {
		bid.CrID = "DefaultCRID"
	}
}

func GenerateUniqueBidIDComm() string {
	id := uuid.New()
	return id.String()
}

func EncodeURL(url string) string {
	str := base64.StdEncoding.EncodeToString([]byte(url))
	return str
}
