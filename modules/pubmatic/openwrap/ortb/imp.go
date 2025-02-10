package ortb

import (
	"slices"

	"github.com/prebid/openrtb/v20/openrtb2"
	"github.com/prebid/prebid-server/v3/util/ptrutil"
)

func DeepCloneImpression(imp *openrtb2.Imp) *openrtb2.Imp {
	clone := *imp

	clone.Metric = slices.Clone(imp.Metric)
	clone.Banner = DeepCopyImpBanner(imp.Banner)
	clone.Video = DeepCopyImpVideo(imp.Video)
	clone.Audio = DeepCopyImpAudio(imp.Audio)
	clone.Native = DeepCopyImpNative(imp.Native)
	clone.PMP = DeepCopyImpPMP(imp.PMP)
	clone.ClickBrowser = ptrutil.Clone(imp.ClickBrowser)
	clone.Secure = ptrutil.Clone(imp.Secure)
	clone.IframeBuster = slices.Clone(imp.IframeBuster)
	clone.Qty = ptrutil.Clone(imp.Qty)
	clone.Refresh = ptrutil.Clone(imp.Refresh)
	clone.Ext = slices.Clone(imp.Ext)
	return &clone
}

func DeepCopyImpVideo(video *openrtb2.Video) *openrtb2.Video {
	if video == nil {
		return nil
	}

	videoCopy := *video
	videoCopy.MIMEs = slices.Clone(video.MIMEs)
	videoCopy.StartDelay = ptrutil.Clone(video.StartDelay)
	videoCopy.Protocols = slices.Clone(video.Protocols)
	videoCopy.W = ptrutil.Clone(video.W)
	videoCopy.H = ptrutil.Clone(video.H)
	videoCopy.RqdDurs = slices.Clone(video.RqdDurs)
	videoCopy.Skip = ptrutil.Clone(video.Skip)
	videoCopy.BAttr = slices.Clone(video.BAttr)
	videoCopy.BoxingAllowed = ptrutil.Clone(video.BoxingAllowed)
	videoCopy.PlaybackMethod = slices.Clone(video.PlaybackMethod)
	videoCopy.Delivery = slices.Clone(video.Delivery)
	videoCopy.Pos = ptrutil.Clone(video.Pos)
	videoCopy.CompanionAd = slices.Clone(video.CompanionAd)
	videoCopy.API = slices.Clone(video.API)
	videoCopy.CompanionType = slices.Clone(video.CompanionType)
	videoCopy.DurFloors = slices.Clone(video.DurFloors)
	videoCopy.Ext = slices.Clone(video.Ext)
	return &videoCopy
}

func DeepCopyImpNative(native *openrtb2.Native) *openrtb2.Native {
	if native == nil {
		return nil
	}

	nativeCopy := *native
	nativeCopy.API = slices.Clone(native.API)
	nativeCopy.BAttr = slices.Clone(native.BAttr)
	nativeCopy.Ext = slices.Clone(native.Ext)
	return &nativeCopy
}

func DeepCopyImpBanner(banner *openrtb2.Banner) *openrtb2.Banner {
	if banner == nil {
		return nil
	}

	bannerCopy := *banner
	bannerCopy.Format = slices.Clone(banner.Format)
	bannerCopy.W = ptrutil.Clone(banner.W)
	bannerCopy.H = ptrutil.Clone(banner.H)
	bannerCopy.BType = slices.Clone(banner.BType)
	bannerCopy.BAttr = slices.Clone(banner.BAttr)
	bannerCopy.MIMEs = slices.Clone(banner.MIMEs)
	bannerCopy.ExpDir = slices.Clone(banner.ExpDir)
	bannerCopy.API = slices.Clone(banner.API)
	bannerCopy.Vcm = ptrutil.Clone(banner.Vcm)
	bannerCopy.Ext = slices.Clone(banner.Ext)
	return &bannerCopy
}

func DeepCopyImpAudio(audio *openrtb2.Audio) *openrtb2.Audio {
	if audio == nil {
		return nil
	}

	audioCopy := *audio
	audioCopy.MIMEs = slices.Clone(audio.MIMEs)
	audioCopy.Protocols = slices.Clone(audio.Protocols)
	audioCopy.StartDelay = ptrutil.Clone(audio.StartDelay)
	audioCopy.RqdDurs = slices.Clone(audio.RqdDurs)
	audioCopy.BAttr = slices.Clone(audio.BAttr)
	audioCopy.Delivery = slices.Clone(audio.Delivery)
	audioCopy.CompanionAd = slices.Clone(audio.CompanionAd)
	audioCopy.API = slices.Clone(audio.API)
	audioCopy.CompanionType = slices.Clone(audio.CompanionType)
	audioCopy.Stitched = ptrutil.Clone(audio.Stitched)
	audioCopy.NVol = ptrutil.Clone(audio.NVol)
	audioCopy.DurFloors = slices.Clone(audio.DurFloors)
	audioCopy.Ext = slices.Clone(audio.Ext)
	return &audioCopy
}

func DeepCopyImpPMP(pmp *openrtb2.PMP) *openrtb2.PMP {
	if pmp == nil {
		return nil
	}

	pmpCopy := *pmp
	pmpCopy.Deals = slices.Clone(pmp.Deals)
	pmpCopy.Ext = slices.Clone(pmp.Ext)
	return &pmpCopy
}
