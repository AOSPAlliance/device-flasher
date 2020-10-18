package flash

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gitlab.com/calyxos/device-flasher/internal/device"
	"gitlab.com/calyxos/device-flasher/internal/factoryimage"
	"gitlab.com/calyxos/device-flasher/internal/flash/mocks"
	"gitlab.com/calyxos/device-flasher/internal/platformtools"
	"gitlab.com/calyxos/device-flasher/internal/platformtools/fastboot"
	"testing"
)

func TestFlash(t *testing.T) {
	ctrl := gomock.NewController(t)

	testADBDevice := &device.Device{
		ID:            "adbserial",
		Codename:      device.Codename("crosshatch"),
		DiscoveryTool: device.ADB,
	}
	testDeviceFastboot := &device.Device{
		ID:            "fastbootserial",
		Codename:      device.Codename("crosshatch"),
		DiscoveryTool: device.Fastboot,
	}

	tests := map[string]struct {
		device  *device.Device
		prepare func(*mocks.MockFactoryImageFlasher, *mocks.MockPlatformToolsFlasher,
			*mocks.MockADBFlasher, *mocks.MockFastbootFlasher)
		expectedErr error
	}{
		"happy path flash successful": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testADBDevice.ID).Return(nil)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
		"device discovered through fastboot skips adb reboot": {
			device: testDeviceFastboot,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testDeviceFastboot.Codename).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testDeviceFastboot.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDeviceFastboot.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(testDeviceFastboot, platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testDeviceFastboot.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testDeviceFastboot.ID).Return(nil)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
		"unlocked device skips unlocking step": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testADBDevice.ID).Return(nil)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
		"factory image validation failure": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(factoryimage.ErrorValidation)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: factoryimage.ErrorValidation,
		},
		"adb reboot bootloader error not fatal": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Return(errors.New("not fatal"))
				mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testADBDevice.ID).Return(nil)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
		"get bootloader status failure": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unknown, fastboot.ErrorCommandFailure)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: fastboot.ErrorCommandFailure,
		},
		"set bootloader status unlock failure": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Unlocked).Return(fastboot.ErrorUnlockBootloader)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: fastboot.ErrorUnlockBootloader,
		},
		"flash all error": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Return(factoryimage.ErrorFailedToFlash)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: factoryimage.ErrorFailedToFlash,
		},
		"set bootloader status lock failure": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Locked).Return(fastboot.ErrorLockBootloader)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: fastboot.ErrorLockBootloader,
		},
		"reboot error is not fatal": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(nil)
				mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Return(nil)
				mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Unlocked).Return(nil)
				mockPlatformTools.EXPECT().Path().Return(platformtools.PlatformToolsPath("/tmp"))
				mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Return(nil)
				mockFastboot.EXPECT().SetBootloaderLockStatus(testADBDevice.ID, fastboot.Locked).Return(nil)
				mockFastboot.EXPECT().Reboot(testADBDevice.ID).Return(fastboot.ErrorRebootFailure)
				mockADB.EXPECT().KillServer()
			},
			expectedErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockFactoryImage := mocks.NewMockFactoryImageFlasher(ctrl)
			mockPlatformTools := mocks.NewMockPlatformToolsFlasher(ctrl)
			mockADB := mocks.NewMockADBFlasher(ctrl)
			mockFastboot := mocks.NewMockFastbootFlasher(ctrl)

			if tc.prepare != nil {
				tc.prepare(mockFactoryImage, mockPlatformTools, mockADB, mockFastboot)
			}

			flash := New(&Config{
				HostOS:        "TestOS",
				FactoryImage:  mockFactoryImage,
				PlatformTools: mockPlatformTools,
				ADB:           mockADB,
				Fastboot:      mockFastboot,
				Logger:        logrus.StandardLogger(),
			})

			err := flash.Flash(tc.device)
			if tc.expectedErr == nil {
				assert.Nil(t, err)
			} else {
				assert.True(t, errors.Is(err, tc.expectedErr), true)
			}
		})
	}
}
