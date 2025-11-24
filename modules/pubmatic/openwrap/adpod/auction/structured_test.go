package auction

// func TestSelectStructuredPodBids(t *testing.T) {
// 	tests := []struct {
// 		name   string
// 		podCfg models.AdpodConfig
// 		resp   *openrtb2.BidResponse
// 		want   []*openrtb2.Bid // expected winners (only ID + ImpID are validated)
// 	}{
// 		{
// 			name: "basic_highest_price_per_slot",
// 			podCfg: models.AdpodConfig{
// 				Slots: []models.SlotConfig{
// 					{Id: "imp1", MinDuration: 15, MaxDuration: 30},
// 					{Id: "imp2", MinDuration: 15, MaxDuration: 30},
// 				},
// 			},
// 			resp: &openrtb2.BidResponse{
// 				SeatBid: []openrtb2.SeatBid{
// 					{
// 						Seat: "test",
// 						Bid: []openrtb2.Bid{
// 							{ID: "b1-low", ImpID: "imp1", Price: 2.0, Dur: 30},
// 							{ID: "b1-high", ImpID: "imp1", Price: 3.0, Dur: 30},
// 							{ID: "b2-low", ImpID: "imp2", Price: 1.0, Dur: 30},
// 							{ID: "b2-high", ImpID: "imp2", Price: 5.0, Dur: 30},
// 						},
// 					},
// 				},
// 			},
// 			want: []*openrtb2.Bid{
// 				{ID: "b1-high", ImpID: "imp1"},
// 				{ID: "b2-high", ImpID: "imp2"},
// 			},
// 		},
// 		{
// 			name: "domain_exclusion_across_slots",
// 			podCfg: models.AdpodConfig{
// 				Slots: []models.SlotConfig{
// 					{Id: "imp1", MinDuration: 15, MaxDuration: 30},
// 					{Id: "imp2", MinDuration: 15, MaxDuration: 30},
// 				},
// 				Exclusion: models.ExclusionConfig{
// 					AdvertiserDomainExclusion: true,
// 				},
// 			},
// 			resp: &openrtb2.BidResponse{
// 				SeatBid: []openrtb2.SeatBid{
// 					{
// 						Seat: "s1",
// 						Bid: []openrtb2.Bid{
// 							{ID: "s1-imp1", ImpID: "imp1", Price: 3.0, Dur: 30, ADomain: []string{"a.com"}},
// 							{ID: "s1-imp2-conflict", ImpID: "imp2", Price: 2.0, Dur: 30, ADomain: []string{"a.com"}},
// 							{ID: "s1-imp2-ok", ImpID: "imp2", Price: 1.5, Dur: 30, ADomain: []string{"b.com"}},
// 						},
// 					},
// 				},
// 			},
// 			want: []*openrtb2.Bid{
// 				{ID: "s1-imp1", ImpID: "imp1"},
// 				{ID: "s1-imp2-ok", ImpID: "imp2"},
// 			},
// 		},
// 		{
// 			name: "deal_priority_over_higher_price",
// 			podCfg: models.AdpodConfig{
// 				Slots: []models.SlotConfig{
// 					{Id: "imp1", MinDuration: 15, MaxDuration: 30},
// 				},
// 			},
// 			resp: &openrtb2.BidResponse{
// 				SeatBid: []openrtb2.SeatBid{
// 					{
// 						Seat: "s1",
// 						Bid: []openrtb2.Bid{
// 							{ID: "deal", ImpID: "imp1", Price: 2.0, Dur: 30, DealID: "d1"},
// 							{ID: "nondeal", ImpID: "imp1", Price: 5.0, Dur: 30},
// 						},
// 					},
// 				},
// 			},
// 			want: []*openrtb2.Bid{
// 				{ID: "deal", ImpID: "imp1"},
// 			},
// 		},
// 		{
// 			name:   "empty_inputs",
// 			podCfg: models.AdpodConfig{},
// 			resp:   &openrtb2.BidResponse{},
// 			want:   nil,
// 		},
// 		{
// 			name: "deal priority and exclusion present, expect max pod value",
// 			podCfg: models.AdpodConfig{
// 				Slots: []models.SlotConfig{
// 					{Id: "imp1", MinDuration: 15, MaxDuration: 30},
// 					{Id: "imp2", MinDuration: 15, MaxDuration: 30},
// 					{Id: "imp3", MinDuration: 15, MaxDuration: 30},
// 				},
// 				Exclusion: models.ExclusionConfig{
// 					AdvertiserDomainExclusion: true,
// 					IABCategoryExclusion:      true,
// 				},
// 			},
// 			resp: &openrtb2.BidResponse{
// 				SeatBid: []openrtb2.SeatBid{
// 					{
// 						Seat: "s1",
// 						Bid: []openrtb2.Bid{
// 							{ID: "deal", ImpID: "imp1", Price: 3.0, Dur: 30, DealID: "d1", ADomain: []string{"a.com"}, Cat: []string{"IAB1"}},
// 							{ID: "nondeal", ImpID: "imp1", Price: 8.0, Dur: 30, ADomain: []string{"b.com"}},
// 							{ID: "deal", ImpID: "imp2", Price: 4.0, Dur: 30, DealID: "d1", ADomain: []string{"a.com"}, Cat: []string{"IAB1"}},
// 							{ID: "nondeal", ImpID: "imp2", Price: 7.0, Dur: 30},
// 							{ID: "deal", ImpID: "imp3", Price: 2.0, Dur: 30, DealID: "d1", ADomain: []string{"c.com"}, Cat: []string{"IAB3"}},
// 							{ID: "nondeal", ImpID: "imp3", Price: 5.0, Dur: 30, ADomain: []string{"c.com"}, Cat: []string{"IAB3"}},
// 						},
// 					},
// 				},
// 			},
// 			want: []*openrtb2.Bid{
// 				{ID: "nondeal", ImpID: "imp1"},
// 				{ID: "deal", ImpID: "imp2"},
// 				{ID: "deal", ImpID: "imp3"},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := SelectStructuredPodBids(tt.podCfg, tt.resp)

// 			if tt.want == nil {
// 				if got != nil {
// 					t.Fatalf("expected nil, got %d bids", len(got))
// 				}
// 				return
// 			}

// 			if len(got) != len(tt.want) {
// 				t.Fatalf("expected %d bids, got %d", len(tt.want), len(got))
// 			}

// 			// Compare by (ImpID, ID)
// 			gotByImp := map[string]*openrtb2.Bid{}
// 			for _, b := range got {
// 				gotByImp[b.ImpID] = b
// 			}

// 			for _, w := range tt.want {
// 				gb := gotByImp[w.ImpID]
// 				if gb == nil {
// 					t.Errorf("imp %s: expected bid ID %q, got nil", w.ImpID, w.ID)
// 					continue
// 				}
// 				if gb.ID != w.ID {
// 					t.Errorf("imp %s: expected bid ID %q, got %q", w.ImpID, w.ID, gb.ID)
// 				}
// 			}
// 		})
// 	}
// }
