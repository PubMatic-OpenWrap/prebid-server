package openwrap

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/diegoholiveira/jsonlogic/v3"
	"github.com/prebid/openrtb/v19/openrtb2"
	cache "github.com/prebid/prebid-server/modules/pubmatic/openwrap/cache"
	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models"
)

var countryAlpha3ToAlpha2Code = map[string]string{
	"AFG": "AF",
	"ALA": "AX",
	"ALB": "AL",
	"DZA": "DZ",
	"ASM": "AS",
	"AND": "AD",
	"AGO": "AO",
	"AIA": "AI",
	"ATA": "AQ",
	"ATG": "AG",
	"ARG": "AR",
	"ARM": "AM",
	"ABW": "AW",
	"AUS": "AU",
	"AUT": "AT",
	"AZE": "AZ",
	"BEQ": "BQ",
	"BHS": "BS",
	"BHR": "BH",
	"BGD": "BD",
	"BRB": "BB",
	"BLR": "BY",
	"BEL": "BE",
	"BLZ": "BZ",
	"BEN": "BJ",
	"BMU": "BM",
	"BTN": "BT",
	"BOL": "BO",
	"BIH": "BA",
	"BWA": "BW",
	"BVT": "BV",
	"BRA": "BR",
	"VGB": "VG",
	"IOT": "IO",
	"BRN": "BN",
	"BGR": "BG",
	"BFA": "BF",
	"BDI": "BI",
	"KHM": "KH",
	"CMR": "CM",
	"CAN": "CA",
	"CPV": "CV",
	"CYM": "KY",
	"CAF": "CF",
	"CUW": "CW",
	"TCD": "TD",
	"CHL": "CL",
	"CHN": "CN",
	"HKG": "HK",
	"MAC": "MO",
	"CXR": "CX",
	"CCK": "CC",
	"COL": "CO",
	"COM": "KM",
	"COG": "CG",
	"COD": "CD",
	"COK": "CK",
	"CRI": "CR",
	"CIV": "CI",
	"HRV": "HR",
	"CUB": "CU",
	"CYP": "CY",
	"CZE": "CZ",
	"DNK": "DK",
	"DJI": "DJ",
	"DMA": "DM",
	"DOM": "DO",
	"ECU": "EC",
	"EGY": "EG",
	"SLV": "SV",
	"GNQ": "GQ",
	"ERI": "ER",
	"EST": "EE",
	"ETH": "ET",
	"FLK": "FK",
	"FRO": "FO",
	"FJI": "FJ",
	"FIN": "FI",
	"FRA": "FR",
	"GUF": "GF",
	"PYF": "PF",
	"ATF": "TF",
	"GAB": "GA",
	"GMB": "GM",
	"GEO": "GE",
	"DEU": "DE",
	"GHA": "GH",
	"GIB": "GI",
	"GRC": "GR",
	"GRL": "GL",
	"GRD": "GD",
	"GLP": "GP",
	"GUM": "GU",
	"GTM": "GT",
	"GGY": "GG",
	"GIN": "GN",
	"GNB": "GW",
	"GUY": "GY",
	"HTI": "HT",
	"HMD": "HM",
	"VAT": "VA",
	"HND": "HN",
	"HUN": "HU",
	"ISL": "IS",
	"IND": "IN",
	"IDN": "ID",
	"IRN": "IR",
	"IRQ": "IQ",
	"IRL": "IE",
	"IMN": "IM",
	"ISR": "IL",
	"ITA": "IT",
	"JAM": "JM",
	"JPN": "JP",
	"JEY": "JE",
	"JOR": "JO",
	"KAZ": "KZ",
	"KEN": "KE",
	"KIR": "KI",
	"PRK": "KP",
	"KOR": "KR",
	"KWT": "KW",
	"KGZ": "KG",
	"LAO": "LA",
	"LVA": "LV",
	"LBN": "LB",
	"LSO": "LS",
	"LBR": "LR",
	"LBY": "LY",
	"LIE": "LI",
	"LTU": "LT",
	"LUX": "LU",
	"MKD": "MK",
	"MDG": "MG",
	"MWI": "MW",
	"MYS": "MY",
	"MDV": "MV",
	"MLI": "ML",
	"MLT": "MT",
	"MHL": "MH",
	"MTQ": "MQ",
	"MRT": "MR",
	"MUS": "MU",
	"MYT": "YT",
	"MEX": "MX",
	"FSM": "FM",
	"MDA": "MD",
	"MCO": "MC",
	"MNG": "MN",
	"MNE": "ME",
	"MSR": "MS",
	"MAR": "MA",
	"MOZ": "MZ",
	"MMR": "MM",
	"NAM": "NA",
	"NRU": "NR",
	"NPL": "NP",
	"NLD": "NL",
	"ANT": "AN",
	"NCL": "NC",
	"NZL": "NZ",
	"NIC": "NI",
	"NER": "NE",
	"NGA": "NG",
	"NIU": "NU",
	"NFK": "NF",
	"MNP": "MP",
	"NOR": "NO",
	"OMN": "OM",
	"PAK": "PK",
	"PLW": "PW",
	"PSE": "PS",
	"PAN": "PA",
	"PNG": "PG",
	"PRY": "PY",
	"PER": "PE",
	"PHL": "PH",
	"PCN": "PN",
	"POL": "PL",
	"PRT": "PT",
	"PRI": "PR",
	"QAT": "QA",
	"REU": "RE",
	"ROU": "RO",
	"RUS": "RU",
	"RWA": "RW",
	"BLM": "BL",
	"SHN": "SH",
	"KNA": "KN",
	"LCA": "LC",
	"MAF": "MF",
	"SPM": "PM",
	"VCT": "VC",
	"WSM": "WS",
	"SMR": "SM",
	"STP": "ST",
	"SAU": "SA",
	"SEN": "SN",
	"SRB": "RS",
	"SYC": "SC",
	"SLE": "SL",
	"SGP": "SG",
	"SVK": "SK",
	"SVN": "SI",
	"SLB": "SB",
	"SOM": "SO",
	"ZAF": "ZA",
	"SGS": "GS",
	"SSD": "SS",
	"ESP": "ES",
	"LKA": "LK",
	"SDN": "SD",
	"SUR": "SR",
	"SJM": "SJ",
	"SWZ": "SZ",
	"SXM": "SX",
	"SWE": "SE",
	"CHE": "CH",
	"SYR": "SY",
	"TWN": "TW",
	"TJK": "TJ",
	"TZA": "TZ",
	"THA": "TH",
	"TLS": "TL",
	"TGO": "TG",
	"TKL": "TK",
	"TON": "TO",
	"TTO": "TT",
	"TUN": "TN",
	"TUR": "TR",
	"TKM": "TM",
	"TCA": "TC",
	"TUV": "TV",
	"UGA": "UG",
	"UKR": "UA",
	"ARE": "AE",
	"GBR": "GB",
	"USA": "US",
	"UMI": "UM",
	"URY": "UY",
	"UZB": "UZ",
	"VUT": "VU",
	"VEN": "VE",
	"VNM": "VN",
	"VIR": "VI",
	"WLF": "WF",
	"ESH": "EH",
	"YEM": "YE",
	"ZMB": "ZM",
	"ZWE": "ZW",
}

