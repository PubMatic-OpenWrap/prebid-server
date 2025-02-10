package errortypes

// Defines numeric codes for well-known errors.
const (
	UnknownErrorCode = 999
	TimeoutErrorCode = iota
	BadInputErrorCode
	BlockedAppErrorCode
	BadServerResponseErrorCode
	FailedToRequestBidsErrorCode
	BidderTemporarilyDisabledErrorCode
	AccountDisabledErrorCode
	AcctRequiredErrorCode
	NoConversionRateErrorCode
	MalformedAcctErrorCode
	ModuleRejectionErrorCode
	TmaxTimeoutErrorCode
	FailedToMarshalErrorCode
	FailedToUnmarshalErrorCode
	InvalidImpFirstPartyDataErrorCode

	// NYC: shall we have different range for OW error codes to avoid change in codes with introduction of new PBS error codes.
	NoBidPriceErrorCode
	BidderFailedSchemaValidationErrorCode
	AdpodPrefilteringErrorCode
	BidRejectionFloorsErrorCode
)

// Defines numeric codes for well-known warnings.
const (
	UnknownWarningCode               = 10999
	InvalidPrivacyConsentWarningCode = iota + 10000
	AccountLevelDebugDisabledWarningCode
	BidderLevelDebugDisabledWarningCode
	DisabledCurrencyConversionWarningCode
	AlternateBidderCodeWarningCode
	MultiBidWarningCode
	AdServerTargetingWarningCode
	BidAdjustmentWarningCode
	FloorBidRejectionWarningCode
	InvalidBidResponseDSAWarningCode
	SecCookieDeprecationLenWarningCode
	SecBrowsingTopicsWarningCode
	AdpodPostFilteringWarningCode
	InvalidVastVersionWarningCode
)

// Coder provides an error or warning code with severity.
type Coder interface {
	Code() int
	Severity() Severity
}

// ReadCode returns the error or warning code, or UnknownErrorCode if unavailable.
func ReadCode(err error) int {
	if e, ok := err.(Coder); ok {
		return e.Code()
	}
	return UnknownErrorCode
}
