package collector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deviceNames(devices []*api.Device) []string {
	names := make([]string, 0, len(devices))
	for _, d := range devices {
		names = append(names, d.DeviceName())
	}

	return names
}

func TestDeviceRegistry_Sync(t *testing.T) {
	mockDir := t.TempDir()
	apitest.SetMockRoot(t, mockDir)

	spec := apitest.StaticRNGDSpec()
	apitest.Build(t, apitest.MockSpec{Devices: spec.Devices[:7]})

	registry, err := NewDeviceRegistry()
	require.NoError(t, err)
	initial := registry.Devices()
	require.Len(t, initial, 7)

	// npu7 appears.
	apitest.Build(t, apitest.MockSpec{Devices: spec.Devices[7:]})

	added, removed, err := registry.Sync()
	require.NoError(t, err)
	assert.Equal(t, []string{"npu7"}, added)
	assert.Empty(t, removed)

	afterAdd := registry.Devices()
	require.Len(t, afterAdd, 8)
	assert.Equal(t, []string{"npu0", "npu1", "npu2", "npu3", "npu4", "npu5", "npu6", "npu7"}, deviceNames(afterAdd))
	for i := range initial {
		assert.Same(t, initial[i], afterAdd[i], "known devices must keep their pointers")
	}

	// Nothing changed.
	added, removed, err = registry.Sync()
	require.NoError(t, err)
	assert.Empty(t, added)
	assert.Empty(t, removed)
	for i, d := range registry.Devices() {
		assert.Same(t, afterAdd[i], d)
	}

	// npu7 disappears.
	require.NoError(t, os.RemoveAll(filepath.Join(mockDir, "sys/class/rngd_mgmt", "rngd!npu7mgmt")))

	listed, err := api.ListDevices()
	require.NoError(t, err)
	require.Len(t, listed, 7, "removing the management entry must make ListDevices drop npu7")

	added, removed, err = registry.Sync()
	require.NoError(t, err)
	assert.Empty(t, added)
	assert.Equal(t, []string{"npu7"}, removed)

	afterRemove := registry.Devices()
	require.Len(t, afterRemove, 7)
	for i, d := range afterRemove {
		assert.Same(t, initial[i], d, "surviving devices must keep their pointers")
		assert.True(t, d.Liveness())

		files, err := d.DeviceFiles()
		require.NoError(t, err)
		assert.NotEmpty(t, files)
	}
}

func TestDeviceRegistry_SyncRemovesAllOnEmptyResult(t *testing.T) {
	mockDir := t.TempDir()
	apitest.SetMockRoot(t, mockDir)
	apitest.Build(t, apitest.StaticRNGDSpec())

	registry, err := NewDeviceRegistry()
	require.NoError(t, err)
	before := registry.Devices()
	require.Len(t, before, apitest.StaticRNGDDeviceCount)

	// Removing every management entry leaves the PCI entries in place, so ListDevices
	// returns an empty slice rather than an error.
	require.NoError(t, os.RemoveAll(filepath.Join(mockDir, "sys/class/rngd_mgmt")))

	added, removed, err := registry.Sync()
	require.NoError(t, err)
	assert.Empty(t, added)
	assert.Equal(t, deviceNames(before), removed)
	assert.Empty(t, registry.Devices())
}

func TestNewDeviceRegistry_NoDevices(t *testing.T) {
	mockDir := t.TempDir()
	apitest.SetMockRoot(t, mockDir)
	apitest.Build(t, apitest.StaticRNGDSpec())

	// Removing every management entry leaves the PCI entries in place, so ListDevices
	// returns an empty slice rather than an error.
	require.NoError(t, os.RemoveAll(filepath.Join(mockDir, "sys/class/rngd_mgmt")))

	listed, err := api.ListDevices()
	require.NoError(t, err)
	require.Empty(t, listed)

	registry, err := NewDeviceRegistry()
	require.NoError(t, err)
	require.NotNil(t, registry)
	assert.Empty(t, registry.Devices())
}
