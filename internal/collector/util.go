package collector

import (
	"fmt"
	"slices"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

type deviceInfo struct {
	arch      string
	device    string
	uuid      string
	cores     []uint32
	coreLabel string
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
		arch:      info.Arch().ToString(),
		device:    info.Name(),
		uuid:      info.UUID(),
		cores:     cores,
		coreLabel: core,
	}, nil
}

func newMetric() Metric {
	return Metric{
		arch:                "",
		device:              "",
		core:                "",
		uuid:                "",
		kubernetesNode:      "",
		kubernetesNamespace: "",
		kubernetesPod:       "",
		kubernetesContainer: "",
	}
}
