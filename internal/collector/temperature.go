package collector

import (
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type temperatureCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
}

const (
	ambient = "ambient"
	peak    = "peak"
)

var _ Collector = (*temperatureCollector)(nil)

func NewTemperatureCollector(devices []smi.Device) Collector {
	return &temperatureCollector{
		devices: devices,
	}
}

func (t *temperatureCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_temperature",
		Help: "The current temperature of NPU device",
	},
		[]string{
			arch,
			core,
			device,
			kubernetesNodeName,
			label,
			uuid,
		})
}

func (t *temperatureCollector) Collect() error {
	var metricContainer MetricContainer

	for _, d := range t.devices {
		metric := Metric{}

		info, err := getDeviceInfo(d)
		if err != nil {
			return err
		}

		metric[arch] = info.arch
		metric[device] = info.device
		metric[uuid] = info.uuid
		metric[core] = info.core
		metric[kubernetesNodeName] = info.node

		deviceTemperature, err := d.DeviceTemperature()
		if err != nil {
			return err
		}

		metric[ambient] = deviceTemperature.Ambient()
		metric[peak] = deviceTemperature.SocPeak()
		metricContainer = append(metricContainer, metric)
	}

	return t.postProcess(metricContainer)
}

func (t *temperatureCollector) postProcess(metrics MetricContainer) error {
	for _, metric := range metrics {
		if value, ok := metric[ambient]; ok {

			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				label:              ambient,
				uuid:               metric[uuid].(string),
			}).Set(value.(float64))
		}

		if value, ok := metric[peak]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				label:              peak,
				uuid:               metric[uuid].(string),
			}).Set(value.(float64))
		}
	}

	return nil
}
