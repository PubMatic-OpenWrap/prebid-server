package openwrap

import (
	"net/url"
	"strconv"
	"strings"

	"git.pubmatic.com/PubMatic/go-common/logger"
	"github.com/buger/jsonparser"
	"github.com/mileusna/useragent"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
)

// AMP predefined fluid sizes
var fluidSizes = []openrtb2.Format{
	{W: 250, H: 250},
	{W: 300, H: 50},
	{W: 300, H: 100},
	{W: 300, H: 200},
	{W: 300, H: 250},
	{W: 300, H: 600},
	{W: 320, H: 50},
	{W: 320, H: 75},
	{W: 320, H: 100},
	{W: 320, H: 180},
	{W: 320, H: 250},
	{W: 320, H: 280},
	{W: 320, H: 480},
	{W: 336, H: 280},
}

func (m *OpenWrap) applyAmpChanges(rctx models.RequestCtx, bidRequest *openrtb2.BidRequest) {
	if rctx.AmpParams.ImpID == "" {
		return
	}
	bidRequest.Imp[0].ID = rctx.AmpParams.ImpID
	bidRequest.Imp[0].BidFloor = rctx.AmpParams.BidFloor
	bidRequest.Imp[0].BidFloorCur = rctx.AmpParams.BidFloorCur
	bidRequest.Imp[0].Banner.Format = getFormatSizes(rctx.AmpParams.Width, rctx.AmpParams.Height, rctx.AmpParams.Multisize, fluidSizes)

	if bidRequest.Device != nil {
		if os := useragent.Parse(rctx.UA); os.OS != "" {
			bidRequest.Device.OS = os.OS
		}
	}

	if bidRequest.Site == nil {
		bidRequest.Site = &openrtb2.Site{}
	}

	if bidRequest.Site.Publisher == nil {
		bidRequest.Site.Publisher = &openrtb2.Publisher{}
	}

	if bidRequest.Site.Publisher.ID == "" {
		bidRequest.Site.Publisher.ID = rctx.PubIDStr
	}

	if bidRequest.Site.Page == "" {
		if rctx.AmpParams.Curl != "" {
			bidRequest.Site.Page = rctx.AmpParams.Curl
		} else if rctx.AmpParams.Origin != "" {
			bidRequest.Site.Page = rctx.AmpParams.Origin
		} else if rctx.AmpParams.Purl != "" {
			bidRequest.Site.Page = rctx.AmpParams.Purl
		}
	}

	bidRequest.Test = getAmpTestParamter(rctx.AmpParams.Purl)

	if rctx.SSAuction == 0 {
		bidRequest.Ext, _ = jsonparser.Set(bidRequest.Ext, []byte("true"), "prebid", "returnallbidstatus")
	}

	// jsonparser.Set(bidRequest.Ext, []byte(`{"wrapper":{"profileid":`+rctx.ProfileIDStr+`,"versionid"}}`), "prebid", "bidderparams", "pubmatic")
}

func getFormatSizes(w, h, multisize string, fluidSizes []openrtb2.Format) []openrtb2.Format {

	var errw, errh error
	var width, height int64

	if w != models.FluidStr {
		width, errw = strconv.ParseInt(w, 10, 64)
	}

	if h != models.FluidStr {
		height, errw = strconv.ParseInt(h, 10, 64)
	}

	//if both w and h are blank or any of w or h is invalid(means not integer or "fluid")
	// or both w and h are fluid
	// then consider ms if present else consider fluid sizes
	if nil != errw || nil != errh || (w == "" && h == "") || (w == models.FluidStr && h == models.FluidStr) {
		if formatArr := convertMultisizeToFormatArr(multisize); len(formatArr) > 0 {
			return formatArr
		}
		return fluidSizes
	}

	//if w and h are integers, then use w,h and ms for format
	// add w and h to format Arr
	if w != models.FluidStr && h != models.FluidStr {
		formatArr := []openrtb2.Format{
			{
				W: width,
				H: height,
			},
		}
		//add ms to format arr
		formatArr = append(formatArr, convertMultisizeToFormatArr(multisize)...)
		return formatArr
	}

	//if only h is fluid
	if h == models.FluidStr {
		if formatArr := convertMultisizeToFormatArr(multisize); len(formatArr) > 0 {
			return formatArr
		}
		return findApplicableSizesFromFluidSizes(fluidSizes, width, true)

	}

	//if only w is fluid
	if w == models.FluidStr {
		if formatArr := convertMultisizeToFormatArr(multisize); len(formatArr) > 0 {
			return formatArr
		}
		return findApplicableSizesFromFluidSizes(fluidSizes, height, false)
	}
	return nil
}

func convertMultisizeToFormatArr(multisize string) []openrtb2.Format {
	formatArr := make([]openrtb2.Format, 0)
	if multisize != "" {
		sizeStrings := strings.Split(multisize, ",")

		for _, sizeString := range sizeStrings {

			wh := strings.Split(sizeString, "x")
			if len(wh) != 2 {
				logger.Error("Invalid value for field multisizes 'ms'. Must be in format (w1xh1,w2xh2)")
				continue
			}

			w, err := strconv.ParseInt(wh[0], 10, 64)
			if err != nil {
				logger.Error("Invalid integer passed in multisizes 'ms'. Must be in format (w1xh1,w2xh2)")
				continue
			}

			h, err := strconv.ParseInt(wh[1], 10, 64)
			if err != nil {
				logger.Error("Invalid integer passed in multisizes 'ms'. Must be in format (w1xh1,w2xh2)")
				continue
			}

			formatArr = append(formatArr, openrtb2.Format{
				W: w,
				H: h,
			})
		}
		return formatArr
	}
	return formatArr
}

func findApplicableSizesFromFluidSizes(fluidSizes []openrtb2.Format, val int64, isFixedWidth bool) []openrtb2.Format {
	formatArr := make([]openrtb2.Format, 0)

	for _, format := range fluidSizes {
		if (isFixedWidth && format.W <= val) || (!isFixedWidth && format.H <= val) {
			formatArr = append(formatArr, format)
		}
	}
	return formatArr
}

func getAmpTestParamter(purl string) int8 {
	pubblisherUrl, err := url.Parse(purl)
	if err != nil {
		return 0
	}

	test, _ := strconv.ParseBool(pubblisherUrl.Query().Get(models.PubmaticTest))
	if test {
		return 1
	}

	return 0
}
