package exchange

import (
	"testing"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func Test_updateContentObjectForBidder(t *testing.T) {

	createBidderRequest := func(BidRequest *openrtb2.BidRequest) []BidderRequest {
		newReq := *BidRequest
		newReq.ID = "2"
		return []BidderRequest{{
			BidderName: "pubmatic",
			BidRequest: BidRequest,
		},
			{
				BidderName: "appnexus",
				BidRequest: &newReq,
			},
		}
	}

	type args struct {
		BidRequest *openrtb2.BidRequest
		requestExt *openrtb_ext.ExtRequest
	}
	tests := []struct {
		name                    string
		args                    args
		wantedAllBidderRequests []BidderRequest
	}{
		{
			name: "no_transparency_object",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title: "Title1",
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
							},
						},
					},
				},
			},
		},
		{
			name: "no_content_object_in_app_site",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
					},
					Site: &openrtb2.Site{
						ID:   "1",
						Name: "Site1",
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: true,
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
						},
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Site1",
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
						},
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Site1",
						},
					},
				},
			},
		},
		{
			name: "no_partner_default_rules_in_transparency",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					Site: &openrtb2.Site{
						ID:   "1",
						Name: "Test",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Test",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Test",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
			},
		},
		{
			name: "include_all_keys_for_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					Site: &openrtb2.Site{
						ID:   "1",
						Name: "Test",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: true,
										Keys:    []string{},
									},
									"appnexus": {
										Include: false,
										Keys:    []string{},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Test",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Test",
						},
					},
				},
			},
		},
		{
			name: "exclude_all_keys_for_pubmatic_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: false,
										Keys:    []string{},
									},
									"appnexus": {
										Include: true,
										Keys:    []string{},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
			},
		},
		{
			name: "include_title_field_for_pubmatic_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: true,
										Keys:    []string{"title"},
									},
									"appnexus": {
										Include: false,
										Keys:    []string{"genre"},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
							},
						},
					},
				},
			},
		},
		{
			name: "exclude_title_field_for_pubmatic_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: false,
										Keys:    []string{"title"},
									},
									"appnexus": {
										Include: true,
										Keys:    []string{"genre"},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Genre: "Genre1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Genre: "Genre1",
							},
						},
					},
				},
			},
		},
		{
			name: "use_default_rule_for_pubmatic_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title:    "Title1",
							Genre:    "Genre1",
							Series:   "Series1",
							Season:   "Season1",
							Artist:   "Artist1",
							Album:    "Album1",
							ISRC:     "isrc1",
							Producer: &openrtb2.Producer{},
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"default": {
										Include: true,
										Keys: []string{
											"id", "episode", "series", "season", "artist", "genre", "album", "isrc", "producer", "url", "cat", "prodq", "videoquality", "context", "contentrating", "userrating", "qagmediarating", "livestream", "sourcerelationship", "len", "language", "embeddable", "data", "ext"},
									},
									"pubmatic": {
										Include: true,
										Keys:    []string{"title", "genre"},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Genre:    "Genre1",
								Series:   "Series1",
								Season:   "Season1",
								Artist:   "Artist1",
								Album:    "Album1",
								ISRC:     "isrc1",
								Producer: &openrtb2.Producer{},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allBidderRequests := createBidderRequest(tt.args.BidRequest)
			updateContentObjectForBidder(allBidderRequests, tt.args.requestExt)
			assert.Equal(t, tt.wantedAllBidderRequests, allBidderRequests, tt.name)
		})
	}
}

