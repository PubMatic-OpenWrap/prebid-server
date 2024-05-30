package adbutler_onsite

import (
	"fmt"
	"text/template"

	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/config"
	"github.com/prebid/prebid-server/macros"
	"github.com/prebid/prebid-server/openrtb_ext"
)

type AdButlerOnsiteAdapter struct {
	endpoint *template.Template
}

// Builder builds a new instance of the AdButler onsite adapter for the given bidder with the given config.
func Builder(bidderName openrtb_ext.BidderName, config config.Adapter) (adapters.Bidder, error) {
	
	endpointtemplate, err := template.New("endpointTemplate").Parse(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse endpoint url template: %v", err)
	}

	bidder := &AdButlerOnsiteAdapter{
		endpoint: endpointtemplate,
	}
	return bidder, nil
}

func (a *AdButlerOnsiteAdapter) buildEndpointURL(accountID, zoneID string) (string, error) {
	endpointParams := macros.EndpointTemplateParams{
		AccountID: accountID,
		ZoneID:    zoneID,
	}
	return macros.ResolveMacros(a.endpoint, endpointParams)
}
