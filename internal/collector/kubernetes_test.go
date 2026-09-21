package collector

import (
	"context"
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	podResourcesAPI "k8s.io/kubelet/pkg/apis/podresources/v1"
)

func TestBuildMultiWiseCacheFromPodResources(t *testing.T) {
	deviceWise, coreWise := buildMultiWiseCacheFromPodResources(testPodResourcesResponse())
	emptyDevices, emptyCores := buildMultiWiseCacheFromPodResources(nil)
	assert.Empty(t, emptyDevices)
	assert.Empty(t, emptyCores)
	assert.Len(t, deviceWise["uuid-plugin"], 1)
	assert.Len(t, deviceWise["uuid-mixed"], 1)
	assert.NotContains(t, deviceWise, "uuid-ignored")

	assert.Len(t, deviceWise["npu.furiosa.ai/npu-dra"], 1)
	assert.Len(t, deviceWise["npu.furiosa.ai/npu-mixed"], 1)
	assert.Len(t, deviceWise["npu.furiosa.ai/npu-shared"], 2)
	assert.Len(t, deviceWise["npu.furiosa.ai/npu-multi"], 2)
	assert.NotContains(t, deviceWise, "npu.furiosa.ai/npu-driver-ignored")

	for _, coreIdx := range getAllocatedPE() {
		assert.Len(t, coreWise["npu.furiosa.ai/npu-dra"][coreIdx], 1)
		assert.Len(t, coreWise["npu.furiosa.ai/npu-shared"][coreIdx], 2)
	}
}

func TestKubeResourcesMapper_TransformDeviceMetrics_DeviceWise(t *testing.T) {
	deviceWise, coreWise := buildMultiWiseCacheFromPodResources(testPodResourcesResponse())
	mapper := &kubeResourcesMapper{
		enabled:         true,
		deviceWiseCache: deviceWise,
		coreWiseCache:   coreWise,
	}

	metrics := MetricContainer{
		newMetricWithIdentity("uuid-plugin", "npu-plugin", "0-7"),
		newMetricWithIdentity("uuid-unmatched", "npu-dra", "0-7"),
		newMetricWithIdentity("uuid-mixed", "npu-mixed", "0-7"),
		newMetricWithIdentity("uuid-unmatched", "npu-shared", "0-7"),
		newMetricWithIdentity("uuid-none", "npu-none", "0-7"),
	}
	snapshot := cloneMetricContainer(metrics)

	transformed := mapper.TransformDeviceMetrics(metrics, false)

	assert.Equal(t, snapshot, metrics)
	if !assert.Len(t, transformed, 6) {
		return
	}
	assert.Equal(t, []string{
		metricPodKey("ns-plugin", "plugin-pod", "plugin-container", "0-7"),
		metricPodKey("ns-dra", "dra-pod", "dra-container", "0-7"),
		metricPodKey("ns-mixed", "mixed-pod", "mixed-container", "0-7"),
		metricPodKey("ns-shared-a", "shared-pod-a", "shared-container-a", "0-7"),
		metricPodKey("ns-shared-b", "shared-pod-b", "shared-container-b", "0-7"),
		metricPodKey("", "", "", "0-7"),
	}, metricPodKeys(transformed))
	assert.Equal(t, "uuid-none", transformed[5][uuid])
	assert.Equal(t, "npu-none", transformed[5][device])
}

func TestKubeResourcesMapper_TransformDeviceMetrics_CoreWise(t *testing.T) {
	deviceWise, coreWise := buildMultiWiseCacheFromPodResources(testPodResourcesResponse())
	mapper := &kubeResourcesMapper{
		enabled:         true,
		deviceWiseCache: deviceWise,
		coreWiseCache:   coreWise,
	}

	metrics := make(MetricContainer, 0, 11)
	for _, coreIdx := range getAllocatedPE() {
		metrics = append(metrics, newMetricWithIdentity("uuid-unmatched", "npu-dra", strconv.Itoa(coreIdx)))
	}
	metrics = append(metrics,
		newMetricWithIdentity("uuid-unmatched", "npu-shared", "3"),
		newMetricWithIdentity("uuid-none", "npu-none", "4"),
		newMetricWithIdentity("uuid-plugin", "npu-plugin", "invalid"),
	)
	snapshot := cloneMetricContainer(metrics)

	transformed := mapper.TransformDeviceMetrics(metrics, true)

	assert.Equal(t, snapshot, metrics)
	if !assert.Len(t, transformed, 12) {
		return
	}
	for _, coreIdx := range getAllocatedPE() {
		assert.Equal(t, metricPodKey("ns-dra", "dra-pod", "dra-container", strconv.Itoa(coreIdx)), metricPodKeys(transformed[coreIdx : coreIdx+1])[0])
	}
	assert.Equal(t, []string{
		metricPodKey("ns-shared-a", "shared-pod-a", "shared-container-a", "3"),
		metricPodKey("ns-shared-b", "shared-pod-b", "shared-container-b", "3"),
		metricPodKey("", "", "", "4"),
		metricPodKey("", "", "", "invalid"),
	}, metricPodKeys(transformed[8:]))
	assert.Equal(t, "uuid-none", transformed[10][uuid])
	assert.Equal(t, "uuid-plugin", transformed[11][uuid])
}

