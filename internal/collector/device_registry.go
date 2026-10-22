package collector

import (
	"sync"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

// DeviceSource provides the devices to collect metrics from.
type DeviceSource interface {
	// Devices returns the currently known devices. Callers must not modify the returned slice.
	Devices() []*api.Device
}

// StaticDeviceSource is a DeviceSource over a fixed set of devices.
type StaticDeviceSource []*api.Device

var _ DeviceSource = StaticDeviceSource(nil)

func (s StaticDeviceSource) Devices() []*api.Device {
	return s
}

// DeviceRegistry is a DeviceSource that re-enumerates devices on Sync.
type DeviceRegistry struct {
	mu      sync.RWMutex
	devices []*api.Device
}

var _ DeviceSource = (*DeviceRegistry)(nil)

// NewDeviceRegistry enumerates the devices once and returns a registry over them.
// It fails only when the enumeration fails; an empty result is a valid, empty registry.
func NewDeviceRegistry() (*DeviceRegistry, error) {
	r := &DeviceRegistry{}

	if _, _, err := r.Sync(); err != nil {
		return nil, err
	}

	return r, nil
}

// Devices returns a copy of the currently known devices in api.ListDevices order.
func (r *DeviceRegistry) Devices() []*api.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	devices := make([]*api.Device, len(r.devices))
	copy(devices, r.devices)

	return devices
}

// Sync re-enumerates devices and replaces the known set, returning the names of the added and removed devices.
// A device that is still present keeps its pointer, as other components key their state by *api.Device.
// On an enumeration error the known set is kept and the error is returned; an empty result removes every device.
func (r *DeviceRegistry) Sync() (added, removed []string, err error) {
	listed, err := api.ListDevices()
	if err != nil {
		return nil, nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	known := make(map[string]*api.Device, len(r.devices))
	for _, d := range r.devices {
		known[deviceKey(d)] = d
	}

	devices := make([]*api.Device, 0, len(listed))
	for i := range listed {
		key := deviceKey(&listed[i])

		if d, ok := known[key]; ok {
			devices = append(devices, d)
			delete(known, key)
			continue
		}

		devices = append(devices, &listed[i])
		added = append(added, listed[i].DeviceName())
	}

	for _, d := range r.devices {
		if _, ok := known[deviceKey(d)]; ok {
			removed = append(removed, d.DeviceName())
		}
	}

	r.devices = devices

	return added, removed, nil
}

func deviceKey(d *api.Device) string {
	return d.DeviceName() + "|" + d.UUID() + "|" + d.BDF().String()
}
