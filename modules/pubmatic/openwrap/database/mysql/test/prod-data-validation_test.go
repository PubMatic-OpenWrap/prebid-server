package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prebid/prebid-server/modules/pubmatic/openwrap/models/adunitconfig"
)

func TestMain(t *testing.T) {
	data := `{"configPattern":"_AU_","config":{"/21682623839/TrueID_TV_Video/Live_TV/True-Premier-Football-HD1":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/NewTV":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/AmarinTV-HD":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/9MCOT-HD":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/CH8":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/NationTV":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/CH5":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/TNN16":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/CH3-HD":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"default":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/CH7-HD":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/VOD":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/ThairathTV-HD":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}},"/21682623839/TrueID_TV_Video/Live_TV/True4U":{"video":{"config":{"context":"instream"},"enabled":true},"device":{"deviceType":3}}}}`

	a := adunitconfig.AdUnitConfig{}
	err := json.Unmarshal([]byte(data), &a)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", a)
}
