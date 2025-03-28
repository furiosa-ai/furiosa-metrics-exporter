package collector

import (
	"fmt"
	"slices"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

type deviceInfo struct {
	arch            string
	device          string
	uuid            string
	cores           []uint32
	coreLabel       string
	bdf             string
	firmwareVersion string
	pertVersion     string
	driverVersion   string
}

func getDeviceInfo(device smi.Device) (*deviceInfo, error) {
	info, err := device.DeviceInfo()
	if err != nil {
		return nil, err
	}

	driverInfo, err := smi.DriverInfo()
	if err != nil {
		return nil, err
	}

	files, err := device.DeviceFiles()
	if err != nil {
		return nil, err
	}

	accumulatedCores := map[uint32]uint32{}
	for _, file := range files {
		for _, core := range file.Cores() {
			accumulatedCores[core] = core
		}
	}

	cores := make([]uint32, 0, len(accumulatedCores))
	for core := range accumulatedCores {
		cores = append(cores, core)
	}

	start := slices.Min(cores)
	end := slices.Max(cores)

	var core string
	if start == end {
		core = fmt.Sprintf("%d", start)
	} else {
		core = fmt.Sprintf("%d-%d", start, end)
	}

	return &deviceInfo{
		arch:            info.Arch().ToString(),
		device:          info.Name(),
		uuid:            info.UUID(),
		cores:           cores,
		coreLabel:       core,
		bdf:             info.BDF(),
		firmwareVersion: info.FirmwareVersion().String(),
		pertVersion:     info.PertVersion().String(),
		driverVersion:   driverInfo.String(),
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

func newDeviceWiseMetric(d smi.Device) (Metric, error) {
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
	metric[driverVersion] = info.driverVersion

	return metric, nil
}

func deepCopyMetric(origin Metric) Metric {
	dst := make(Metric, len(origin))
	for k, v := range origin {
		dst[k] = v
	}
	return dst
}