func TestListPods_V1RoundTripAndError(t *testing.T) {
	t.Run("roundtrip dynamic resources", func(t *testing.T) {
		response := &podResourcesAPI.ListPodResourcesResponse{
			PodResources: []*podResourcesAPI.PodResources{{
				Name:      "dra-pod",
				Namespace: "ns-dra",
				Containers: []*podResourcesAPI.ContainerResources{{
					Name: "dra-container",
					DynamicResources: []*podResourcesAPI.DynamicResource{{
						ClaimName:      "claim-a",
						ClaimNamespace: "claims-ns",
						ClaimResources: []*podResourcesAPI.ClaimResource{{
							DriverName: furiosaDRADriverName,
							PoolName:   "pool0",
							DeviceName: "npu-dra",
							CdiDevices: []*podResourcesAPI.CDIDevice{{Name: "furiosa.example/device=npu-dra"}},
						}},
					}},
				}},
			}},
		}
		conn := newTestPodResourcesClient(t, &fakePodResourcesServer{response: response})

		resp, err := listPods(conn)

		assert.NoError(t, err)
		assert.True(t, proto.Equal(response, resp), "pod resources must survive the v1 gRPC roundtrip")
	})

	t.Run("list error", func(t *testing.T) {
		conn := newTestPodResourcesClient(t, &fakePodResourcesServer{listErr: status.Error(codes.Internal, "boom")})

		resp, err := listPods(conn)

		assert.Nil(t, resp)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to get pod resources")
			assert.Contains(t, err.Error(), "boom")
		}
	})
}

