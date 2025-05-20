package feature

import (
	"context"
	"time"
)

const (
	FeatureFlexSlot = "flexslot"
)

func (fl *FeatureLoader) LoadGoogleSDKFeatures() []Feature {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Millisecond*time.Duration(fl.cfg.MaxDbContextTimeout)))
	defer cancel()

	rows, err := fl.db.QueryContext(ctx, fl.cfg.Queries.GetBannerSizesQuery)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var slotSizes []string
	for rows.Next() {
		var slotSize string
		if err := rows.Scan(&slotSize); err != nil {
			return nil
		}
		slotSizes = append(slotSizes, slotSize)
	}
	if err := rows.Err(); err != nil {
		return nil
	}

	// Flexslot
	flexSlot := Feature{
		Name: FeatureFlexSlot,
		Data: slotSizes,
	}

	return []Feature{flexSlot}
}
