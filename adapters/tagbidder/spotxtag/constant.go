package spotxtag

const (
	spotxURL                 = `https://search.spotxchange.com/vast/2.00/85394?VPI=MP4&app[bundle]=[REPLACE_ME]&app[name]=[REPLACE_ME]&app[cat]=[REPLACE_ME]&app[domain]=[REPLACE_ME]&app[privacypolicy]=[REPLACE_ME]&app[storeurl]=[REPLACE_ME]&app[ver]=[REPLACE_ME]&cb=[REPLACE_ME]&device[devicetype]=[REPLACE_ME]&device[ifa]=[REPLACE_ME]&device[make]=[REPLACE_ME]&device[model]=[REPLACE_ME]&device[dnt]=[REPLACE_ME]&player_height=[REPLACE_ME]&player_width=[REPLACE_ME]&ip_addr=[REPLACE_ME]&device[ua]=[REPLACE_ME]]&schain=[REPLACE_ME]`
	spotxURLWithMacros       = `https://search.spotxchange.com/vast/2.00/%%Channel%%?VPI=MP4&app[bundle]=%%MacroAppBundle%%&app[name]=%%MacroAppName%%&app[cat]=%%MacroCategory%%&app[domain]=%%MacroDomain%%&app[privacypolicy]=%%MacroPrivacyPolicy%%&app[storeurl]=%%MacroAppStoreURL%%&app[ver]=%%MacroAppVersion%%&cb=%%MacroCacheBuster%%&device[devicetype]=%%MacroDeviceType%%&device[ifa]=%%MacroDeviceIFA%%&device[make]=%%MacroMake%%&device[model]=%%MacroModel%%&device[dnt]=%%MacroDNT%%&player_height=%%MacroVideoPlayerHeight%%&player_width=%%MacroVideoPlayerWidth%%&ip_addr=%%MacroIP%%&device[ua]=%%MacroUserAgent%%`
)
