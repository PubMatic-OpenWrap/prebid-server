package macros

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringBasedProcessor(t *testing.T) {

	p, _ := NewProcessor(STRING_BASED, Config{
		Delimiter: "##",
	})
	tURL := "http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##"
	expected := "http://tracker.com?macro_1=vast&macro_2=consent&custom=1234&custom=##shri##"
	actual, err := p.Replace(tURL, testData)
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, expected, actual, fmt.Sprintf("Expected [%s] found - %s", expected, actual))
}

func TestTemplateBasedProcessor(t *testing.T) {
	tURL := "http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##"
	p, _ := NewProcessor(TEMPLATE_BASED, Config{
		Delimiter: "##",
		Templates: []string{tURL},
	})
	// expect ##shri## is replaced with empty
	expected := "http://tracker.com?macro_1=vast&macro_2=consent&custom=1234&custom="
	actual, err := p.Replace(tURL, testData)
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, expected, actual, fmt.Sprintf("Expected [%s] found - %s", expected, actual))
	fmt.Println(actual)
}

func TestStringCachedIndexBasedProcessor(t *testing.T) {
	Delimiter := "##"
	// tURL := fmt.Sprintf("http://tracker.com?macro_1=%sPBS_EVENTTYPE&%smacro_2=%sPBS_GDPRCONSENT%s&custom=%sPBS_MACRO_profileid%s&custom=%sshri%s", Delimiter, Delimiter, Delimiter, Delimiter, Delimiter, Delimiter, Delimiter, Delimiter)
	tURL, expected := buildLongInputURL(1000, Delimiter)
	// println(expected)
	p, _ := NewProcessor(STRING_INDEX_CACHED, Config{
		Delimiter: Delimiter,
		Templates: []string{tURL},
	})
	// expect ##shri## is replaced with empty
	// expected := "http://tracker.com?macro_1=vast&macro_2=consent&custom=1234&custom=&url=http://mydomain.com/myPage?key=value"
	actual, err := p.Replace(tURL, testData)
	// println(actual)
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, expected, actual, fmt.Sprintf("Expected [%s] found - %s", expected, actual))
	fmt.Println("\n" + actual)

}

func BenchmarkStringBasedProcessor(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stringBasedProcessor.Replace(tURL, testData)
	}
}

//const tURL = "http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##&url=##PBS_PAGEURL##"

var tmplProcessor IProcessor
var stringBasedProcessor IProcessor
var tmplProcessorAlwaysInit IProcessor
var stringIndexBasedMacroProcessor IProcessor
var stringCachedIndexBasedProcessor IProcessor
var tURL, URL2, URL3, URL4 string

func init() {

	tURL = buildLongInputURL0(100, "##")

	URL2 = buildLongInputURL0(5000, "##")
	URL3 = buildLongInputURL0(10000, "##")
	URL4 = buildLongInputURL0(15000, "##")

	//fmt.Println(tURL)
	tmplProcessor, _ = NewProcessor(TEMPLATE_BASED, Config{
		Delimiter: "##",
		Templates: []string{tURL, URL2, URL3, URL4},
	})
	stringBasedProcessor, _ = NewProcessor(STRING_BASED, Config{
		Delimiter: "##",
	})

	tmplProcessorAlwaysInit, _ = NewProcessor(TEMPLATE_BASED_INIT_ALWAYS, Config{
		Delimiter: "##",
	})

	stringIndexBasedMacroProcessor, _ = NewProcessor(VAST_BIDDER_MACRO_PROCESSOR, Config{
		Delimiter: "##",
	})

	stringCachedIndexBasedProcessor, _ = NewProcessor(STRING_INDEX_CACHED, Config{
		Delimiter:   "##",
		Templates:   []string{tURL, URL2, URL3, URL4},
		valueConfig: MacroValueConfig{
			// UrlEscape: true,
		},
	})
}

func BenchmarkTemplateBasedProcessor(b *testing.B) {
	for n := 0; n < b.N; n++ {
		tmplProcessor.Replace(tURL, testData)
	}
}

func BenchmarkTemplateBasedProcessorInitAlways(b *testing.B) {
	for n := 0; n < b.N; n++ {
		tmplProcessorAlwaysInit.Replace(tURL, testData)
	}
}

func BenchmarkStringIndexBasedMacroProcessor(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stringIndexBasedMacroProcessor.Replace(tURL, testData)
	}
}

func BenchmarkStringCachedIndexBasedProcessor(b *testing.B) {
	for n := 0; n < b.N; n++ {
		stringCachedIndexBasedProcessor.Replace(tURL, testData)
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

func getSampleTemplateURL(Delimiter string) string {
	macros := make([]string, 0)
	for macro := range testData {
		macros = append(macros, macro)
	}
	sample := "http://tracker.com?"
	for cnt, macro := range macros {
		sample += fmt.Sprintf("macro_%d=%s%s%s&", cnt+1, Delimiter, macro, Delimiter)
		cnt += 1
	}
	// no macro value test sample
	sample += fmt.Sprintf("no_macro=%sNO_MACRO%s", Delimiter, Delimiter)
	return sample
}

func TestStringIndexBasedMacroProcessor(t *testing.T) {
	p, _ := NewProcessor(VAST_BIDDER_MACRO_PROCESSOR, Config{
		Delimiter: "##",
	})
	tURL := "http://tracker.com?macro_1=##PBS_EVENTTYPE##&macro_2=##PBS_GDPRCONSENT##&custom=##PBS_MACRO_profileid##&custom=##shri##"
	expected := "http://tracker.com?macro_1=vast&macro_2=consent&custom=1234&custom=##shri##"
	actual, err := p.Replace(tURL, testData)
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Equal(t, expected, actual, fmt.Sprintf("Expected [%s] found - %s", expected, actual))
}

func buildLongInputURL0(noOfMacros int, Delimiter string) string {
	url, _ := buildLongInputURL(noOfMacros, Delimiter)
	return url
}
func buildLongInputURL(noOfMacros int, Delimiter string) (string, string) {
	url := ""
	cnt := 0
	expected := ""
	for cnt <= noOfMacros {
		for macro, value := range testData {
			url += fmt.Sprintf("key_%d=%s%s%s&", cnt, Delimiter, macro, Delimiter)
			expected += fmt.Sprintf("key_%d=%s&", cnt, value)
			cnt++
			if cnt > noOfMacros {
				break
			}
		}
	}
	return url, expected
}
