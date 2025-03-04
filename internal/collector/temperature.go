package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/kubernetes"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type temperatureCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
	nodeName string
}

const (
	ambient = "ambient"
	peak    = "peak"
)

var _ Collector = (*temperatureCollector)(nil)

func NewTemperatureCollector(devices []smi.Device, nodeName string) Collector {
	return &temperatureCollector{
		devices:  devices,
		nodeName: nodeName,
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
			pod,
		})
}

func (t *temperatureCollector) Collect(devicePodMap map[string][]kubernetes.PodInfo) error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric := Metric{}

		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		deviceTemperature, err := d.DeviceTemperature()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if podInfos, ok := devicePodMap[info.uuid]; ok {
			if !(len(podInfos) == 1 && len(podInfos[0].AllocatedPE) == 8) {
				metric[arch] = info.arch
				metric[device] = info.device
				metric[uuid] = info.uuid
				metric[core] = info.coreLabel
				metric[kubernetesNodeName] = t.nodeName
				metric[pod] = ""
				metric[ambient] = deviceTemperature.Ambient()
				metric[peak] = deviceTemperature.SocPeak()

				metricContainer = append(metricContainer, metric)
			}

			for _, podInfo := range podInfos {
				metric := Metric{}
				metric[arch] = info.arch
				metric[core] = podInfo.CoreLabel
				metric[device] = info.device
				metric[uuid] = info.uuid
				metric[kubernetesNodeName] = t.nodeName
				metric[pod] = podInfo.Name
				metric[ambient] = deviceTemperature.Ambient()
				metric[peak] = deviceTemperature.SocPeak()

				metricContainer = append(metricContainer, metric)
			}
		} else {
			metric[arch] = info.arch
			metric[device] = info.device
			metric[uuid] = info.uuid
			metric[core] = info.coreLabel
			metric[kubernetesNodeName] = t.nodeName
			metric[pod] = ""
			metric[ambient] = deviceTemperature.Ambient()
			metric[peak] = deviceTemperature.SocPeak()

			metricContainer = append(metricContainer, metric)
		}
	}

	if err := t.postProcess(metricContainer); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (t *temperatureCollector) postProcess(metrics MetricContainer) error {
	t.gaugeVec.Reset()

	for _, metric := range metrics {
		if value, ok := metric[ambient]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				label:              ambient,
				uuid:               metric[uuid].(string),
				pod:                metric[pod].(string),
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
				pod:                metric[pod].(string),
			}).Set(value.(float64))
		}
	}

	return nil
}
