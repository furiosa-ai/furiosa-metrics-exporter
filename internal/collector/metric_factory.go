package collector

import (
	"fmt"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"slices"
)

type MetricFactory interface {
	NewDeviceWiseMetric(d smi.Device) (Metric, error)
}

var _ MetricFactory = (*metricFactory)(nil)

func NewMetricFactory(nodeName, driverVersion string) MetricFactory {
	return &metricFactory{
		nodeName:      nodeName,
		driverVersion: driverVersion,
	}
}

type metricFactory struct {
	nodeName      string
	driverVersion string
}

func (m *metricFactory) NewDeviceWiseMetric(d smi.Device) (Metric, error) {
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
	metric[pertVersion] = info.pertVersion
	metric[driverVersion] = m.driverVersion
	metric[kubernetesNode] = m.nodeName

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
	pertVersion     string
}

func getDeviceInfo(device smi.Device) (*deviceInfo, error) {
	info, err := device.DeviceInfo()
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
		arch:            info.Arch().ToString(),
		device:          info.Name(),
		uuid:            info.UUID(),
		cores:           cores,
		coreLabel:       coreLabel,
		bdf:             info.BDF(),
		firmwareVersion: info.FirmwareVersion().String(),
		pertVersion:     info.PertVersion().String(),
	}, nil
}

func defaultMetricLabels() []string {
	return []string{
		arch,
		core,
		device,
		uuid,
		bdf,
		firmwareVersion,
		pertVersion,
		driverVersion,
		kubernetesNode,
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
	for k, v := range origin {
		// NOTE(@bg): At the moment, we don't use slices or maps for the metric value
		// So, it is safe to do a shallow copy. If we start using slices or maps, we need to do a recursive deep copy.
		dst[k] = v
	}
	return dst
}
