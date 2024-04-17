package publisherfeature

type Feature interface {
	IsFscApplicable(pubId int, seat string, dspId int) bool
	IsAmpMultiformatEnabled(pubId int) bool
	IsTBFFeatureEnabled(pubid int, profid int) bool
}