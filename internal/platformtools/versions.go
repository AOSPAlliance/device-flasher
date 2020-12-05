package platformtools

type SupportedHostOS string

const (
	OSDarwin  SupportedHostOS = "darwin"
	OSLinux   SupportedHostOS = "linux"
	OSWindows SupportedHostOS = "windows"
)

var SupportedHostOSes = []SupportedHostOS{OSDarwin, OSLinux, OSWindows}

type SupportedVersion string

const (
	Version_29_0_6 SupportedVersion = "29.0.6"
	Version_30_0_5 SupportedVersion = "30.0.5"
)

var SupportedVersions = []SupportedVersion{Version_29_0_6, Version_30_0_5}

type VersionInfo struct {
	Release     SupportedVersion
	TemplateURL string
	CheckSum    string
}

var Downloads = map[SupportedVersion]map[SupportedHostOS]VersionInfo{
	Version_29_0_6: {
		OSDarwin: VersionInfo{
			Version_29_0_6,
			"%v/platform-tools_r29.0.6-darwin.zip",
			"7555e8e24958cae4cfd197135950359b9fe8373d4862a03677f089d215119a3a"},
		OSLinux: VersionInfo{
			Version_29_0_6,
			"%v/platform-tools_r29.0.6-linux.zip",
			"cc9e9d0224d1a917bad71fe12d209dfffe9ce43395e048ab2f07dcfc21101d44"},
		OSWindows: VersionInfo{
			Version_29_0_6,
			"%v/platform-tools_r29.0.6-windows.zip",
			"247210e3c12453545f8e1f76e55de3559c03f2d785487b2e4ac00fe9698a039c"},
	},
	Version_30_0_5: {
		OSDarwin: VersionInfo{
			Version_30_0_5,
			"%v/eabcd8b4b7ab518c6af9c941af8494072f17ec4b.platform-tools_r30.0.5-darwin.zip",
			"e5780bad71a53cf9d693e1053a0748f49e4a67cc1f71d16a94ab4c943af3345f"},
		OSLinux: VersionInfo{
			Version_30_0_5,
			"%v/platform-tools_r30.0.5-linux.zip",
			"d6d72d006c03bd55d49b6cef9f00295db02f0a31da10e121427e1f4cb43e7cb9"},
		OSWindows: VersionInfo{
			Version_30_0_5,
			"%v/platform-tools_r30.0.5-windows.zip",
			"549ba2bdc31f335eb8a504f005f77606a479cc216d6b64a3e8b64c780003661f",
		},
	},
}
