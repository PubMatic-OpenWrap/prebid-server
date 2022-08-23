package native_video

import "testing"

func TestMergeAdTemplate1MovingCar(t *testing.T) {
	// AdTemplate1 / Example1 (Moving Car)
	Merge(AdTemplate1, "/tmp/test1.mp4", Object{
		Subtype: BackgroundVideo,
		ID:      1,
		// FilePath: "http://localhost/hack22/background_AdobeExpress.mp4",
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/background_AdobeExpress.mp4",
		Duration: 10,
		Price:    5,
	}, Object{
		Subtype: Main,
		ID:      2,
		// FilePath: "http://localhost/hack22/moving_car_AdobeExpress.mp4",
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/moving_car_AdobeExpress.mp4",
		Duration: 10,
		Price:    5,
	}, Object{
		Subtype: Audio,
		ID:      1,
		// FilePath: "http://localhost/hack22/moving_car.mp3",
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/moving_car.mp3",
	}, Object{
		Subtype: Audio,
		ID:      2,
		// FilePath: "http://localhost/hack22/bmw_audio_ad.mp3",
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/bmw_audio_ad.mp3",
	})
}

// coca cola ad with 1 audio and video background
func TestMergeAdTemplate1Example2(t *testing.T) {
	Merge(AdTemplate1, "/tmp/coca-cola_ad.mp4", Object{
		Subtype:  Main,
		ID:       1,
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/coca_cola_AdobeExpress.mp4",
	}, Object{
		Subtype:  BackgroundVideo,
		ID:       2,
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/restaurant-background.mp4",
	}, Object{
		Subtype:  Audio,
		ID:       1,
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/coca_cola_audio.mp3",
	})
}

/*
ffmpeg -y -i ./templates/image.png -i ./templates/umbrella_2.mp4   -i ./templates/umbrella_audio.mp3   -filter_complex '[1:v]chromakey=0x42FB00:0.1:0.2[ckout];[0:v][ckout]overlay[out]' -map '[out]' -map '2' -shortest umbrella_ad_3.mp4
*/
func TestMergeAdTemplate2UmbrellaAd(t *testing.T) {
	Merge(AdTemplate2, "/tmp/umbrella_ad.mp4", Object{
		Subtype:  Main,
		ID:       1,
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/umbrella_2.mp4",
	}, Object{
		Subtype:  BackgroundImage,
		ID:       2,
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/image.png",
	}, Object{
		Subtype:  Audio,
		ID:       1,
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/umbrella_audio.mp3",
	})
}

func TestMergeTemplate3TourismAd(t *testing.T) {
	Merge(AdTemplate3, "/tmp/template_3_ad.mp4", Object{
		Subtype:  Main,
		ID:       1,
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/us.mp4",
	}, Object{
		Subtype:  BackgroundImage,
		ID:       2,
		FilePath: "/Users/shriprasadmarathe/workspace/native_in_video/templates/aus_travel.jpg",
	})
}
