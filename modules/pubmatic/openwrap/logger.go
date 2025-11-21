package openwrap

import (
	"fmt"

	"github.com/prebid/openrtb/v20/openrtb2"
)

func updateIncomingSlotsWithFormat(incomingSlots []string, format []openrtb2.Format) []string {
	if len(format) == 0 {
		return incomingSlots
	}

	sizes := make(map[string]struct{})
	for _, size := range incomingSlots {
		sizes[size] = struct{}{}
	}

	// Add new sizes from format
	for _, f := range format {
		sizes[fmt.Sprintf("%dx%d", f.W, f.H)] = struct{}{}
	}

	updatedSlots := make([]string, 0, len(sizes))
	for k := range sizes {
		updatedSlots = append(updatedSlots, k)
	}
	return updatedSlots
}
