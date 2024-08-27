package ortb

import (
	"github.com/PubMatic-OpenWrap/prebid-server/v2/util/ptrutil"
	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v2/util/sliceutil"
)

func DeepCloneImpression(imp *openrtb2.Imp) *openrtb2.Imp {
	clone := *imp

	clone.Metric = sliceutil.Clone(imp.Metric)
	clone.Banner = DeepCopyImpBanner(imp.Banner)
	clone.Video = DeepCopyImpVideo(imp.Video)
	clone.Audio = DeepCopyImpAudio(imp.Audio)
	clone.Native = DeepCopyImpNative(imp.Native)
	clone.PMP = DeepCopyImpPMP(imp.PMP)
	clone.ClickBrowser = ptrutil.Clone(imp.ClickBrowser)
	clone.Secure = ptrutil.Clone(imp.Secure)
	clone.IframeBuster = sliceutil.Clone(imp.IframeBuster)
	clone.Qty = ptrutil.Clone(imp.Qty)
	clone.Refresh = ptrutil.Clone(imp.Refresh)
	clone.Ext = sliceutil.Clone(imp.Ext)
	return &clone
}

func DeepCopyImpVideo(video *openrtb2.Video) *openrtb2.Video {
	if video == nil {
		return nil
	}

	videoCopy := *video
	videoCopy.MIMEs = sliceutil.Clone(video.MIMEs)
	videoCopy.StartDelay = ptrutil.Clone(video.StartDelay)
	videoCopy.Protocols = sliceutil.Clone(video.Protocols)
	videoCopy.W = ptrutil.Clone(video.W)
	videoCopy.H = ptrutil.Clone(video.H)
	videoCopy.RqdDurs = sliceutil.Clone(video.RqdDurs)
	videoCopy.Skip = ptrutil.Clone(video.Skip)
	videoCopy.BAttr = sliceutil.Clone(video.BAttr)
	videoCopy.BoxingAllowed = ptrutil.Clone(video.BoxingAllowed)
	videoCopy.PlaybackMethod = sliceutil.Clone(video.PlaybackMethod)
	videoCopy.Delivery = sliceutil.Clone(video.Delivery)
	videoCopy.Pos = ptrutil.Clone(video.Pos)
	videoCopy.CompanionAd = sliceutil.Clone(video.CompanionAd)
	videoCopy.API = sliceutil.Clone(video.API)
	videoCopy.CompanionType = sliceutil.Clone(video.CompanionType)
	videoCopy.DurFloors = sliceutil.Clone(video.DurFloors)
	videoCopy.Ext = sliceutil.Clone(video.Ext)
	return &videoCopy
}

func DeepCopyImpNative(native *openrtb2.Native) *openrtb2.Native {
	if native == nil {
		return nil
	}

	nativeCopy := *native
	nativeCopy.API = sliceutil.Clone(native.API)
	nativeCopy.BAttr = sliceutil.Clone(native.BAttr)
	nativeCopy.Ext = sliceutil.Clone(native.Ext)
	return &nativeCopy
}

func DeepCopyImpBanner(banner *openrtb2.Banner) *openrtb2.Banner {
	if banner == nil {
		return nil
	}

	bannerCopy := *banner
	bannerCopy.Format = sliceutil.Clone(banner.Format)
	bannerCopy.W = ptrutil.Clone(banner.W)
	bannerCopy.H = ptrutil.Clone(banner.H)
	bannerCopy.BType = sliceutil.Clone(banner.BType)
	bannerCopy.BAttr = sliceutil.Clone(banner.BAttr)
	bannerCopy.MIMEs = sliceutil.Clone(banner.MIMEs)
	bannerCopy.ExpDir = sliceutil.Clone(banner.ExpDir)
	bannerCopy.API = sliceutil.Clone(banner.API)
	bannerCopy.Vcm = ptrutil.Clone(banner.Vcm)
	bannerCopy.Ext = sliceutil.Clone(banner.Ext)
	return &bannerCopy
}

func DeepCopyImpAudio(audio *openrtb2.Audio) *openrtb2.Audio {
	if audio == nil {
		return nil
	}

	audioCopy := *audio
	audioCopy.MIMEs = sliceutil.Clone(audio.MIMEs)
	audioCopy.Protocols = sliceutil.Clone(audio.Protocols)
	audioCopy.StartDelay = ptrutil.Clone(audio.StartDelay)
	audioCopy.RqdDurs = sliceutil.Clone(audio.RqdDurs)
	audioCopy.BAttr = sliceutil.Clone(audio.BAttr)
	audioCopy.Delivery = sliceutil.Clone(audio.Delivery)
	audioCopy.CompanionAd = sliceutil.Clone(audio.CompanionAd)
	audioCopy.API = sliceutil.Clone(audio.API)
	audioCopy.CompanionType = sliceutil.Clone(audio.CompanionType)
	audioCopy.Stitched = ptrutil.Clone(audio.Stitched)
	audioCopy.NVol = ptrutil.Clone(audio.NVol)
	audioCopy.DurFloors = sliceutil.Clone(audio.DurFloors)
	audioCopy.Ext = sliceutil.Clone(audio.Ext)
	return &audioCopy
}

func DeepCopyImpPMP(pmp *openrtb2.PMP) *openrtb2.PMP {
	if pmp == nil {
		return nil
	}

	pmpCopy := *pmp
	pmpCopy.Deals = sliceutil.Clone(pmp.Deals)
	pmpCopy.Ext = sliceutil.Clone(pmp.Ext)
	return &pmpCopy
}