func TestKubeResourcesMapper_TransformDeviceMetrics_EdgeCases(t *testing.T) {
	deviceWiseCache, coreWiseCache := buildMultiWiseCacheFromPodResources(testPodResourcesResponse())
	for _, coreWise := range []bool{false, true} {
		t.Run("coreWise="+strconv.FormatBool(coreWise), func(t *testing.T) {
			mapper := &kubeResourcesMapper{enabled: true, deviceWiseCache: deviceWiseCache, coreWiseCache: coreWiseCache}
			t.Run("missing identity and unknown devices", func(t *testing.T) {
				metrics := MetricContainer{nil, {}, {uuid: 123, device: 456, core: "0"},
					newMetricWithIdentity("", "", "0"),
					newMetricWithIdentity("unknown", "npu-driver-ignored", "0")}
				assert.Equal(t, metrics, mapper.TransformDeviceMetrics(metrics, coreWise))
			})
			t.Run("UUID and device name key separation", func(t *testing.T) {
				metrics := MetricContainer{
					newMetricWithIdentity("npu-dra", "unknown", "0"),
					newMetricWithIdentity("unknown", "uuid-plugin", "0"),
				}
				assert.Equal(t, metrics, mapper.TransformDeviceMetrics(metrics, coreWise))
			})
			t.Run("plugin and DRA users are combined", func(t *testing.T) {
				metric := newMetricWithIdentity("uuid-plugin", "npu-dra", "0")
				result := mapper.TransformDeviceMetrics(MetricContainer{metric}, coreWise)
				if !assert.Len(t, result, 2) {
					return
				}
				assert.Equal(t, "plugin-pod", result[0][kubernetesPod])
				assert.Equal(t, "dra-pod", result[1][kubernetesPod])
			})
			t.Run("disabled and empty caches", func(t *testing.T) {
				metrics := MetricContainer{newMetricWithIdentity("uuid-plugin", "npu-dra", "0")}
				assert.Equal(t, metrics, (&kubeResourcesMapper{deviceWiseCache: deviceWiseCache, coreWiseCache: coreWiseCache}).TransformDeviceMetrics(metrics, coreWise))
				assert.Equal(t, metrics, (&kubeResourcesMapper{enabled: true}).TransformDeviceMetrics(metrics, coreWise))
				assert.Empty(t, mapper.TransformDeviceMetrics(nil, coreWise))
			})
			t.Run("plugin and mixed mappings preserve all values", func(t *testing.T) {
				for _, identity := range [][2]string{{"uuid-plugin", "npu-plugin"}, {"uuid-mixed", "npu-mixed"}} {
					metric := newMetricWithIdentity(identity[0], identity[1], "0")
					metric["value"] = float64(42)
					snapshot := deepCopyMetric(metric)
					result := mapper.TransformDeviceMetrics(MetricContainer{metric}, coreWise)
					if !assert.Len(t, result, 1) {
						return
					}
					assert.NotEmpty(t, result[0][kubernetesPod])
					assert.Equal(t, identity[0], result[0][uuid])
					assert.Equal(t, identity[1], result[0][device])
					assert.Equal(t, float64(42), result[0]["value"])
					result[0]["value"] = float64(99)
					assert.Equal(t, snapshot, metric)
				}
			})
			t.Run("multiple containers and devices", func(t *testing.T) {
				result := mapper.TransformDeviceMetrics(MetricContainer{
					newMetricWithIdentity("unknown", "npu-multi", "0"),
					newMetricWithIdentity("unknown", "npu-shared-2", "0"),
				}, coreWise)
				if !assert.Len(t, result, 3) {
					return
				}
				assert.Equal(t, "container-1", result[0][kubernetesContainer])
				assert.Equal(t, "container-2", result[1][kubernetesContainer])
				assert.Equal(t, "shared-pod-a", result[2][kubernetesPod])
			})
		})
	}

	t.Run("invalid cores", func(t *testing.T) {
		mapper := &kubeResourcesMapper{enabled: true, deviceWiseCache: deviceWiseCache, coreWiseCache: coreWiseCache}
		for _, coreValue := range []any{nil, 0, "", "invalid", "0-7", "-1", "8"} {
			metric := newMetricWithIdentity("uuid-plugin", "npu-dra", "0")
			metric[core] = coreValue
			assert.Equal(t, MetricContainer{metric}, mapper.TransformDeviceMetrics(MetricContainer{metric}, true))
		}
	})
}

