package wakanda

type Wakanda struct {
	SFTP                  SFTP
	HostName              string
	DCName                string
	PodName               string
	MaxDurationInMin      int
	CleanupFrequencyInMin int
}
type SFTP struct {
	User        string
	Password    string
	ServerIP    string
	Destination string
}
