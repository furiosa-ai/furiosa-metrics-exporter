package collector

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

type deviceInfo struct {
	arch   string
	device string
	uuid   string
	core   string
	node   string
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

	uniqueCores := make([]uint32, 0, len(accumulatedCores))
	for core := range accumulatedCores {
		uniqueCores = append(uniqueCores, core)
	}

	start := slices.Min(uniqueCores)
	end := slices.Max(uniqueCores)

	var core string
	if start == end {
		core = fmt.Sprintf("%d", start)
	} else {
		core = fmt.Sprintf("%d-%d", start, end)
	}

	nodeName := os.Getenv("NODE_NAME")

	return &deviceInfo{
		arch:   info.Arch().ToString(),
		device: filepath.Base(info.Name()),
		uuid:   info.UUID(),
		core:   core,
		node:   nodeName,
	}, nil
}
