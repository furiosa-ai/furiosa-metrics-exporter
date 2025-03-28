package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPowerCollector_PostProcessing(t *testing.T) {
	p := &powerCollector{}
	p.Register()

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

	err := p.postProcess(tc)
	assert.NoError(t, err)

	expected := `
# HELP furiosa_npu_hw_power The current power of NPU device
# TYPE furiosa_npu_hw_power gauge
furiosa_npu_hw_power{arch="rngd",core="0-7",device="npu0",driver_version="",firmware_version="",kubernetes_container_name="",kubernetes_namespace_name="",kubernetes_node_name="",kubernetes_pod_name="",label="rms",pci_bus_id="bdf",pert_version="",uuid="uuid"} 4795000
`

	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_hw_power")
	assert.NoError(t, err)
}

func TestPowerCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
