package collector

import (
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

var DeviceSMICacheMap = make(map[string]deviceSMICache)

type deviceSMICache struct {
	deviceInfo         smi.DeviceInfo
	deviceFiles        []smi.DeviceFile
	coreUtilization    smi.CoreUtilization
	coreFrequency      smi.CoreFrequency
	liveness           bool
	performanceCounter smi.DevicePerformanceCounter
	temperature        smi.DeviceTemperature
	power              float64
}

func SyncDeviceSMICache(devices []smi.Device) {
	for _, d := range devices {
		deviceInfo, err := d.DeviceInfo()
		if err != nil {
			continue
		}

		deviceFiles, err := d.DeviceFiles()
		if err != nil {
			continue
		}

		coreUtilization, err := d.CoreUtilization()
		if err != nil {
			continue
		}

		coreFrequency, err := d.CoreFrequency()
		if err != nil {
			continue
		}

		liveness, err := d.Liveness()
		if err != nil {
			continue
		}

		performanceCounter, err := d.DevicePerformanceCounter()
		if err != nil {
			continue
		}

		temperature, err := d.DeviceTemperature()
		if err != nil {
			continue
		}

		power, err := d.PowerConsumption()
		if err != nil {
			continue
		}

		DeviceSMICacheMap[deviceInfo.Name()] = deviceSMICache{
			deviceInfo:         deviceInfo,
			deviceFiles:        deviceFiles,
			coreUtilization:    coreUtilization,
			coreFrequency:      coreFrequency,
			liveness:           liveness,
			performanceCounter: performanceCounter,
			temperature:        temperature,
			power:              power,
		}
	}
}
