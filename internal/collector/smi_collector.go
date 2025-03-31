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

func SyncDeviceSMICache(devices []smi.Device) [][]error {
	results := make([][]error, len(devices))

	for i, d := range devices {
		errors := make([]error, 0)

		deviceInfo, err := d.DeviceInfo()
		if err != nil {
			errors = append(errors, err)
		}

		deviceFiles, err := d.DeviceFiles()
		if err != nil {
			errors = append(errors, err)
		}

		coreUtilization, err := d.CoreUtilization()
		if err != nil {
			errors = append(errors, err)
		}

		coreFrequency, err := d.CoreFrequency()
		if err != nil {
			errors = append(errors, err)
		}

		liveness, err := d.Liveness()
		if err != nil {
			errors = append(errors, err)
		}

		performanceCounter, err := d.DevicePerformanceCounter()
		if err != nil {
			errors = append(errors, err)
		}

		temperature, err := d.DeviceTemperature()
		if err != nil {
			errors = append(errors, err)
		}

		power, err := d.PowerConsumption()
		if err != nil {
			errors = append(errors, err)
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

		results[i] = errors
	}

	return results
}
