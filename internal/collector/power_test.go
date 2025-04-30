package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func newFakePowerCollector() Collector {
	return &powerCollector{
		devices:       nil,
		metricFactory: nil,
		kubeResMapper: NewFakeKubeResourcesMapper(),
	}
}

func TestPowerCollector_PostProcessing(t *testing.T) {
	collector := newFakePowerCollector()
	collector.Register()

	tc := MetricContainer{}
	metric := newMetric()
	metric[arch] = "rngd"
	metric[core] = "0-7"
	metric[device] = "npu0"
	metric[uuid] = "uuid"
	metric[bdf] = "bdf"
	metric[label] = rms
	metric[rms] = float64(4795000)

	tc = append(tc, metric)

	err := collector.postProcess(tc)
	assert.NoError(t, err)

	expected := `
# HELP furiosa_npu_hw_power The current power of NPU device
# TYPE furiosa_npu_hw_power gauge
furiosa_npu_hw_power{arch="rngd",core="0-7",device="npu0",label="rms",pci_bus_id="bdf",uuid="uuid"} 4795000
`

	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_hw_power")
	assert.NoError(t, err)
}

func TestPowerCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
