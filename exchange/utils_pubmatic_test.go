package exchange

import (
	"testing"

	"github.com/mxmCherry/openrtb/v15/openrtb2"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/stretchr/testify/assert"
)

func Test_updateContentObjectForBidder(t *testing.T) {
	type args struct {
		allBidderRequests []BidderRequest
		requestExt        *openrtb_ext.ExtRequest
	}
	tests := []struct {
		name                    string
		args                    args
		wantedAllBidderRequests []BidderRequest
	}{
		{
			name: "No Transparency Object",
			args: args{
				allBidderRequests: []BidderRequest{
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
			},
		},
		{
			name: "No Content Object in App/Site",
			args: args{
				allBidderRequests: []BidderRequest{
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
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Transparency: &openrtb_ext.TransparencyExt{
							Content: map[string]openrtb_ext.TransparencyRule{},
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
			},
		},
		{
			name: "No partner/ default rules in tranpsarency",
			args: args{
				allBidderRequests: []BidderRequest{
					{
						BidderName: "pubmatic",
						BidRequest: &openrtb2.BidRequest{
							ID: "1",
							Site: &openrtb2.Site{
								ID:   "1",
								Name: "Test",
								Content: &openrtb2.Content{
									Title: "Title1",
								},
							},
						},
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Transparency: &openrtb_ext.TransparencyExt{
							Content: map[string]openrtb_ext.TransparencyRule{},
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
						},
					},
				},
			},
		},
		{
			name: "Include All keys for pubmatic bidder",
			args: args{
				allBidderRequests: []BidderRequest{
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
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Transparency: &openrtb_ext.TransparencyExt{
							Content: map[string]openrtb_ext.TransparencyRule{
								"pubmatic": {
									Include: true,
									Keys:    []string{},
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
			},
		},
		{
			name: "Exclude All keys for pubmatic bidder",
			args: args{
				allBidderRequests: []BidderRequest{
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
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Transparency: &openrtb_ext.TransparencyExt{
							Content: map[string]openrtb_ext.TransparencyRule{
								"pubmatic": {
									Include: false,
									Keys:    []string{},
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
			},
		},
		{
			name: "Include title field for pubmatic bidder",
			args: args{
				allBidderRequests: []BidderRequest{
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
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Transparency: &openrtb_ext.TransparencyExt{
							Content: map[string]openrtb_ext.TransparencyRule{
								"pubmatic": {
									Include: true,
									Keys:    []string{"title"},
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
			},
		},
		{
			name: "Exclude title field for pubmatic bidder",
			args: args{
				allBidderRequests: []BidderRequest{
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
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Transparency: &openrtb_ext.TransparencyExt{
							Content: map[string]openrtb_ext.TransparencyRule{
								"pubmatic": {
									Include: false,
									Keys:    []string{"title"},
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
			},
		},
		{
			name: "Use default rule for pubmatic bidder",
			args: args{
				allBidderRequests: []BidderRequest{
					{
						BidderName: "pubmatic",
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
					},
				},
				requestExt: &openrtb_ext.ExtRequest{
					Prebid: openrtb_ext.ExtRequestPrebid{
						Transparency: &openrtb_ext.TransparencyExt{
							Content: map[string]openrtb_ext.TransparencyRule{
								"default": {
									Include: true,
									Keys: []string{
										"id", "episode", "series", "season", "artist", "genre", "album", "isrc", "producer", "url", "cat", "prodq", "videoquality", "context", "contentrating", "userrating", "qagmediarating", "livestream", "sourcerelationship", "len", "language", "embeddable", "data", "ext"},
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
			updateContentObjectForBidder(tt.args.allBidderRequests, tt.args.requestExt)
			assert.Equal(t, tt.wantedAllBidderRequests, tt.args.allBidderRequests, tt.name)
		})
	}
}

func Test_excludeKeys(t *testing.T) {
	type args struct {
		contentObject *openrtb2.Content
		keys          []string
	}
	tests := []struct {
		name string
		args args
		want *openrtb2.Content
	}{
		{
			name: "exclude key episode",
			args: args{
				contentObject: &openrtb2.Content{ID: "123", Episode: 456},
				keys:          []string{"episode"},
			},
			want: &openrtb2.Content{ID: "123"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := excludeKeys(tt.args.contentObject, tt.args.keys)
			assert.Equal(t, tt.want, got)
		})
	}
}