func GetFilteredBidders(rCtx models.RequestCtx, bidRequest *openrtb2.BidRequest, cache cache.Cache, partnerConfigMap map[int]map[string]string) (map[string]bool, bool) {
	filteredBidders := map[string]bool{}
	biddingConditionPerBidder := cache.GetBidderFilterConditions(rCtx)
	if len(biddingConditionPerBidder) == 0 {
		return filteredBidders, false
	}

	data := generateEvaluationData(bidRequest)
	allPartnersDroppedFlag := true
	for _, partnerConfig := range partnerConfigMap {
		if partnerConfig[models.SERVER_SIDE_FLAG] != "1" {
			continue
		}

		biddingCondition, ok := biddingConditionPerBidder[partnerConfig[models.BidderCode]]
		if !ok || evaluateBiddingCondition(data, biddingCondition) {
			allPartnersDroppedFlag = false
			filteredBidders[partnerConfig[models.BidderCode]] = true
		}
	}

	return filteredBidders, allPartnersDroppedFlag
}

func generateEvaluationData(BidRequest *openrtb2.BidRequest) *bytes.Reader {
	jsonStr := bytes.Buffer{}
	jsonStr.WriteByte('{')
	country := getCountryFromRequest(BidRequest)
	fmt.Fprintf(&jsonStr, `"%s":"%v"`, "country", country)
	jsonStr.WriteByte('}')
	return bytes.NewReader(jsonStr.Bytes())
}

func getCountryFromRequest(bidRequest *openrtb2.BidRequest) string {
	if bidRequest.Device != nil && bidRequest.Device.Geo != nil && bidRequest.Device.Geo.Country != "" {
		return countryAlpha3ToAlpha2Code[bidRequest.Device.Geo.Country]
	}

	if bidRequest.User != nil && bidRequest.User.Geo != nil && bidRequest.User.Geo.Country != "" {
		return countryAlpha3ToAlpha2Code[bidRequest.User.Geo.Country]
	}

	// return country using  netacuity
	return ""
}

func evaluateBiddingCondition(data, logic *bytes.Reader) bool {
	var result bytes.Buffer
	err := jsonlogic.Apply(logic, data, &result)
	if err != nil {
		fmt.Errorf("Error evaluating bidding-conditions | Error: %v for Logic %v and Data: %v", err, logic, data)
		return false
	}
	return strings.TrimRight(result.String(), "\r\n") == "true"
}