func Benchmark_updateContentObjectForBidder(b *testing.B) {

	createBidderRequest := func(BidRequest *openrtb2.BidRequest) []BidderRequest {
		newReq := *BidRequest
		newReq.ID = "2"
		return []BidderRequest{{
			BidderName: "pubmatic",
			BidRequest: BidRequest,
		},
			{
				BidderName: "appnexus",
				BidRequest: &newReq,
			},
		}
	}

	type args struct {
		BidRequest *openrtb2.BidRequest
		requestExt *openrtb_ext.ExtRequest
	}
	tests := []struct {
		name                    string
		args                    args
		wantedAllBidderRequests []BidderRequest
	}{
		{
			name: "no_transparency_object",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title: "Title1",
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
							},
						},
					},
				},
			},
		},
		{
			name: "no_content_object_in_app_site",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
					},
					Site: &openrtb2.Site{
						ID:   "1",
						Name: "Site1",
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: true,
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
						},
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Site1",
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
						},
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Site1",
						},
					},
				},
			},
		},
		{
			name: "no_partner_default_rules_in_transparency",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					Site: &openrtb2.Site{
						ID:   "1",
						Name: "Test",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Test",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Test",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
			},
		},
		{
			name: "include_all_keys_for_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					Site: &openrtb2.Site{
						ID:   "1",
						Name: "Test",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: true,
										Keys:    []string{},
									},
									"appnexus": {
										Include: false,
										Keys:    []string{},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Test",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						Site: &openrtb2.Site{
							ID:   "1",
							Name: "Test",
						},
					},
				},
			},
		},
		{
			name: "exclude_all_keys_for_pubmatic_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: false,
										Keys:    []string{},
									},
									"appnexus": {
										Include: true,
										Keys:    []string{},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
			},
		},
		{
			name: "include_title_field_for_pubmatic_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: true,
										Keys:    []string{"title"},
									},
									"appnexus": {
										Include: false,
										Keys:    []string{"genre"},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
							},
						},
					},
				},
			},
		},
		{
			name: "exclude_title_field_for_pubmatic_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title: "Title1",
							Genre: "Genre1",
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"pubmatic": {
										Include: false,
										Keys:    []string{"title"},
									},
									"appnexus": {
										Include: true,
										Keys:    []string{"genre"},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Genre: "Genre1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Genre: "Genre1",
							},
						},
					},
				},
			},
		},
		{
			name: "use_default_rule_for_pubmatic_bidder",
			args: args{
				BidRequest: &openrtb2.BidRequest{
					ID: "1",
					App: &openrtb2.App{
						ID:     "1",
						Name:   "Test",
						Bundle: "com.pubmatic.app",
						Content: &openrtb2.Content{
							Title:    "Title1",
							Genre:    "Genre1",
							Series:   "Series1",
							Season:   "Season1",
							Artist:   "Artist1",
							Album:    "Album1",
							ISRC:     "isrc1",
							Producer: &openrtb2.Producer{},
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						ExtOWRequestPrebid: openrtb_ext.ExtOWRequestPrebid{
							Transparency: &openrtb_ext.TransparencyExt{
								Content: map[string]openrtb_ext.TransparencyRule{
									"default": {
										Include: true,
										Keys: []string{
											"id", "episode", "series", "season", "artist", "genre", "album", "isrc", "producer", "url", "cat", "prodq", "videoquality", "context", "contentrating", "userrating", "qagmediarating", "livestream", "sourcerelationship", "len", "language", "embeddable", "data", "ext"},
									},
									"pubmatic": {
										Include: true,
										Keys:    []string{"title", "genre"},
									},
								},
							},
						},
					},
				},
			},
			wantedAllBidderRequests: []BidderRequest{
				{
					BidderName: "pubmatic",
					BidRequest: &openrtb2.BidRequest{
						ID: "1",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Title: "Title1",
								Genre: "Genre1",
							},
						},
					},
				},
				{
					BidderName: "appnexus",
					BidRequest: &openrtb2.BidRequest{
						ID: "2",
						App: &openrtb2.App{
							ID:     "1",
							Name:   "Test",
							Bundle: "com.pubmatic.app",
							Content: &openrtb2.Content{
								Genre:    "Genre1",
								Series:   "Series1",
								Season:   "Season1",
								Artist:   "Artist1",
								Album:    "Album1",
								ISRC:     "isrc1",
								Producer: &openrtb2.Producer{},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			allBidderRequests := createBidderRequest(tt.args.BidRequest)
			for i := 0; i < b.N; i++ {
				updateContentObjectForBidder(allBidderRequests, tt.args.requestExt)
			}
			//assert.Equal(t, tt.wantedAllBidderRequests, allBidderRequests, tt.name)
		})
	}
}

func TestUpdateContentObj(t *testing.T) {
	tests := []struct {
		name          string
		request       *openrtb2.BidRequest
		contentObject *openrtb2.Content
		isApp         bool
		expected      *openrtb2.BidRequest
	}{
		{
			name: "update_app_content",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					ID:      "app1",
					Name:    "TestApp",
					Content: &openrtb2.Content{ID: "old_content"},
				},
			},
			contentObject: &openrtb2.Content{ID: "new_content"},
			isApp:         true,
			expected: &openrtb2.BidRequest{
				App: &openrtb2.App{
					ID:      "app1",
					Name:    "TestApp",
					Content: &openrtb2.Content{ID: "new_content"},
				},
			},
		},
		{
			name: "update_site_content",
			request: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					ID:      "site1",
					Name:    "TestSite",
					Content: &openrtb2.Content{ID: "old_content"},
				},
			},
			contentObject: &openrtb2.Content{ID: "new_content"},
			isApp:         false,
			expected: &openrtb2.BidRequest{
				Site: &openrtb2.Site{
					ID:      "site1",
					Name:    "TestSite",
					Content: &openrtb2.Content{ID: "new_content"},
				},
			},
		},
		{
			name: "set_app_content_to_nil",
			request: &openrtb2.BidRequest{
				App: &openrtb2.App{
					ID:      "app1",
					Name:    "TestApp",
					Content: &openrtb2.Content{ID: "old_content"},
				},
			},
			contentObject: nil,
			isApp:         true,
			expected: &openrtb2.BidRequest{
				App: &openrtb2.App{
					ID:      "app1",
					Name:    "TestApp",
					Content: nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updateContentObj(tt.request, tt.contentObject, tt.isApp)
			assert.Equal(t, tt.expected, tt.request)
		})
	}
}

