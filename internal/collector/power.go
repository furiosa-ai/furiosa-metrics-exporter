package collector

import (
	"errors"
	"strings"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/kubernetes"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	rms = "rms"
)

type powerCollector struct {
	devices         []smi.Device
	gaugeVec        *prometheus.GaugeVec
	gaugeVecWithPod *prometheus.GaugeVec
	nodeName        string
}

var _ Collector = (*powerCollector)(nil)

func NewPowerCollector(devices []smi.Device, nodeName string) Collector {
	return &powerCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *powerCollector) Register(registryWithPod *prometheus.Registry) {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_power",
		Help: "The current power of NPU device",
	},
		[]string{
			arch,
			device,
			label,
			core,
			kubernetesNodeName,
			uuid,
		})

	t.gaugeVecWithPod = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_power",
		Help: "The current power of NPU device",
	},
		[]string{
			arch,
			device,
			label,
			core,
			kubernetesNodeName,
			uuid,
			pod,
		})
	registryWithPod.MustRegister(t.gaugeVecWithPod)
}

func (t *powerCollector) Collect(devicePodMap map[string]kubernetes.PodInfo) error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric := Metric{}

		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		isPodExists := false
		for deviceID := range devicePodMap {
			if strings.Contains(deviceID, info.uuid) {
				isPodExists = true
				break
			}
		}

		metric[arch] = info.arch
		metric[device] = info.device
		metric[uuid] = info.uuid
		metric[core] = info.coreLabel
		metric[kubernetesNodeName] = t.nodeName
		if isPodExists {
			metric[pod] = devicePodMap[info.uuid].Name
		}

		power, err := d.PowerConsumption()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		metric[rms] = power
		metricContainer = append(metricContainer, metric)
	}

	if err := t.postProcess(metricContainer, devicePodMap); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (t *powerCollector) postProcess(metrics MetricContainer, devicePodMap map[string]kubernetes.PodInfo) error {
	t.gaugeVec.Reset()
	t.gaugeVecWithPod.Reset()

	for _, metric := range metrics {
		if value, ok := metric["rms"]; ok {
			for deviceID, podInfo := range devicePodMap {
				if strings.Contains(deviceID, metric[uuid].(string)) {
					coreNum := kubernetes.GetCoreNum(deviceID)

					t.gaugeVecWithPod.With(prometheus.Labels{
						arch:               metric[arch].(string),
						core:               coreNum,
						device:             metric[device].(string),
						kubernetesNodeName: metric[kubernetesNodeName].(string),
						label:              rms,
						uuid:               metric[uuid].(string),
						pod:                podInfo.Name,
					}).Set(value.(float64))
				}
			}
			if _, ok := devicePodMap[metric[uuid].(string)]; !ok {
				t.gaugeVec.With(prometheus.Labels{
					arch:               metric[arch].(string),
					core:               metric[core].(string),
					device:             metric[device].(string),
					kubernetesNodeName: metric[kubernetesNodeName].(string),
					label:              rms,
					uuid:               metric[uuid].(string),
				}).Set(value.(float64))
			}
		}
	}

	return nil
}
