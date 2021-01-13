// +build en

package strings

func init() {
	Title = Vendor + " Installer"

	Back = "Back"
	Done = "Done"
	Finish = "Finish"
	Flash = "Flash"
	Log = "Log"
	Next = "Next"
	Retry = "Retry"
	Save = "Save"
	Select = "Select"
	View = "View"

	HeadsUp = "Hey, heads up..."

	DeviceInstructions = "Please follow the instructions on the device"
	DeviceRebootRequired = "Your device will reboot!\n\nWhen your phone reboots you'll need to complete the following steps:\n1. Disconnect the cable and power the device off\n2. Press and hold the volume down and power keys to boot the device into fastboot mode\n3. Reconnect the cable to your device"
	LockBootloader = DeviceInstructions + " to lock the bootloader"
	UnlockBootloader = DeviceInstructions + " to unlock the bootloader"
	SetupPlatformTools = "Setting up platform tools..."

	DeveloperModeHeader = "Enable Developer Mode"
	DeveloperModeBody = "1. Go to Settings > About Phone\n2. Tap \"Build number\" 7 times"
	InstallHeader = "Installing " + Vendor
	InstallBody = "* Do not interact with your device unless asked to\n* Do not unplug your device"
	OemLockHeader = "Re-enable OEM lock"
	OemLockBody = "1. Go to Settings > System > Advanced > Developer Options\n2. Tap the toggle labelled \"OEM Unlocking\" to disable it"
	OemUnlockingHeader = "Enable OEM Unlocking"
	OemUnlockingBody = "1. Go to Settings > System > Advanced > Developer Options\n2. Tap the toggle labelled \"OEM Unlocking\" to enable it\n3. Press \"Enable\" on the \"Allow OEM unlocking?\" prompt"
	PlugDeviceHeader = "Connect to Your Computer"
	PlugDeviceBody = "1. Plug the device into the computer\n2. Press \"Allow\" on the \"Allow USB debugging?\" prompt"
	PrepareDeviceHeader = "Prepare Your Device"
	PrepareDeviceBody = "1. Connect to a wifi network\n2. Remove your SIM card"
	SelectHeader = "Select the " + Vendor + " image"
	SuccessHeader = "You've successfully installed\n" + Vendor + "!"
	SuccessBody = "It's now safe to unplug your device"
	UsbDebuggingHeader = "Enable USB debugging"
	UsbDebuggingBody = "1. Go to Settings > System > Advanced > Developer Options\n2. Tap the toggle labelled \"USB debugging\" to enable it\n3. Press \"OK\" on the \"Allow USB debugging?\" prompt"
}
