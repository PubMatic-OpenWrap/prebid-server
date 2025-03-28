package gocache

import (
	"testing"

	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v3/modules/pubmatic/openwrap/models/adunitconfig"
	"github.com/stretchr/testify/assert"
)

func Test_validUPixels(t *testing.T) {
	type args struct {
		pixel []adunitconfig.UniversalPixel
	}
	tests := []struct {
		name string
		args args
		want []adunitconfig.UniversalPixel
	}{
		{
			name: "No partners",
			args: args{
				pixel: []adunitconfig.UniversalPixel{
					{
						Id:        123,
						Pixel:     "sample.com",
						PixelType: models.PixelTypeJS,
						Pos:       models.PixelPosAbove,
						MediaType: "video",
					},
				},
			},
			want: []adunitconfig.UniversalPixel{
				{
					Id:        123,
					Pixel:     "sample.com",
					PixelType: models.PixelTypeJS,
					Pos:       models.PixelPosAbove,
					MediaType: "video",
				},
			},
		},
		{
			name: "No Pixel",
			args: args{
				pixel: []adunitconfig.UniversalPixel{
					{
						Id:        123,
						PixelType: models.PixelTypeJS,
						Pos:       models.PixelPosAbove,
						MediaType: "video",
					},
				},
			},
			want: nil,
		},
		{
			name: "Invalid Pixeltype",
			args: args{
				pixel: []adunitconfig.UniversalPixel{
					{
						Id:        123,
						Pixel:     "sample.com",
						PixelType: "invalid",
						Pos:       models.PixelPosAbove,
						MediaType: "banner",
						Partners:  []string{"pubmatic", "appnexus"},
					},
				},
			},
			want: nil,
		},
		{
			name: "Pixeltype Not Present",
			args: args{
				pixel: []adunitconfig.UniversalPixel{
					{
						Id:        123,
						Pixel:     "sample.com",
						Pos:       models.PixelPosAbove,
						MediaType: "banner",
						Partners:  []string{"pubmatic", "appnexus"},
					},
				},
			},
			want: nil,
		},
		{
			name: "Invalid Value of Pos and Other Valid Pixel",
			args: args{
				pixel: []adunitconfig.UniversalPixel{
					{
						Id:        123,
						Pixel:     "sample.com",
						PixelType: models.PixelTypeJS,
						Pos:       "invalid",
						MediaType: "banner",
						Partners:  []string{"pubmatic", "appnexus"},
					},
					{
						Id:        123,
						Pixel:     "sample.com",
						PixelType: models.PixelTypeJS,
						Pos:       models.PixelPosAbove,
						MediaType: "banner",
						Partners:  []string{"pubmatic", "appnexus"},
					},
				},
			},
			want: []adunitconfig.UniversalPixel{
				{
					Id:        123,
					Pixel:     "sample.com",
					PixelType: models.PixelTypeJS,
					Pos:       models.PixelPosAbove,
					MediaType: "banner",
					Partners:  []string{"pubmatic", "appnexus"},
				},
			},
		},
		{
			name: "No Pos Value",
			args: args{
				pixel: []adunitconfig.UniversalPixel{
					{
						Id:        123,
						Pixel:     "sample.com",
						PixelType: models.PixelTypeJS,
						MediaType: "banner",
						Partners:  []string{"pubmatic", "appnexus"},
					},
				},
			},
			want: []adunitconfig.UniversalPixel{
				{
					Id:        123,
					Pixel:     "sample.com",
					PixelType: models.PixelTypeJS,
					MediaType: "banner",
					Partners:  []string{"pubmatic", "appnexus"},
				},
			},
		},
		{
			name: "Valid UPixel",
			args: args{
				pixel: []adunitconfig.UniversalPixel{
					{
						Id:        123,
						Pixel:     "sample.com",
						PixelType: models.PixelTypeJS,
						Pos:       models.PixelPosAbove,
						MediaType: "banner",
						Partners:  []string{"pubmatic", "appnexus"},
					},
				},
			},
			want: []adunitconfig.UniversalPixel{
				{
					Id:        123,
					Pixel:     "sample.com",
					PixelType: models.PixelTypeJS,
					Pos:       models.PixelPosAbove,
					MediaType: "banner",
					Partners:  []string{"pubmatic", "appnexus"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validUPixels(tt.args.pixel)
			assert.Equal(t, tt.want, got)
		})
	}
}
