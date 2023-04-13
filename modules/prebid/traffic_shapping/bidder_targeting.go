package trafficshapping

// Bidder rules : Profile version level
// allow only whitelisted keywords

func getBidderTargeting(rtbRequest RTBRequest) map[string]Expression {
	// consider appnexus if series = friends and country = india
	seriesExp := Eq{
		Key:   "series",
		Value: "friends",
	}
	countryExp := Eq{
		Key:   "country",
		Value: "India",
	}
	appNexusRule := And{
		Left:  seriesExp,
		Right: countryExp,
	}

	// consider freewheelSsp if series = saregampa
	fSeriesExp := Eq{
		Key:   "series",
		Value: "saregampa",
	}
	freewheelSspRule := fSeriesExp

	// consider unruly if device.IFA field is set and user.gender = male
	// rtbReq := RTBRequest{
	// 	request: &openrtb2.BidRequest{},
	// }
	deviceIFARule := rtbRequest.IsPresent("device.IFA")
	userGenderRule := Eq{
		Key:   "user.gender",
		Value: "male",
	}
	unrulyRule := And{
		Left:  deviceIFARule,
		Right: userGenderRule,
	}

	return map[string]Expression{
		"appnexus":     appNexusRule,
		"freewheelssp": freewheelSspRule,
		"unruly":       unrulyRule,
	}
}

// func GetRTBFieldTargeting() {

// 	// target unruly if device.IFA field is set and user.gender = male
// 	rtbReqTargeting := RTBRequestTargeting{
// 		request: &openrtb2.BidRequest{},
// 	}
// 	deviceIFARule := rtbReqTargeting.IsPresent("device.IFA")
// 	userGenderRule := Eq{
// 		Key:   "user.gender",
// 		Value: "male",
// 	}
// 	unrulyRule := And{
// 		Left:  deviceIFARule,
// 		Right: userGenderRule,
// 	}
// 	unrulyRule.Evaluate(map[string]string{})
// }
