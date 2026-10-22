package collector

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

// CoreUtilization is the utilization of a single PE core over a time window.
type CoreUtilization struct {
	Core              uint32
	TimeWindow        time.Duration
	PeUsagePercentage float64
}

// CoreUtilizationSampler provides per-core utilization computed from periodic performance counter samples.
type CoreUtilizationSampler interface {
	GetCoreUtilization(d *api.Device) ([]CoreUtilization, error)
}

type performanceCounterSample struct {
	before map[uint32]api.PePerformanceCounter
	after  map[uint32]api.PePerformanceCounter
	// err is the error from the most recent update; nil when after is current.
	err error
}

type coreUtilizationSampler struct {
	lock    sync.RWMutex
	samples map[*api.Device]performanceCounterSample
}

var _ CoreUtilizationSampler = (*coreUtilizationSampler)(nil)

// NewCoreUtilizationSampler takes two initial samples 100ms apart, then keeps sampling the current devices of deviceSource every interval until ctx is done.
// Samples of devices that are no longer in deviceSource are dropped.
// A failed read keeps the previous counters but makes GetCoreUtilization return the error until the next successful read.
func NewCoreUtilizationSampler(ctx context.Context, deviceSource DeviceSource, interval time.Duration) CoreUtilizationSampler {
	devices := deviceSource.Devices()

	s := &coreUtilizationSampler{
		samples: make(map[*api.Device]performanceCounterSample, len(devices)),
	}

	s.updateUtilization(devices)
	time.Sleep(100 * time.Millisecond)
	s.updateUtilization(devices)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.updateUtilization(deviceSource.Devices())

			case <-ctx.Done():
				return
			}
		}
	}()

	return s
}

// updateUtilization samples every device in devices and drops the samples of devices that are not in devices.
func (s *coreUtilizationSampler) updateUtilization(devices []*api.Device) {
	s.lock.Lock()
	for d := range s.samples {
		if !slices.Contains(devices, d) {
			delete(s.samples, d)
		}
	}
	s.lock.Unlock()

	for _, d := range devices {
		counters, err := d.PerformanceCounter()

		s.lock.Lock()
		prev := s.samples[d]

		switch {
		case err != nil:
			// Keep the last good pair, but flag it so GetCoreUtilization reports the failure.
			prev.err = err
			s.samples[d] = prev

		case prev.after == nil:
			// First successful read for this device.
			s.samples[d] = performanceCounterSample{before: counters, after: counters}

		default:
			s.samples[d] = performanceCounterSample{before: prev.after, after: counters}
		}
		s.lock.Unlock()
	}
}

func (s *coreUtilizationSampler) GetCoreUtilization(d *api.Device) ([]CoreUtilization, error) {
	if d == nil {
		return nil, fmt.Errorf("device is nil")
	}

	s.lock.RLock()
	sample, exists := s.samples[d]
	s.lock.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no performance counter sample found for device %s", d.DeviceName())
	}

	if sample.err != nil {
		return nil, fmt.Errorf("latest performance counter read failed for device %s: %w", d.DeviceName(), sample.err)
	}

	cores := make([]uint32, 0, len(sample.before))
	for c := range sample.before {
		cores = append(cores, c)
	}
	slices.Sort(cores)

	result := make([]CoreUtilization, 0, len(cores))
	for _, c := range cores {
		before := sample.before[c]
		after, ok := sample.after[c]
		if !ok {
			continue
		}

		result = append(result, CoreUtilization{
			Core:              c,
			TimeWindow:        after.DurationSince(before),
			PeUsagePercentage: after.UtilizationSince(before) * 100,
		})
	}

	return result, nil
}
