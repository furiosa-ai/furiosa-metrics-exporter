package collector

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCoreUtilizationCollector_PostProcessing(t *testing.T) {
	cu := &coreUtilizationCollector{}
	cu.Register()

	tc := MetricContainer{}
	for i := 0; i < 8; i++ {
		tc = append(tc, Metric{
			arch:               "rngd",
			core:               strconv.Itoa(i),
			device:             "npu0",
			kubernetesNodeName: "node",
			uuid:               "uuid",
			peUtilization:      float64(rand.Intn(100)),
		})
	}

	err := cu.postProcess(tc)
	if err != nil {
		t.Errorf("unexpected error:%s\n", err)
	}

	expected := func() string {
		stringBuilder := strings.Builder{}
		stringBuilder.WriteString("# HELP furiosa_npu_core_utilization The current core utilization of NPU device\n")
		stringBuilder.WriteString("# TYPE furiosa_npu_core_utilization gauge\n")

		for i := range tc {
			stringBuilder.WriteString(
				fmt.Sprintf("furiosa_npu_core_utilization{arch=\"rngd\",core=\"%s\",device=\"npu0\",kubernetes_node_name=\"node\",uuid=\"uuid\"} %f\n", tc[i][core], tc[i][peUtilization]),
			)
		}

		return stringBuilder.String()
	}()

	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_core_utilization")
	if err != nil {
		t.Errorf("unexpected error:%s\n", err)
	}
}
