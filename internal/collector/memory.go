package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	dramUsage = "dramUsage"
	dramTotal  = "dramTotal"
)

type memoryCollector struct {
	devices       []smi.Device
	metricFactory MetricFactory
	kubeResMapper KubeResourcesMapper

	dramUsageGaugeVec *prometheus.GaugeVec
	dramTotalGaugeVec  *prometheus.GaugeVec
}

var _ Collector = (*memoryCollector)(nil)

func NewMemoryCollector(devices []smi.Device, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &memoryCollector{
		devices:       devices,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *memoryCollector) GetDevices() []smi.Device {
	return t.devices
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
		Help: "The total dram of NPU device",
	}

	t.dramTotalGaugeVec = prometheus.NewGaugeVec(dramTotalOpts, defaultMetricLabels())
	prometheus.MustRegister(NewLabelFilterCollector(
		t.dramTotalGaugeVec,
		prometheus.Opts(dramTotalOpts),
		prometheus.GaugeValue,
	))
}

func (t *memoryCollector) Collect(metrics map[smi.Device]Metric) error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric, exists := metrics[d]
		if !exists {
			continue
		}

		memUtil, err := d.MemoryUtilization()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		dram := memUtil.Dram().Memory()
		dramShared := memUtil.DramShared().Memory()

		dramUsageBytes := uint64(0)
		dramTotalBytes := uint64(0)

		for _, mem := range dram {
			dramUsageBytes += mem.InUseBytes()
			dramTotalBytes += mem.TotalBytes()
		}
		for _, mem := range dramShared {
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
