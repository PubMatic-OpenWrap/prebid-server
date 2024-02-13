package models

type RequestCtx struct {
	UA                                     string
	VastUnwrapEnabled                      bool
	PubID, ProfileID, DisplayID, VersionID int
	Endpoint                               string
	VastUnwrapStatsEnabled                 bool
	Redirect                               bool
}
