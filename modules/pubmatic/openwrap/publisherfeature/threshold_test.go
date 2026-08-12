package publisherfeature

import "testing"

func TestPredictThresholdValue(t *testing.T) {
	// threshold 100 => always true (rand.Intn(100) < 100)
	if got := predictThresholdValue(100); got != true {
		t.Errorf("predictThresholdValue(100) = %v, want true", got)
	}
}

func TestIsUnderThreshold(t *testing.T) {
	tests := []struct {
		name               string
		disabledPublishers map[int]struct{}
		thresholdsPerDsp   map[int]int
		pubid, dspid       int
		want               int
	}{
		{
			name:               "publisher disabled => 0",
			disabledPublishers: map[int]struct{}{5890: {}},
			thresholdsPerDsp:   map[int]int{6: 100},
			pubid:              5890, dspid: 6,
			want: 0,
		},
		{
			name:               "enabled and threshold 100 => 1",
			disabledPublishers: map[int]struct{}{58903: {}},
			thresholdsPerDsp:   map[int]int{6: 100},
			pubid:              5890, dspid: 6,
			want: 1,
		},
		{
			name:               "dsp not in map => 0",
			disabledPublishers: map[int]struct{}{},
			thresholdsPerDsp:   map[int]int{6: 100},
			pubid:              58907, dspid: 90,
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnderThreshold(tt.disabledPublishers, tt.thresholdsPerDsp, tt.pubid, tt.dspid)
			if got != tt.want {
				t.Errorf("isUnderThreshold() = %v, want %v", got, tt.want)
			}
		})
	}
}