func TestDeepCopyContentNetworkObj(t *testing.T) {
	tests := []struct {
		name     string
		network  *openrtb2.Network
		expected *openrtb2.Network
	}{
		{
			name:     "nil_network",
			network:  nil,
			expected: nil,
		},
		{
			name: "network_with_all_fields",
			network: &openrtb2.Network{
				ID:     "network1",
				Name:   "Test Network",
				Domain: "test.com",
				Ext:    []byte(`{"key":"value"}`),
			},
			expected: &openrtb2.Network{
				ID:     "network1",
				Name:   "Test Network",
				Domain: "test.com",
				Ext:    []byte(`{"key":"value"}`),
			},
		},
		{
			name: "network_without_ext",
			network: &openrtb2.Network{
				ID:     "network2",
				Name:   "Test Network 2",
				Domain: "test2.com",
			},
			expected: &openrtb2.Network{
				ID:     "network2",
				Name:   "Test Network 2",
				Domain: "test2.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepCopyContentNetworkObj(tt.network)
			assert.Equal(t, tt.expected, result)
			if tt.network != nil {
				assert.NotSame(t, tt.network, result)
				assert.NotSame(t, tt.network.Ext, result.Ext)
			}
		})
	}
}

func TestDeepCopyContentChannelObj(t *testing.T) {
	tests := []struct {
		name     string
		channel  *openrtb2.Channel
		expected *openrtb2.Channel
	}{
		{
			name:     "nil_channel",
			channel:  nil,
			expected: nil,
		},
		{
			name: "channel_with_all_fields",
			channel: &openrtb2.Channel{
				ID:     "channel1",
				Name:   "Test Channel",
				Domain: "test.com",
				Ext:    []byte(`{"key":"value"}`),
			},
			expected: &openrtb2.Channel{
				ID:     "channel1",
				Name:   "Test Channel",
				Domain: "test.com",
				Ext:    []byte(`{"key":"value"}`),
			},
		},
		{
			name: "channel_without_ext",
			channel: &openrtb2.Channel{
				ID:     "channel2",
				Name:   "Test Channel 2",
				Domain: "test2.com",
			},
			expected: &openrtb2.Channel{
				ID:     "channel2",
				Name:   "Test Channel 2",
				Domain: "test2.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepCopyContentChannelObj(tt.channel)
			assert.Equal(t, tt.expected, result)
			if tt.channel != nil {
				assert.NotSame(t, tt.channel, result)
				assert.NotSame(t, tt.channel.Ext, result.Ext)
			}
		})
	}
}

func TestDeepCopyContentProducer(t *testing.T) {
	tests := []struct {
		name     string
		producer *openrtb2.Producer
		expected *openrtb2.Producer
	}{
		{
			name:     "nil_producer",
			producer: nil,
			expected: nil,
		},
		{
			name: "producer_with_all_fields",
			producer: &openrtb2.Producer{
				ID:     "producer1",
				Name:   "Test Producer",
				Domain: "test.com",
				Cat:    []string{"cat1", "cat2"},
				Ext:    []byte(`{"key":"value"}`),
			},
			expected: &openrtb2.Producer{
				ID:     "producer1",
				Name:   "Test Producer",
				Domain: "test.com",
				Cat:    []string{"cat1", "cat2"},
				Ext:    []byte(`{"key":"value"}`),
			},
		},
		{
			name: "producer_without_optional_fields",
			producer: &openrtb2.Producer{
				ID:   "producer2",
				Name: "Test Producer 2",
			},
			expected: &openrtb2.Producer{
				ID:   "producer2",
				Name: "Test Producer 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepCopyContentProducer(tt.producer)
			assert.Equal(t, tt.expected, result)
			if tt.producer != nil {
				assert.NotSame(t, tt.producer, result)
				if tt.producer.Cat != nil {
					assert.NotSame(t, tt.producer.Cat, result.Cat)
				}
				if tt.producer.Ext != nil {
					assert.NotSame(t, tt.producer.Ext, result.Ext)
				}
			}
		})
	}
}

