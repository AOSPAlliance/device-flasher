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
	Version_30_0_4 SupportedVersion = "30.0.4"
)

var SupportedVersions = []SupportedVersion{Version_29_0_6, Version_30_0_4}

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
	Version_30_0_4: {
		OSDarwin: VersionInfo{
			Version_30_0_4,
			"%v/fbad467867e935dce68a0296b00e6d1e76f15b15.platform-tools_r30.0.4-darwin.zip",
			"e0db2bdc784c41847f854d6608e91597ebc3cef66686f647125f5a046068a890"},
		OSLinux: VersionInfo{
			Version_30_0_4,
			"%v/platform-tools_r30.0.4-linux.zip",
			"5be24ed897c7e061ba800bfa7b9ebb4b0f8958cc062f4b2202701e02f2725891"},
		OSWindows: VersionInfo{
			Version_30_0_4,
			"%v/platform-tools_r30.0.4-windows.zip",
			"413182fff6c5957911e231b9e97e6be4fc6a539035e3dfb580b5c54bd5950fee",
		},
	},
}
