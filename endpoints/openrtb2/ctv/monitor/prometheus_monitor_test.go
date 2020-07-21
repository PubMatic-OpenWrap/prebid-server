package monitor

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv"
	"github.com/PubMatic-OpenWrap/prebid-server/endpoints/openrtb2/ctv/impressions"
	"github.com/PubMatic-OpenWrap/prebid-server/openrtb_ext"
	"github.com/prometheus/client_golang/prometheus"
)

func TestMeasureExecutionTime(t *testing.T) {
	monitor := newPrometheusMonitor("algo1")
	monitor.MeasureExecutionTime(time.Now())
}

func TestUseMeasureExecutionTime(t *testing.T) {
	monitor := newPrometheusMonitor("algo1")
	start := time.Now()
	defer monitor.MeasureExecutionTime(start)

	sum := 18 + 45
	fmt.Println(sum)
}

func TestWithPrometheus(t *testing.T) {

	http.Handle("/metrics", prometheus.Handler())

	algorithm := impressions.MaximizeForDuration
	monitor := newPrometheusMonitor("a1_max")
	combMonitor := newPrometheusMonitor("a1_combgen")

	http.HandleFunc("/testcomb", func(w http.ResponseWriter, r *http.Request) {
		// major time for combination generator
			monitor = combMonitor
		buckets := ctv.BidsBuckets{}
		buckets[5] = make([]*ctv.Bid, 2)  // 2 ads of 5 seconds
		buckets[10] = make([]*ctv.Bid, 4) // 4 ads of 10 seconds
		buckets[15] = make([]*ctv.Bid, 3) // 3 ads of 15 seconds

		maxAds := new(int)
		*maxAds = 10
		minAds := new(int)
		*minAds = rand.Intn(*maxAds)

		start := time.Now()
		defer monitor.MeasureExecutionTime(start)

		comb := ctv.NewCombination(buckets, 10, 90, &openrtb_ext.VideoAdPod{
			MinAds: minAds,
			MaxAds: maxAds,
		})

		comnCnt := 0
		for len(comb.Get()) > 0 {
			comnCnt++
		}

		result := strconv.Itoa(comnCnt) + " combinations"

		monitor.Scenario(result)
		w.Write([]byte(result))
	})

	http.HandleFunc("/testalgo", func(w http.ResponseWriter, r *http.Request) {

		// major time for impression generation algo 2
		start := time.Now()
		defer monitor.MeasureExecutionTime(start)

		maxDur := new(int)
		*maxDur = 90
		dur := new(int)
		*dur = rand.Intn(*maxDur)
		ads := new(int)
		*ads = rand.Intn(6)
		maxads := new(int)
		*maxads = 15
		imp := impressions.NewImpressions(5, 90, &openrtb_ext.VideoAdPod{
			MinDuration: dur,
			MaxDuration: maxDur,
			MinAds:      ads,
			MaxAds:      ads,
		}, algorithm)
		imps := imp.Get()

		result := strconv.Itoa(len(imps)) + " Impressions"
		monitor.Scenario(result)
		w.Write([]byte(result))
	})

	port := "8080"
	if val, ok := os.LookupEnv("port"); ok && len(val) > 0 {
		port = val
	}
	server := http.Server{
		Addr:           fmt.Sprintf(":%s", port),
		ReadTimeout:    3 * time.Second,
		WriteTimeout:   3 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server.ListenAndServe()

}
