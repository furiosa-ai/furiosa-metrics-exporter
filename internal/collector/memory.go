package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	dramUsage = "dramUsage"
	dramTotal = "dramTotal"
)

type memoryCollector struct {
	deviceSource  DeviceSource
	metricFactory MetricFactory
	kubeResMapper KubeResourcesMapper

	dramUsageGaugeVec *prometheus.GaugeVec
	dramTotalGaugeVec *prometheus.GaugeVec
}

var _ Collector = (*memoryCollector)(nil)

func NewMemoryCollector(deviceSource DeviceSource, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &memoryCollector{
		deviceSource:  deviceSource,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *memoryCollector) GetDevices() []*api.Device {
	return t.deviceSource.Devices()
}

func (t *memoryCollector) GetMetricFactory() MetricFactory {
	return t.metricFactory
}

func (t *memoryCollector) Register() {
	dramUsageOpts := prometheus.GaugeOpts{
		Name: "furiosa_npu_dram_usage",
		Help: "The current used dram of NPU device (Bytes)",
	}

	t.dramUsageGaugeVec = prometheus.NewGaugeVec(dramUsageOpts, defaultMetricLabels())
	prometheus.MustRegister(NewLabelFilterCollector(
		t.dramUsageGaugeVec,
		prometheus.Opts(dramUsageOpts),
		prometheus.GaugeValue,
	))

	dramTotalOpts := prometheus.GaugeOpts{
		Name: "furiosa_npu_dram_total",
		Help: "The total dram of NPU device (Bytes)",
	}

	t.dramTotalGaugeVec = prometheus.NewGaugeVec(dramTotalOpts, defaultMetricLabels())
	prometheus.MustRegister(NewLabelFilterCollector(
		t.dramTotalGaugeVec,
		prometheus.Opts(dramTotalOpts),
		prometheus.GaugeValue,
	))
}

func (t *memoryCollector) Collect(metrics map[*api.Device]Metric) error {
	devices := t.GetDevices()
	metricContainer := make(MetricContainer, 0, len(devices))

	errs := make([]error, 0)
	for _, d := range devices {
		metric, exists := metrics[d]
		if !exists {
			continue
		}

		memoryUsage, err := d.MemoryUsage()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		dramUsageBytes := uint64(0)
		dramTotalBytes := uint64(0)

		// Each entry is a disjoint PE fusion group that already includes its share of the shared DRAM.
		for _, mem := range memoryUsage {
			dramUsageBytes += mem.InUseBytes()
			dramTotalBytes += mem.TotalBytes()
		}

		metric[dramUsage] = dramUsageBytes
		metric[dramTotal] = dramTotalBytes
		metricContainer = append(metricContainer, metric)
	}

	if err := t.postProcess(metricContainer); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (t *memoryCollector) postProcess(metrics MetricContainer) error {
	transformed := t.kubeResMapper.TransformDeviceMetrics(metrics, false)
	t.dramUsageGaugeVec.Reset()
	t.dramTotalGaugeVec.Reset()

	for _, metric := range transformed {
		if value, ok := metric[dramUsage]; ok {
			t.dramUsageGaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				uuid:                metric[uuid].(string),
				bdf:                 metric[bdf].(string),
				firmwareVersion:     metric[firmwareVersion].(string),
				driverVersion:       metric[driverVersion].(string),
				hostname:            metric[hostname].(string),
				kubernetesNamespace: metric[kubernetesNamespace].(string),
				kubernetesPod:       metric[kubernetesPod].(string),
				kubernetesContainer: metric[kubernetesContainer].(string),
			}).Set(float64(value.(uint64)))
		}

		if value, ok := metric[dramTotal]; ok {
			t.dramTotalGaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				uuid:                metric[uuid].(string),
				bdf:                 metric[bdf].(string),
				firmwareVersion:     metric[firmwareVersion].(string),
				driverVersion:       metric[driverVersion].(string),
				hostname:            metric[hostname].(string),
				kubernetesNamespace: metric[kubernetesNamespace].(string),
				kubernetesPod:       metric[kubernetesPod].(string),
				kubernetesContainer: metric[kubernetesContainer].(string),
			}).Set(float64(value.(uint64)))
		}
	}

	return nil
}
