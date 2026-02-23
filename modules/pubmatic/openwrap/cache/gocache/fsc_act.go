package gocache

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
)

// GetFSCAndACTThresholdsPerDSP returns both FSC and ACT DSP thresholds in one DB call when the
// database supports it, to avoid duplicate round-trips.
func (c *cache) GetFSCAndACTThresholdsPerDSP() (fscMap map[int]int, actMap map[int]int, err error) {
	fscMap, actMap, err = c.db.GetFSCAndACTThresholdsPerDSP()
	if err != nil {
		c.metricEngine.RecordDBQueryFailure(models.AllDspFscAndActPcntQuery, "", "")
		glog.Errorf(models.ErrDBQueryFailed, models.AllDspFscAndActPcntQuery, "", "", err)
		return nil, nil, fmt.Errorf("[ErrorFscActDspUpdate]:%w", err)
	}
	return fscMap, actMap, nil
}