func TestDeepCopyContentObj(t *testing.T) {
	tests := []struct {
		name     string
		content  *openrtb2.Content
		include  bool
		keys     []string
		expected *openrtb2.Content
	}{
		{
			name:     "nil_content",
			content:  nil,
			include:  true,
			keys:     []string{},
			expected: nil,
		},
		{
			name: "include_false_no_keys_returns_nil",
			content: &openrtb2.Content{
				ID:    "test",
				Title: "Test Content",
			},
			include:  false,
			keys:     []string{},
			expected: nil,
		},
		{
			name: "include_true_no_keys_returns_full_copy",
			content: &openrtb2.Content{
				ID:      "test",
				Title:   "Test Content",
				Episode: 1,
				Series:  "Test Series",
				Season:  "Season 1",
				Artist:  "Test Artist",
				Genre:   "Test Genre",
				Album:   "Test Album",
				ISRC:    "TEST123",
			},
			include: true,
			keys:    []string{},
			expected: &openrtb2.Content{
				ID:      "test",
				Title:   "Test Content",
				Episode: 1,
				Series:  "Test Series",
				Season:  "Season 1",
				Artist:  "Test Artist",
				Genre:   "Test Genre",
				Album:   "Test Album",
				ISRC:    "TEST123",
			},
		},
		{
			name: "include_true_specific_keys_keeps_only_specified",
			content: &openrtb2.Content{
				ID:      "test",
				Title:   "Test Content",
				Episode: 1,
				Series:  "Test Series",
			},
			include: true,
			keys:    []string{"id", "title"},
			expected: &openrtb2.Content{
				ID:      "test",
				Title:   "Test Content",
				Episode: 0,
				Series:  "",
			},
		},
		{
			name: "include_false_specific_keys_clears_only_specified",
			content: &openrtb2.Content{
				ID:      "test",
				Title:   "Test Content",
				Episode: 1,
				Series:  "Test Series",
			},
			include: false,
			keys:    []string{"id", "title"},
			expected: &openrtb2.Content{
				ID:      "",
				Title:   "",
				Episode: 1,
				Series:  "Test Series",
			},
		},
		{
			name: "deep_copy_complex_fields",
			content: &openrtb2.Content{
				ID:    "test",
				Title: "Test Content",
				Producer: &openrtb2.Producer{
					ID:   "prod1",
					Name: "Test Producer",
					Cat:  []string{"cat1", "cat2"},
				},
				Network: &openrtb2.Network{
					ID:   "net1",
					Name: "Test Network",
				},
				Channel: &openrtb2.Channel{
					ID:   "chan1",
					Name: "Test Channel",
				},
				Cat:  []string{"category1", "category2"},
				Data: []openrtb2.Data{{ID: "data1", Name: "Test Data"}},
			},
			include: true,
			keys:    []string{},
			expected: &openrtb2.Content{
				ID:    "test",
				Title: "Test Content",
				Producer: &openrtb2.Producer{
					ID:   "prod1",
					Name: "Test Producer",
					Cat:  []string{"cat1", "cat2"},
				},
				Network: &openrtb2.Network{
					ID:   "net1",
					Name: "Test Network",
				},
				Channel: &openrtb2.Channel{
					ID:   "chan1",
					Name: "Test Channel",
				},
				Cat:  []string{"category1", "category2"},
				Data: []openrtb2.Data{{ID: "data1", Name: "Test Data"}},
			},
		},
		{
			name: "include_true_complex_field_keys",
			content: &openrtb2.Content{
				ID:    "test",
				Title: "Test Content",
				Producer: &openrtb2.Producer{
					ID:   "prod1",
					Name: "Test Producer",
				},
				Network: &openrtb2.Network{
					ID:   "net1",
					Name: "Test Network",
				},
			},
			include: true,
			keys:    []string{"id", "producer"},
			expected: &openrtb2.Content{
				ID:    "test",
				Title: "",
				Producer: &openrtb2.Producer{
					ID:   "prod1",
					Name: "Test Producer",
				},
				Network: nil,
			},
		},
		{
			name: "include_false_complex_field_keys",
			content: &openrtb2.Content{
				ID:    "test",
				Title: "Test Content",
				Producer: &openrtb2.Producer{
					ID:   "prod1",
					Name: "Test Producer",
				},
				Network: &openrtb2.Network{
					ID:   "net1",
					Name: "Test Network",
				},
			},
			include: false,
			keys:    []string{"id", "producer"},
			expected: &openrtb2.Content{
				ID:       "",
				Title:    "Test Content",
				Producer: nil,
				Network: &openrtb2.Network{
					ID:   "net1",
					Name: "Test Network",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepCopyContentObj(tt.content, tt.include, tt.keys)
			assert.Equal(t, tt.expected, result, "Test case %s failed", tt.name)
		})
	}
}
