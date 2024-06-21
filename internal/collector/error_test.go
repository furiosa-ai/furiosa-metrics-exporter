package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestErrorCollector_PostProcessing(t *testing.T) {
	e := &errorCollector{}
	e.Register()

	tc := MetricContainer{
		{
			arch:             "rngd",
			device:           "npu0",
			uuid:             "uuid",
			core:             "0-7",
			axiPostError:     float64(0),
			axiFetchError:    float64(0),
			axiDiscardError:  float64(0),
			axiDoorbellDone:  float64(0),
			pciePostError:    float64(0),
			pcieFetchError:   float64(0),
			pcieDiscardError: float64(0),
			pcieDoorbellDone: float64(0),
			deviceError:      float64(0),
		},
	}

	err := e.postProcess(tc)
	if err != nil {
		t.Errorf("unexpected error:%s\n", err)
	}

	expected := `
# HELP furiosa_npu_error The current active error counts of NPU device
# TYPE furiosa_npu_error gauge
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",label="axi_post_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",label="axi_fetch_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",label="axi_discard_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",label="axi_doorbell_done",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",label="pcie_post_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",label="pcie_fetch_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",label="pcie_discard_error",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",label="pcie_doorbell_done",uuid="uuid"} 0
furiosa_npu_error{arch="rngd",core="0-7",device="npu0",label="device_error",uuid="uuid"} 0
`
	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_error")
	if err != nil {
		t.Errorf("unexpected error:%s\n", err)
	}
}

func TestErrorCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
