package collector

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

type MetricFactory interface {
	NewDeviceWiseMetric(d *api.Device) (Metric, error)
	// SyncDriverVersion re-reads the driver version and reports whether it changed.
	// On error the previously read value is kept (empty until the first success).
	SyncDriverVersion() (changed bool, err error)
	// DriverVersion returns the last successfully read driver version, or "".
	DriverVersion() string
}

var _ MetricFactory = (*metricFactory)(nil)

func NewMetricFactory(nodeName string) MetricFactory {
	return &metricFactory{
		nodeName: nodeName,
	}
}

// metricFactory builds the label set shared by every device-wise metric. The driver version is a
// node-wide value that the exporter loop reads once per collect interval via SyncDriverVersion,
// not per metric, because api.DriverInfo enumerates every device.
type metricFactory struct {
	nodeName string

	mu            sync.Mutex
	driverVersion string
}

func (m *metricFactory) SyncDriverVersion() (bool, error) {
	info, err := api.DriverInfo()
	if err != nil {
		return false, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	version := info.String()
	changed := version != m.driverVersion
	m.driverVersion = version

	return changed, nil
}

func (m *metricFactory) DriverVersion() string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.driverVersion
}

func (m *metricFactory) NewDeviceWiseMetric(d *api.Device) (Metric, error) {
	metric := newMetric()
	info, err := getDeviceInfo(d)
	if err != nil {
		return nil, err
	}

	metric[arch] = info.arch
	metric[core] = info.coreLabel
	metric[device] = info.device
	metric[uuid] = info.uuid
	metric[bdf] = info.bdf
	metric[firmwareVersion] = info.firmwareVersion
	metric[driverVersion] = m.DriverVersion()
	metric[hostname] = m.nodeName

	return metric, nil
}

type deviceInfo struct {
	arch            string
	device          string
	uuid            string
	cores           []uint32
	coreLabel       string
	bdf             string
	firmwareVersion string
}

func getDeviceInfo(device *api.Device) (*deviceInfo, error) {
	firmware, err := device.FirmwareVersion()
	if err != nil {
		return nil, err
	}

	files, err := device.DeviceFiles()
	if err != nil {
		return nil, err
	}

	accumulatedCores := map[uint32]uint32{}
	for _, file := range files {
		for _, c := range file.Cores() {
			accumulatedCores[c] = c
		}
	}

	cores := make([]uint32, 0, len(accumulatedCores))
	for c := range accumulatedCores {
		cores = append(cores, c)
	}

	start := slices.Min(cores)
	end := slices.Max(cores)

	var coreLabel string
	if start == end {
		coreLabel = fmt.Sprintf("%d", start)
	} else {
		coreLabel = fmt.Sprintf("%d-%d", start, end)
	}

	return &deviceInfo{
		arch:            archName(device),
		device:          device.DeviceName(),
		uuid:            device.UUID(),
		cores:           cores,
		coreLabel:       coreLabel,
		bdf:             device.BDF().String(),
		firmwareVersion: firmware.String(),
	}, nil
}

// archName returns the lowercase architecture name (e.g. "rngd") used as the arch label.
func archName(device *api.Device) string {
	return strings.ToLower(device.Arch().String())
}

func defaultMetricLabels() []string {
	return []string{
		arch,
		core,
		device,
		uuid,
		bdf,
		firmwareVersion,
		driverVersion,
		hostname,
		kubernetesNamespace,
		kubernetesPod,
		kubernetesContainer,
	}
}

func newMetric() Metric {
	labels := defaultMetricLabels()

	metric := make(Metric, len(labels))
	for _, l := range labels {
		metric[l] = ""
	}

	return metric
}

func deepCopyMetric(origin Metric) Metric {
	dst := make(Metric, len(origin))
	// NOTE(@bg): At the moment, we don't use slices or maps for the metric value
	// So, it is safe to do a shallow copy. If we start using slices or maps, we need to do a recursive deep copy.
	maps.Copy(dst, origin)
	return dst
}
