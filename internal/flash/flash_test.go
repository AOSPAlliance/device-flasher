package flash

import (
	"errors"
	"github.com/aospalliance/device-flasher/internal/device"
	"github.com/aospalliance/device-flasher/internal/factoryimage"
	"github.com/aospalliance/device-flasher/internal/flash/mocks"
	"github.com/aospalliance/device-flasher/internal/platformtools"
	"github.com/aospalliance/device-flasher/internal/platformtools/fastboot"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFlash(t *testing.T) {
	ctrl := gomock.NewController(t)

	testADBDevice := &device.Device{
		ID:            "adbserial",
		Codename:      device.Codename("crosshatch"),
		DiscoveryTool: device.ADB,
		CustomHooks:   nil,
	}
	testDeviceFastboot := &device.Device{
		ID:            "fastbootserial",
		Codename:      device.Codename("crosshatch"),
		DiscoveryTool: device.Fastboot,
		CustomHooks:   nil,
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
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(nil),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().Reboot(testADBDevice.ID).Times(1).Return(nil),
				)
			},
			expectedErr: nil,
		},
		"device discovered through fastboot skips adb reboot": {
			device: testDeviceFastboot,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testDeviceFastboot.Codename).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testDeviceFastboot.ID).Return(fastboot.Locked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFastboot.EXPECT().UnlockBootloader(testDeviceFastboot.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testDeviceFastboot.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testDeviceFastboot, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(nil),
					mockFastboot.EXPECT().LockBootloader(testDeviceFastboot.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testDeviceFastboot.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().Reboot(testDeviceFastboot.ID).Times(1).Return(nil),
				)
			},
			expectedErr: nil,
		},
		"unlocked device skips unlocking step": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(nil),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().Reboot(testADBDevice.ID).Times(1).Return(nil),
				)
			},
			expectedErr: nil,
		},
		"factory image validation failure": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Return(factoryimage.ErrorValidation)
			},
			expectedErr: factoryimage.ErrorValidation,
		},
		"adb reboot bootloader error not fatal": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(errors.New("not fatal")),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(nil),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().Reboot(testADBDevice.ID).Times(1).Return(nil),
				)
			},
			expectedErr: nil,
		},
		"get bootloader status failure": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unknown, fastboot.ErrorCommandFailure),
				)
			},
			expectedErr: fastboot.ErrorCommandFailure,
		},
		"unlock bootloader failure": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(fastboot.ErrorUnlockBootloader),
				)
			},
			expectedErr: fastboot.ErrorUnlockBootloader,
		},
		"unlock bootloader tests for devices that immediately return - retry passes": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(nil),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().Reboot(testADBDevice.ID).Times(1).Return(nil),
				)
			},
			expectedErr: nil,
		},
		"unlock bootloader tests for devices that immediately return - max retry fails": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil).Times(1),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil).Times(1),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil).Times(1),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
				)
			},
			expectedErr: ErrMaxLockUnlockRetries,
		},
		"flash all error": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(factoryimage.ErrorFailedToFlash),
				)
			},
			expectedErr: factoryimage.ErrorFailedToFlash,
		},
		"lock bootloader failure": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(nil),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(fastboot.ErrorLockBootloader),
				)
			},
			expectedErr: fastboot.ErrorLockBootloader,
		},
		"lock bootloader tests for devices that immediately return - retry passes": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(nil),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().Reboot(testADBDevice.ID).Times(1).Return(nil),
				)
			},
			expectedErr: nil,
		},
		"lock bootloader tests for devices that immediately return - max retries fail": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(nil),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
				)
			},
			expectedErr: ErrMaxLockUnlockRetries,
		},
		"reboot error is not fatal": {
			device: testADBDevice,
			prepare: func(mockFactoryImage *mocks.MockFactoryImageFlasher, mockPlatformTools *mocks.MockPlatformToolsFlasher,
				mockADB *mocks.MockADBFlasher, mockFastboot *mocks.MockFastbootFlasher) {
				gomock.InOrder(
					mockFactoryImage.EXPECT().Validate(testADBDevice.Codename).Times(1).Return(nil),
					mockADB.EXPECT().RebootIntoBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFastboot.EXPECT().UnlockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Unlocked, nil).Times(1),
					mockPlatformTools.EXPECT().Path().AnyTimes().Return(platformtools.PlatformToolsPath("/tmp")),
					mockFactoryImage.EXPECT().FlashAll(testADBDevice, platformtools.PlatformToolsPath("/tmp")).Times(1).Return(nil),
					mockFastboot.EXPECT().LockBootloader(testADBDevice.ID).Times(1).Return(nil),
					mockFastboot.EXPECT().GetBootloaderLockStatus(testADBDevice.ID).Return(fastboot.Locked, nil).Times(1),
					mockFastboot.EXPECT().Reboot(testADBDevice.ID).Times(1).Return(fastboot.ErrorRebootFailure),
				)
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

			tc.prepare(mockFactoryImage, mockPlatformTools, mockADB, mockFastboot)

			logger := logrus.StandardLogger()
			logger.SetLevel(logrus.DebugLevel)
			flash := New(&Config{
				HostOS:                    "TestOS",
				FactoryImage:              mockFactoryImage,
				PlatformTools:             mockPlatformTools,
				ADB:                       mockADB,
				Fastboot:                  mockFastboot,
				Logger:                    logger,
				LockUnlockValidationPause: time.Millisecond,
				LockUnlockRetries:         DefaultLockUnlockRetries,
				LockUnlockRetryInterval:   time.Millisecond,
			})

			err := flash.Flash(tc.device)
			if tc.expectedErr == nil {
				assert.Nil(t, err)
			} else {
				assert.True(t, errors.Is(err, tc.expectedErr))
			}
		})
	}
}
