package models

type RequestCtx struct {
	UA                          string
	VastUnwrapEnabled           bool
	PubID, ProfileID, DisplayID int
	Endpoint                    string
	VastUnwrapStatsEnabled      bool
	Redirect                    bool
}
