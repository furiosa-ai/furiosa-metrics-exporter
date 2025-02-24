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
	liveness = "liveness"
)

type livenessCollector struct {
	devices         []smi.Device
	gaugeVec        *prometheus.GaugeVec
	gaugeVecWithPod *prometheus.GaugeVec
	nodeName        string
}

var _ Collector = (*livenessCollector)(nil)

func NewLivenessCollector(devices []smi.Device, nodeName string) Collector {
	return &livenessCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *livenessCollector) Register(registryWithPod *prometheus.Registry) {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_alive",
		Help: "The liveness of NPU device",
	},
		[]string{
			arch,
			core,
			device,
			kubernetesNodeName,
			uuid,
		})

	t.gaugeVecWithPod = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_alive",
		Help: "The liveness of NPU device",
	},
		[]string{
			arch,
			core,
			device,
			kubernetesNodeName,
			uuid,
			pod,
		})
	registryWithPod.MustRegister(t.gaugeVecWithPod)
}

func (t *livenessCollector) Collect(devicePodMap map[string]kubernetes.PodInfo) error {
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
		metric[core] = info.coreLabel
		metric[device] = info.device
		metric[uuid] = info.uuid
		metric[kubernetesNodeName] = t.nodeName
		if isPodExists {
			metric[pod] = devicePodMap[info.uuid].Name
		}

		value, err := d.Liveness()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		metric[liveness] = value
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

func (t *livenessCollector) postProcess(metrics MetricContainer, devicePodMap map[string]kubernetes.PodInfo) error {
	t.gaugeVec.Reset()
	t.gaugeVecWithPod.Reset()

	for _, metric := range metrics {
		if value, ok := metric[liveness]; ok {
			var alive float64
			if value.(bool) {
				alive = 1
			} else {
				alive = 0
			}
			for deviceID, podInfo := range devicePodMap {

				if strings.Contains(deviceID, metric[uuid].(string)) {
					coreNum := kubernetes.GetCoreNum(deviceID)

					t.gaugeVecWithPod.With(prometheus.Labels{
						arch:               metric[arch].(string),
						core:               coreNum,
						device:             metric[device].(string),
						uuid:               metric[uuid].(string),
						kubernetesNodeName: metric[kubernetesNodeName].(string),
						pod:                podInfo.Name,
					}).Set(alive)
				}
			}
			if _, ok := devicePodMap[metric[uuid].(string)]; !ok {
				t.gaugeVec.With(prometheus.Labels{
					arch:               metric[arch].(string),
					core:               metric[core].(string),
					device:             metric[device].(string),
					uuid:               metric[uuid].(string),
					kubernetesNodeName: metric[kubernetesNodeName].(string),
				}).Set(alive)
			}
		}
	}

	return nil
}
