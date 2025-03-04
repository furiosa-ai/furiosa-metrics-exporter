package collector

import (
	"errors"
	"fmt"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/kubernetes"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	peUtilization = "peUtilization"
)

type coreUtilizationCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
	nodeName string
}

var _ Collector = (*coreUtilizationCollector)(nil)

func NewCoreUtilizationCollector(devices []smi.Device, nodeName string) Collector {
	return &coreUtilizationCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *coreUtilizationCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_core_utilization",
		Help: "The current core utilization of NPU device",
	},
		[]string{
			arch,
			device,
			core,
			kubernetesNodeName,
			uuid,
			pod,
		})
}

func (t *coreUtilizationCollector) Collect(devicePodMap map[string][]kubernetes.PodInfo) error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		coreUtilization, err := d.CoreUtilization()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		utilization := coreUtilization.PeUtilization()
		for _, pe := range utilization {

			podName := ""

			if podInfos, ok := devicePodMap[info.uuid]; ok {
				for _, podInfo := range podInfos {
					for _, core := range podInfo.AllocatedPE {
						if core == int(pe.Core()) {
							podName = podInfo.Name
							break
						}
					}
				}
			}

			metric := Metric{
				arch:               info.arch,
				core:               fmt.Sprintf("%d", pe.Core()),
				device:             info.device,
				kubernetesNodeName: t.nodeName,
				uuid:               info.uuid,
				peUtilization:      pe.PeUsagePercentage(),
				pod:                podName,
			}
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

func (t *coreUtilizationCollector) postProcess(metrics MetricContainer) error {
	t.gaugeVec.Reset()

	for _, metric := range metrics {
		if value, ok := metric[peUtilization]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				uuid:               metric[uuid].(string),
				pod:                metric[pod].(string),
			}).Set(value.(float64))
		}
	}

	return nil
}
