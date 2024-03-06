package adunitconfig

import (
	"github.com/prebid/openrtb/v19/openrtb2"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models"
	"github.com/prebid/prebid-server/v2/modules/pubmatic/openwrap/models/adunitconfig"
)

func ReplaceAppObjectFromAdUnitConfig(rCtx models.RequestCtx, app *openrtb2.App) {
	if app == nil {
		return
	}

	var adUnitCfg *adunitconfig.AdConfig
	for _, impCtx := range rCtx.ImpBidCtx {
		if impCtx.BannerAdUnitCtx.AppliedSlotAdUnitConfig != nil {
			adUnitCfg = impCtx.BannerAdUnitCtx.AppliedSlotAdUnitConfig
			break
		}
		if impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig != nil {
			adUnitCfg = impCtx.VideoAdUnitCtx.AppliedSlotAdUnitConfig
			break
		}
	}

	if adUnitCfg == nil || adUnitCfg.App == nil {
		return
	}

	if app.ID == "" {
		app.ID = adUnitCfg.App.ID
	}

	if app.Name == "" {
		app.Name = adUnitCfg.App.Name
	}

	if app.Bundle == "" {
		app.Bundle = adUnitCfg.App.Bundle
	}

	if app.Domain == "" {
		app.Domain = adUnitCfg.App.Domain
	}

	if app.StoreURL == "" {
		app.StoreURL = adUnitCfg.App.StoreURL
	}

	if len(app.Cat) == 0 {
		app.Cat = adUnitCfg.App.Cat
	}

	if len(app.SectionCat) == 0 {
		app.SectionCat = adUnitCfg.App.SectionCat
	}

	if len(app.PageCat) == 0 {
		app.PageCat = adUnitCfg.App.PageCat
	}

	if app.Ver == "" {
		app.Ver = adUnitCfg.App.Ver
	}

	if app.PrivacyPolicy == 0 {
		app.PrivacyPolicy = adUnitCfg.App.PrivacyPolicy
	}

	if app.Paid == 0 {
		app.Paid = adUnitCfg.App.Paid
	}

	if app.Content == nil {
		app.Content = adUnitCfg.App.Content
	}

	if app.Keywords == "" {
		app.Keywords = adUnitCfg.App.Keywords
	}

	if app.Ext == nil {
		app.Ext = adUnitCfg.App.Ext
	}

	if adUnitCfg.App.Publisher != nil {
		if app.Publisher == nil {
			app.Publisher = &openrtb2.Publisher{}
		}

		if app.Publisher.Name == "" {
			app.Publisher.Name = adUnitCfg.App.Publisher.Name
		}

		if len(app.Publisher.Cat) == 0 {
			app.Publisher.Cat = adUnitCfg.App.Publisher.Cat
		}

		if app.Publisher.Domain == "" {
			app.Publisher.Domain = adUnitCfg.App.Publisher.Domain
		}

		if app.Publisher.Ext == nil {
			app.Publisher.Ext = adUnitCfg.App.Publisher.Ext
		}
	}

}
