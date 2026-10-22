package collector

import (
	"os"
	"testing"
	"time"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api/apitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoreUtilizationSampler_GetCoreUtilization(t *testing.T) {
	staticDevices := apitest.StaticRNGDDevices(t)
	devices := make([]*api.Device, 0, len(staticDevices))
	for i := range staticDevices {
		devices = append(devices, &staticDevices[i])
	}

	ctx := t.Context()

	sampler := NewCoreUtilizationSampler(ctx, StaticDeviceSource(devices), time.Hour)

	for _, d := range devices {
		utilization, err := sampler.GetCoreUtilization(d)
		require.NoError(t, err)
		require.Len(t, utilization, 8)

		for i, u := range utilization {
			assert.Equal(t, uint32(i), u.Core)
			assert.GreaterOrEqual(t, u.PeUsagePercentage, float64(0))
			assert.LessOrEqual(t, u.PeUsagePercentage, float64(100))
			assert.GreaterOrEqual(t, u.TimeWindow, time.Duration(0))
		}
	}
}

func TestCoreUtilizationSampler_UnknownDevice(t *testing.T) {
	sampler := &coreUtilizationSampler{samples: map[*api.Device]performanceCounterSample{}}

	_, err := sampler.GetCoreUtilization(nil)
	assert.Error(t, err)

	d := apitest.StaticRNGDDevice(t, 0)
	_, err = sampler.GetCoreUtilization(&d)
	assert.Error(t, err)
}

// TestCoreUtilizationSampler_ReadFailure verifies that a failed performance counter read is surfaced by
// GetCoreUtilization instead of silently serving the previous sample, and that the sampler recovers once reads
// succeed again.
func TestCoreUtilizationSampler_ReadFailure(t *testing.T) {
	dir := t.TempDir()
	apitest.SetMockRoot(t, dir)
	apitest.Build(t, apitest.StaticRNGDSpec())

	listed, err := api.ListDevices()
	require.NoError(t, err)
	require.NotEmpty(t, listed)

	devices := make([]*api.Device, 0, len(listed))
	for i := range listed {
		devices = append(devices, &listed[i])
	}
	d := devices[0]

	ctx := t.Context()

	s := NewCoreUtilizationSampler(ctx, StaticDeviceSource(devices), time.Hour).(*coreUtilizationSampler)

	utilization, err := s.GetCoreUtilization(d)
	require.NoError(t, err)
	require.Len(t, utilization, 8)

	// Inject a read failure by removing the mock sysfs/devfs roots.
	require.NoError(t, os.RemoveAll(dir))
	_, err = d.PerformanceCounter()
	require.Error(t, err, "mock device must fail to read once the mock root is gone")

	s.updateUtilization(devices)

	_, err = s.GetCoreUtilization(d)
	require.Error(t, err)

	s.lock.RLock()
	sample := s.samples[d]
	s.lock.RUnlock()

	assert.ErrorIs(t, err, sample.err)
	assert.NotNil(t, sample.after, "last good counters must be retained across a failed read")

	var smiErr api.SMIError
	require.ErrorAs(t, err, &smiErr)
	assert.Equal(t, api.ErrorKindDeviceNotAvailable, smiErr.Kind())

	// Recover: rebuild the fixture under the same root and sample again.
	apitest.Build(t, apitest.StaticRNGDSpec())

	s.updateUtilization(devices)

	utilization, err = s.GetCoreUtilization(d)
	require.NoError(t, err)
	require.Len(t, utilization, 8)

	s.lock.RLock()
	sample = s.samples[d]
	s.lock.RUnlock()
	assert.NoError(t, sample.err)
}

// TestCoreUtilizationSampler_FirstReadFails verifies that a device whose very first read fails reports the error,
// and that its first successful read is stored as a fresh sample rather than paired with missing counters.
func TestCoreUtilizationSampler_FirstReadFails(t *testing.T) {
	dir := t.TempDir()
	apitest.SetMockRoot(t, dir)
	apitest.Build(t, apitest.StaticRNGDSpec())

	listed, err := api.ListDevices()
	require.NoError(t, err)
	require.NotEmpty(t, listed)

	devices := []*api.Device{&listed[0]}
	d := devices[0]

	require.NoError(t, os.RemoveAll(dir))

	s := &coreUtilizationSampler{samples: map[*api.Device]performanceCounterSample{}}
	s.updateUtilization(devices)

	_, err = s.GetCoreUtilization(d)
	require.Error(t, err)

	apitest.Build(t, apitest.StaticRNGDSpec())

	s.updateUtilization(devices)

	utilization, err := s.GetCoreUtilization(d)
	require.NoError(t, err)
	require.Len(t, utilization, 8)

	s.lock.RLock()
	sample := s.samples[d]
	s.lock.RUnlock()
	assert.NoError(t, sample.err)
	assert.NotNil(t, sample.before)
	assert.NotNil(t, sample.after)
	for c := range sample.before {
		assert.Equal(t, sample.before[c], sample.after[c], "first successful read must be stored as before == after")
	}
}

func TestCoreUtilizationSampler_DropsRemovedDevices(t *testing.T) {
	staticDevices := apitest.StaticRNGDDevices(t)
	devices := make([]*api.Device, 0, len(staticDevices))
	for i := range staticDevices {
		devices = append(devices, &staticDevices[i])
	}

	s := &coreUtilizationSampler{samples: map[*api.Device]performanceCounterSample{}}
	s.updateUtilization(devices)
	require.Len(t, s.samples, len(devices))

	s.updateUtilization(devices[:7])
	assert.Len(t, s.samples, 7)

	_, err := s.GetCoreUtilization(devices[7])
	assert.Error(t, err)
}
