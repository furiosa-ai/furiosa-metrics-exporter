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

// TestMetricFactorySyncDriverVersion must not call t.Parallel(): it installs a process-global mock root.
func TestMetricFactorySyncDriverVersion(t *testing.T) {
	mockDir := t.TempDir()
	apitest.SetMockRoot(t, mockDir)

	factory := NewMetricFactory("node")

	// Nothing is built yet, so DriverInfo fails and the version stays empty.
	changed, err := factory.SyncDriverVersion()
	require.Error(t, err)
	assert.False(t, changed)
	assert.Equal(t, "", factory.DriverVersion())

	apitest.Build(t, apitest.StaticRNGDSpec())

	info, err := api.DriverInfo()
	require.NoError(t, err)
	first := info.String()

	changed, err = factory.SyncDriverVersion()
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, first, factory.DriverVersion())

	changed, err = factory.SyncDriverVersion()
	require.NoError(t, err)
	assert.False(t, changed, "an unchanged version must not be reported as a change")

	// A driver upgrade rewrites the version attribute; the next sync must pick it up.
	versionFiles, err := filepath.Glob(filepath.Join(mockDir, "sys/class/rngd_mgmt", "rngd!npu*mgmt", "version"))
	require.NoError(t, err)
	require.Len(t, versionFiles, apitest.StaticRNGDDeviceCount)
	for _, p := range versionFiles {
		require.NoError(t, os.WriteFile(p, []byte("2025.9.9+abcdef0\n"), 0o644))
	}

	info, err = api.DriverInfo()
	require.NoError(t, err)
	upgraded := info.String()
	require.NotEqual(t, first, upgraded)

	changed, err = factory.SyncDriverVersion()
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, upgraded, factory.DriverVersion())

	// Metrics built after the sync carry the new version.
	devices, err := api.ListDevices()
	require.NoError(t, err)
	require.NotEmpty(t, devices)

	metric, err := factory.NewDeviceWiseMetric(&devices[0])
	require.NoError(t, err)
	assert.Equal(t, upgraded, metric[driverVersion])

	// With no visible device DriverInfo fails and the last known value is kept.
	require.NoError(t, os.RemoveAll(filepath.Join(mockDir, "sys/class/rngd_mgmt")))

	changed, err = factory.SyncDriverVersion()
	require.Error(t, err)
	assert.False(t, changed)
	assert.Equal(t, upgraded, factory.DriverVersion())
}
