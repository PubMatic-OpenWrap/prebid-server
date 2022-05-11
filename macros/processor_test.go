package macros

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringBasedProcessor(t *testing.T) {

	p, _ := NewProcessor(STRING_BASED, Config{
		delimiter:   "##",
		macroValues: testData,
	})
	tURL := "http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##"
	expected := "http://tracker.com?macro_1=vast&macro_2=consent&custom=1234&custom=##shri##"
	actual, err := p.Replace(tURL)
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, expected, actual, fmt.Sprintf("Expected [%s] found - %s", expected, actual))
}

func TestTemplateBasedProcessor(t *testing.T) {
	tURL := "http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##"
	p, _ := NewProcessor(TEMPLATE_BASED, Config{
		delimiter:   "##",
		macroValues: testData,
		templates:   []string{tURL},
	})
	// expect ##shri## is replaced with empty
	expected := "http://tracker.com?macro_1=vast&macro_2=consent&custom=1234&custom="
	actual, err := p.Replace(tURL)
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, expected, actual, fmt.Sprintf("Expected [%s] found - %s", expected, actual))
	fmt.Println(actual)
}

func BenchmarkStringBasedProcessor(b *testing.B) {
	for n := 0; n < b.N; n++ {
		p, _ := NewProcessor(STRING_BASED, Config{
			delimiter:   "##",
			macroValues: testData,
		})
		tURL := "http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##"
		p.Replace(tURL)
	}
}

func BenchmarkTemplateBasedProcessor(b *testing.B) {
	tURL := "http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##"
	p, _ := NewProcessor(TEMPLATE_BASED, Config{
		delimiter:   "##",
		macroValues: testData,
		templates:   []string{tURL},
	})
	for n := 0; n < b.N; n++ {
		p.Replace(tURL)
	}
}

var testData = map[string]string{
	"PBS_EVENTTYPE":       "vast",
	"PBS_VASTEVENT":       "vclick",
	"PBS_APPBUNDLE":       "com.my.app",
	"PBS_DOMAIN":          "mydomain.com",
	"PBS_PUBDOMAIN":       "pub.domain.com",
	"PBS_PAGEURL":         "http://mydomain.com/myPage?key=value",
	"PBS_GDPRCONSENT":     "consent",
	"PBS_LIMITADTRACKING": " yes",
	"PBS_VASTCRTID":       "vast_creative_1",
	"PBS_BIDID":           "bid_123",
	"PBS_AUCTIONID":       "auction_123",
	"PBS_ACCOUNTID":       "5890",
	"PBS_TIMESTAMP":       "12345678",
	"PBS_BIDDER":          "pubmatic",
	"PBS_INTEGRATION":     "video",
	"PBS_LINEID":          "line_item_1",
	"PBS_CHANNEL":         "header_bidding",
	"PBS_ANALYTICS":       "abc_adaptor",
	"PBS_MACRO_profileid": "1234",
}
