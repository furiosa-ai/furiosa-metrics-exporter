package collector

//func TestTotalCycleCountCollector_PostProcessing(t *testing.T) {
//	tests := []struct {
//		description string
//		source      MetricContainer
//		expected    string
//	}{
//		{
//			description: "random performance counter metrics",
//			source: func() MetricContainer {
//				tc := MetricContainer{}
//				for i := 0; i < 8; i++ {
//					metric := newMetric()
//					metric[arch] = "rngd"
//					metric[core] = strconv.Itoa(i)
//					metric[device] = "npu0"
//					metric[uuid] = "uuid"
//					metric[bdf] = "bdf"
//					metric[totalCycleCount] = float64(5678)
//					tc = append(tc, metric)
//				}
//				return tc
//			}(),
//			expected: `
//# HELP furiosa_npu_total_cycle_count The current total cycle count of NPU device
//# TYPE furiosa_npu_total_cycle_count counter
//furiosa_npu_total_cycle_count{arch="rngd",container="",core="0",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 5678
//furiosa_npu_total_cycle_count{arch="rngd",container="",core="1",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 5678
//furiosa_npu_total_cycle_count{arch="rngd",container="",core="2",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 5678
//furiosa_npu_total_cycle_count{arch="rngd",container="",core="3",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 5678
//furiosa_npu_total_cycle_count{arch="rngd",container="",core="4",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 5678
//furiosa_npu_total_cycle_count{arch="rngd",container="",core="5",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 5678
//furiosa_npu_total_cycle_count{arch="rngd",container="",core="6",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 5678
//furiosa_npu_total_cycle_count{arch="rngd",container="",core="7",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 5678
//`,
//		},
//	}
//
//	cc := &totalCycleCountCollector{}
//	cc.Register()
//	for _, tc := range tests {
//		t.Run(tc.description, func(t *testing.T) {
//			err := cc.postProcess(tc.source)
//			assert.NoError(t, err)
//
//			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_total_cycle_count")
//			assert.NoError(t, err)
//		})
//	}
//}
//
//func TestTotalCycleCountCollector_Collect(t *testing.T) {
//	//TODO: add test cases with mock device data
//}
