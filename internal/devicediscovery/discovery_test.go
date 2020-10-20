package devicediscovery

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gitlab.com/calyxos/device-flasher/internal/device"
	"gitlab.com/calyxos/device-flasher/internal/devicediscovery/mocks"
	"testing"
)

func TestDiscoverDevices(t *testing.T) {
	ctrl := gomock.NewController(t)

	testDeviceADB := &device.Device{ID: "serialadb", Codename: device.Codename("adb"), DiscoveryTool: device.ADB}
	testDuplicateFastboot := &device.Device{ID: "serialadb", Codename: device.Codename("fastboot"), DiscoveryTool: device.Fastboot}
	testDeviceFastboot := &device.Device{ID: "serialfastboot", Codename: device.Codename("fastboot"), DiscoveryTool: device.Fastboot}

	tests := map[string]struct {
		device          *device.Device
		prepare         func(*mocks.MockDeviceDiscoverer, *mocks.MockDeviceDiscoverer)
		expectedErr     error
		expectedDevices map[string]*device.Device
	}{
		"discovery successful with adb device": {
			prepare: func(mockADB *mocks.MockDeviceDiscoverer, mockFastboot *mocks.MockDeviceDiscoverer) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{testDeviceADB.ID}, nil)
				mockADB.EXPECT().GetDeviceCodename(testDeviceADB.ID).Return(string(testDeviceADB.Codename), nil)
				mockFastboot.EXPECT().GetDeviceIds().Return(nil, nil)
				mockADB.EXPECT().Name().Return(string(device.ADB))
				mockFastboot.EXPECT().Name().Return(string(device.Fastboot))
			},
			expectedErr:     nil,
			expectedDevices: map[string]*device.Device{testDeviceADB.ID: testDeviceADB},
		},
		"discovery successful with fastboot device": {
			prepare: func(mockADB *mocks.MockDeviceDiscoverer, mockFastboot *mocks.MockDeviceDiscoverer) {
				mockADB.EXPECT().GetDeviceIds().Return(nil, nil)
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{testDeviceFastboot.ID}, nil)
				mockFastboot.EXPECT().GetDeviceCodename(testDeviceFastboot.ID).Return(string(testDeviceFastboot.Codename), nil)
				mockADB.EXPECT().Name().Return(string(device.ADB))
				mockFastboot.EXPECT().Name().Return(string(device.Fastboot))
			},
			expectedErr:     nil,
			expectedDevices: map[string]*device.Device{testDeviceFastboot.ID: testDeviceFastboot},
		},
		"discovery successful with both adb and fastboot devices": {
			prepare: func(mockADB *mocks.MockDeviceDiscoverer, mockFastboot *mocks.MockDeviceDiscoverer) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{testDeviceADB.ID}, nil)
				mockADB.EXPECT().GetDeviceCodename(testDeviceADB.ID).Return(string(testDeviceADB.Codename), nil)
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{testDeviceFastboot.ID}, nil)
				mockFastboot.EXPECT().GetDeviceCodename(testDeviceFastboot.ID).Return(string(testDeviceFastboot.Codename), nil)
				mockADB.EXPECT().Name().Return(string(device.ADB))
				mockFastboot.EXPECT().Name().Return(string(device.Fastboot))
			},
			expectedErr: nil,
			expectedDevices: map[string]*device.Device{
				testDeviceADB.ID:      testDeviceADB,
				testDeviceFastboot.ID: testDeviceFastboot,
			},
		},
		"discovery fails when get device returns empty for both adb and fastboot": {
			prepare: func(mockADB *mocks.MockDeviceDiscoverer, mockFastboot *mocks.MockDeviceDiscoverer) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{}, nil)
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{}, nil)
				mockADB.EXPECT().Name().Return(string(device.ADB))
				mockFastboot.EXPECT().Name().Return(string(device.Fastboot))
			},
			expectedErr:     ErrNoDevicesFound,
			expectedDevices: nil,
		},
		"discovery fails when get device fails for both adb and fastboot": {
			prepare: func(mockADB *mocks.MockDeviceDiscoverer, mockFastboot *mocks.MockDeviceDiscoverer) {
				mockADB.EXPECT().GetDeviceIds().Return(nil, errors.New("failed"))
				mockFastboot.EXPECT().GetDeviceIds().Return(nil, errors.New("failed"))
				mockADB.EXPECT().Name().Return(string(device.ADB))
				mockFastboot.EXPECT().Name().Return(string(device.Fastboot))
			},
			expectedErr:     ErrNoDevicesFound,
			expectedDevices: nil,
		},
		"duplicate fastboot device overwrites existing adb device": {
			prepare: func(mockADB *mocks.MockDeviceDiscoverer, mockFastboot *mocks.MockDeviceDiscoverer) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{testDeviceADB.ID}, nil)
				mockADB.EXPECT().GetDeviceCodename(testDeviceADB.ID).Return(string(testDeviceADB.Codename), nil)
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{testDuplicateFastboot.ID}, nil)
				mockFastboot.EXPECT().GetDeviceCodename(testDuplicateFastboot.ID).Return(string(testDuplicateFastboot.Codename), nil)
				mockADB.EXPECT().Name().Return(string(device.ADB))
				mockFastboot.EXPECT().Name().Return(string(device.Fastboot))
			},
			expectedErr:     nil,
			expectedDevices: map[string]*device.Device{testDuplicateFastboot.ID: testDuplicateFastboot},
		},
		"device in not added if get codename fails": {
			prepare: func(mockADB *mocks.MockDeviceDiscoverer, mockFastboot *mocks.MockDeviceDiscoverer) {
				mockADB.EXPECT().GetDeviceIds().Return([]string{testDeviceADB.ID}, nil)
				mockADB.EXPECT().GetDeviceCodename(testDeviceADB.ID).Return("", errors.New("fail"))
				mockFastboot.EXPECT().GetDeviceIds().Return([]string{testDeviceFastboot.ID}, nil)
				mockFastboot.EXPECT().GetDeviceCodename(testDeviceFastboot.ID).Return("", errors.New("fail"))
				mockADB.EXPECT().Name().Return(string(device.ADB))
				mockFastboot.EXPECT().Name().Return(string(device.Fastboot))
			},
			expectedErr:     ErrNoDevicesFound,
			expectedDevices: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mockADB := mocks.NewMockDeviceDiscoverer(ctrl)
			mockFastboot := mocks.NewMockDeviceDiscoverer(ctrl)

			if tc.prepare != nil {
				tc.prepare(mockADB, mockFastboot)
			}

			discovery := New(mockADB, mockFastboot, logrus.StandardLogger())
			devices, err := discovery.DiscoverDevices()
			assert.Equal(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedDevices, devices)
		})
	}
}
