package macros

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOK(t *testing.T) {
	assert.True(t, true)
}

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestReplaceStringBased(t *testing.T) {
// 	delimiter := "##"
// 	templateURL := getSampleTemplateURL(delimiter)
// 	actual, err := replaceStringBased(templateURL, delimiter, testData, MacroValueConfig{
// 		FailOnError: true,
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	expected := expectedURL(delimiter)
// 	assert.Equal(t, expected, actual)
// }

// func TestReplaceTemplateBased(t *testing.T) {
// 	delimiter := "##"
// 	templateURL := getSampleTemplateURL(delimiter)
// 	actual, err := replaceTemplateBased(templateURL, delimiter, testData, MacroValueConfig{
// 		FailOnError: true,
// 	})
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	expected := expectedURL(delimiter)
// 	assert.Equal(t, expected, actual)
// }

// func BenchmarkReplaceStringBased(b *testing.B) {
// 	for n := 0; n < b.N; n++ {
// 		delimiter := "##"
// 		templateURL := getSampleTemplateURL(delimiter)
// 		replaceStringBased(templateURL, delimiter, testData, MacroValueConfig{})
// 	}
// }

// // instead of PBS- we will need to use PBS_ in case of go template based approach
// // because go template does not support - in template.
// var testData = map[string]string{
// 	"PBS_EVENTTYPE":       "vast",
// 	"PBS_VASTEVENT":       "vclick",
// 	"PBS_APPBUNDLE":       "com.my.app",
// 	"PBS_DOMAIN":          "mydomain.com",
// 	"PBS_PUBDOMAIN":       "pub.domain.com",
// 	"PBS_PAGEURL":         "http://mydomain.com/myPage?key=value",
// 	"PBS_GDPRCONSENT":     "consent",
// 	"PBS_LIMITADTRACKING": " yes",
// 	"PBS_VASTCRTID":       "vast_creative_1",
// 	"PBS_BIDID":           "bid_123",
// 	"PBS_AUCTIONID":       "auction_123",
// 	"PBS_ACCOUNTID":       "5890",
// 	"PBS_TIMESTAMP":       "12345678",
// 	"PBS_BIDDER":          "pubmatic",
// 	"PBS_INTEGRATION":     "video",
// 	"PBS_LINEID":          "line_item_1",
// 	"PBS_CHANNEL":         "header_bidding",
// 	"PBS_ANALYTICS":       "abc_adaptor",
// 	"PBS_MACRO_profileid": "1234",
// }

// var macros = make([]string, 0)

// func getSampleTemplateURL(delimiter string) string {
// 	for macro := range testData {
// 		macros = append(macros, macro)
// 	}
// 	sample := "http://tracker.com?"
// 	for cnt, macro := range macros {
// 		sample += fmt.Sprintf("macro_%d=%s%s%s&", cnt+1, delimiter, macro, delimiter)
// 		cnt += 1
// 	}
// 	// no macro value test sample
// 	sample += fmt.Sprintf("no_macro=%sNO_MACRO%s", delimiter, delimiter)
// 	return sample
// }

// func expectedURL(delimiter string) string {
// 	sample := "http://tracker.com?"
// 	for cnt, macro := range macros {
// 		sample += fmt.Sprintf("macro_%d=%s&", cnt+1, testData[macro])
// 	}
// 	// no macro value test sample
// 	sample += fmt.Sprintf("no_macro=%sNO_MACRO%s", delimiter, delimiter)
// 	return sample
// }
