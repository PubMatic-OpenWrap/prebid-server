package errortypes

// Timeout should be used to flag that a bidder failed to return a response because the PBS timeout timer
// expired before a result was received.
//
// Timeouts will not be written to the app log, since it's not an actionable item for the Prebid Server hosts.
type Timeout struct {
	Message string
}

func (err *Timeout) Error() string {
	return err.Message
}

func (err *Timeout) Code() int {
	return TimeoutErrorCode
}

func (err *Timeout) Severity() Severity {
	return SeverityFatal
}

// BadInput should be used when returning errors which are caused by bad input.
// It should _not_ be used if the error is a server-side issue (e.g. failed to send the external request).
//
// BadInputs will not be written to the app log, since it's not an actionable item for the Prebid Server hosts.
type BadInput struct {
	Message string
}

func (err *BadInput) Error() string {
	return err.Message
}

func (err *BadInput) Code() int {
	return BadInputErrorCode
}

func (err *BadInput) Severity() Severity {
	return SeverityFatal
}

// BlacklistedApp should be used when a request App.ID matches an entry in the BlacklistedApps
// environment variable array
//
// These errors will be written to  http.ResponseWriter before canceling execution
type BlacklistedApp struct {
	Message string
}

func (err *BlacklistedApp) Error() string {
	return err.Message
}

func (err *BlacklistedApp) Code() int {
	return BlacklistedAppErrorCode
}

func (err *BlacklistedApp) Severity() Severity {
	return SeverityFatal
}

// BlacklistedAcct should be used when a request account ID matches an entry in the BlacklistedAccts
// environment variable array
//
// These errors will be written to  http.ResponseWriter before canceling execution
type BlacklistedAcct struct {
	Message string
}

func (err *BlacklistedAcct) Error() string {
	return err.Message
}

func (err *BlacklistedAcct) Code() int {
	return BlacklistedAcctErrorCode
}

func (err *BlacklistedAcct) Severity() Severity {
	return SeverityFatal
}

// AcctRequired should be used when the environment variable ACCOUNT_REQUIRED has been set to not
// process requests that don't come with a valid account ID
//
// These errors will be written to  http.ResponseWriter before canceling execution
type AcctRequired struct {
	Message string
}

func (err *AcctRequired) Error() string {
	return err.Message
}

func (err *AcctRequired) Code() int {
	return AcctRequiredErrorCode
}

func (err *AcctRequired) Severity() Severity {
	return SeverityFatal
}

// BadServerResponse should be used when returning errors which are caused by bad/unexpected behavior on the remote server.
//
// For example:
//
//   - The external server responded with a 500
//   - The external server gave a malformed or unexpected response.
//
// These should not be used to log _connection_ errors (e.g. "couldn't find host"),
// which may indicate config issues for the PBS host company
type BadServerResponse struct {
	Message string
}

func (err *BadServerResponse) Error() string {
	return err.Message
}

func (err *BadServerResponse) Code() int {
	return BadServerResponseErrorCode
}

func (err *BadServerResponse) Severity() Severity {
	return SeverityFatal
}

// FailedToRequestBids is an error to cover the case where an adapter failed to generate any http requests to get bids,
// but did not generate any error messages. This should not happen in practice and will signal that an adapter is poorly
// coded. If there was something wrong with a request such that an adapter could not generate a bid, then it should
// generate an error explaining the deficiency. Otherwise it will be extremely difficult to debug the reason why an
// adapter is not bidding.
type FailedToRequestBids struct {
	Message string
}

func (err *FailedToRequestBids) Error() string {
	return err.Message
}

func (err *FailedToRequestBids) Code() int {
	return FailedToRequestBidsErrorCode
}

func (err *FailedToRequestBids) Severity() Severity {
	return SeverityFatal
}

// AdpodPrefiltering should be used when ctv impression algorithm not able to generate impressions
type AdpodPrefiltering struct {
	Message string
}

func (err *AdpodPrefiltering) Error() string {
	return err.Message
}

func (err *AdpodPrefiltering) Code() int {
	return AdpodPrefilteringErrorCode
}

func (err *AdpodPrefiltering) Severity() Severity {
	return SeverityFatal
}

// BidderTemporarilyDisabled is used at the request validation step, where we want to continue processing as best we
// can rather than returning a 4xx, and still return an error message.
// The initial usecase is to flag deprecated bidders.
type BidderTemporarilyDisabled struct {
	Message string
}

func (err *BidderTemporarilyDisabled) Error() string {
	return err.Message
}

func (err *BidderTemporarilyDisabled) Code() int {
	return BidderTemporarilyDisabledErrorCode
}

func (err *BidderTemporarilyDisabled) Severity() Severity {
	return SeverityWarning
}

// Warning is a generic non-fatal error.
type Warning struct {
	Message     string
	WarningCode int
}

func (err *Warning) Error() string {
	return err.Message
}

func (err *Warning) Code() int {
	return err.WarningCode
}

func (err *Warning) Severity() Severity {
	return SeverityWarning
}

// BidderFailedSchemaValidation is used at the request validation step,
// when the bidder parameters fail the schema validation, we want to
// continue processing the request and still return an error message.
type BidderFailedSchemaValidation struct {
	Message string
}

func (err *BidderFailedSchemaValidation) Error() string {
	return err.Message
}

func (err *BidderFailedSchemaValidation) Code() int {
	return BidderFailedSchemaValidationErrorCode
}

func (err *BidderFailedSchemaValidation) Severity() Severity {
	return SeverityWarning
}

// NoBidPrice should be used when vast response doesn't contain any price value
type NoBidPrice struct {
	Message string
}

func (err *NoBidPrice) Error() string {
	return err.Message
}

func (err *NoBidPrice) Code() int {
	return NoBidPriceErrorCode
}

func (err *NoBidPrice) Severity() Severity {
	return SeverityWarning
}

// NoValidBid should be used when responded bids doesn't contain mandatory fields
type NoValidBid struct {
	Message string
}

func (err *NoValidBid) Error() string {
	return err.Message
}

func (err *NoValidBid) Code() int {
	return NoValidBidErrorCode
}

func (err *NoValidBid) Severity() Severity {
	return SeverityWarning
}

// InvalidSource should be used when responded with invalid source error code
type InvalidSource struct {
	Message string
}

func (err *InvalidSource) Error() string {
	return err.Message
}

func (err *InvalidSource) Code() int {
	return InvalidSourceErrorCode
}

func (err *InvalidSource) Severity() Severity {
	return SeverityWarning
}

// InvalidCatalog should be used when responded with invalid catalog error code
type InvalidCatalog struct {
	Message string
}

func (err *InvalidCatalog) Error() string {
	return err.Message
}

func (err *InvalidCatalog) Code() int {
	return InvalidCatalogErrorCode
}

func (err *InvalidCatalog) Severity() Severity {
	return SeverityWarning
}

// UnknownError should be used when responded with unknown error code
type UnknownError struct {
	Message string
}

func (err *UnknownError) Error() string {
	return err.Message
}

func (err *UnknownError) Code() int {
	return UnknownErrorCode
}

func (err *UnknownError) Severity() Severity {
	return SeverityWarning
}

// AdpodPostFiltering should be used when vast response doesn't contain any price value
type AdpodPostFiltering struct {
	Message string
}

func (err *AdpodPostFiltering) Error() string {
	return err.Message
}

func (err *AdpodPostFiltering) Code() int {
	return AdpodPostFilteringWarningCode
}

func (err *AdpodPostFiltering) Severity() Severity {
	return SeverityWarning
}
