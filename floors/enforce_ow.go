package floors

import "github.com/prebid/prebid-server/v2/currency"

func getOriginalBidCpmUsd(price float64, from string, conversions currency.Conversions) float64 {
	rate, _ := getCurrencyConversionRate(from, "USD", conversions)
	return rate * price
}