func testPodResourcesResponse() *podResourcesAPI.ListPodResourcesResponse {
	return &podResourcesAPI.ListPodResourcesResponse{
		PodResources: []*podResourcesAPI.PodResources{
			nil,
			{Containers: []*podResourcesAPI.ContainerResources{nil, {Devices: []*podResourcesAPI.ContainerDevices{nil}}}},
			{
				Name:      "plugin-pod",
				Namespace: "ns-plugin",
				Containers: []*podResourcesAPI.ContainerResources{{
					Name: "plugin-container",
					Devices: []*podResourcesAPI.ContainerDevices{
						{ResourceName: "furiosa.ai/npu", DeviceIds: []string{"uuid-plugin", "uuid-plugin", ""}},
						{ResourceName: "other.ai/npu", DeviceIds: []string{"uuid-ignored"}},
					},
				}},
			},
			{
				Name:      "dra-pod",
				Namespace: "ns-dra",
				Containers: []*podResourcesAPI.ContainerResources{{
					Name: "dra-container",
					DynamicResources: []*podResourcesAPI.DynamicResource{
						nil,
						{ClaimName: "claim-a", ClaimNamespace: "claims-ns", ClaimResources: []*podResourcesAPI.ClaimResource{
							nil,
							{DriverName: furiosaDRADriverName, DeviceName: "npu-dra"},
							{DriverName: furiosaDRADriverName, DeviceName: "npu-dra"},
							{DriverName: furiosaDRADriverName, DeviceName: ""},
							{DriverName: furiosaDRADriverName + ".extra", DeviceName: "npu-driver-ignored"},
						}},
						{ClaimName: "claim-b", ClaimNamespace: "claims-ns", ClaimResources: []*podResourcesAPI.ClaimResource{
							{DriverName: furiosaDRADriverName, DeviceName: "npu-dra"},
						}},
					},
				}},
			},
			{
				Name:      "mixed-pod",
				Namespace: "ns-mixed",
				Containers: []*podResourcesAPI.ContainerResources{{
					Name: "mixed-container",
					Devices: []*podResourcesAPI.ContainerDevices{{
						ResourceName: "furiosa.ai/npu",
						DeviceIds:    []string{"uuid-mixed"},
					}},
					DynamicResources: []*podResourcesAPI.DynamicResource{{
						ClaimResources: []*podResourcesAPI.ClaimResource{{
							DriverName: furiosaDRADriverName,
							DeviceName: "npu-mixed",
						}},
					}},
				}},
			},
			{
				Name:      "shared-pod-a",
				Namespace: "ns-shared-a",
				Containers: []*podResourcesAPI.ContainerResources{{
					Name: "shared-container-a",
					DynamicResources: []*podResourcesAPI.DynamicResource{{
						ClaimName:      "shared-claim",
						ClaimNamespace: "claims-ns",
						ClaimResources: []*podResourcesAPI.ClaimResource{
							{DriverName: furiosaDRADriverName, DeviceName: "npu-shared"},
							{DriverName: furiosaDRADriverName, DeviceName: "npu-shared-2"},
						},
					}},
				}},
			},
			{
				Name:      "shared-pod-b",
				Namespace: "ns-shared-b",
				Containers: []*podResourcesAPI.ContainerResources{{
					Name: "shared-container-b",
					DynamicResources: []*podResourcesAPI.DynamicResource{{
						ClaimName:      "shared-claim",
						ClaimNamespace: "claims-ns",
						ClaimResources: []*podResourcesAPI.ClaimResource{{
							DriverName: furiosaDRADriverName,
							DeviceName: "npu-shared",
						}},
					}},
				}},
			},
			{
				Name:      "multi-pod",
				Namespace: "ns-multi",
				Containers: []*podResourcesAPI.ContainerResources{
					{
						Name: "container-1",
						DynamicResources: []*podResourcesAPI.DynamicResource{{
							ClaimResources: []*podResourcesAPI.ClaimResource{{
								DriverName: furiosaDRADriverName,
								DeviceName: "npu-multi",
							}},
						}},
					},
					{
						Name: "container-2",
						DynamicResources: []*podResourcesAPI.DynamicResource{{
							ClaimResources: []*podResourcesAPI.ClaimResource{{
								DriverName: furiosaDRADriverName,
								DeviceName: "npu-multi",
							}},
						}},
					},
				},
			},
		},
	}
}

func newMetricWithIdentity(uuidValue, deviceValue, coreValue string) Metric {
	metric := newMetric()
	metric[uuid] = uuidValue
	metric[device] = deviceValue
	metric[core] = coreValue
	return metric
}

func cloneMetricContainer(metrics MetricContainer) MetricContainer {
	cloned := make(MetricContainer, 0, len(metrics))
	for _, metric := range metrics {
		cloned = append(cloned, deepCopyMetric(metric))
	}
	return cloned
}

func metricPodKeys(metrics MetricContainer) []string {
	keys := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		keys = append(keys, metricPodKey(
			metricString(metric, kubernetesNamespace),
			metricString(metric, kubernetesPod),
			metricString(metric, kubernetesContainer),
			metricString(metric, core),
		))
	}
	return keys
}

func metricPodKey(namespace, pod, container, coreValue string) string {
	return namespace + "/" + pod + "/" + container + "/" + coreValue
}

func metricString(metric Metric, key string) string {
	value, _ := metric[key].(string)
	return value
}

type fakePodResourcesServer struct {
	podResourcesAPI.UnimplementedPodResourcesListerServer
	response *podResourcesAPI.ListPodResourcesResponse
	listErr  error
}

func (s *fakePodResourcesServer) List(context.Context, *podResourcesAPI.ListPodResourcesRequest) (*podResourcesAPI.ListPodResourcesResponse, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.response, nil
}

func newTestPodResourcesClient(t *testing.T, server podResourcesAPI.PodResourcesListerServer) *grpc.ClientConn {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	podResourcesAPI.RegisterPodResourcesListerServer(grpcServer, server)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	t.Cleanup(func() {
		grpcServer.Stop()
		_ = listener.Close()
	})

	conn, err := grpc.NewClient("passthrough:///"+listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}
