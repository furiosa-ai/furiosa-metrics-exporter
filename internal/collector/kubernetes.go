package collector

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"

	podResourcesAPI "k8s.io/kubelet/pkg/apis/podresources/v1"
)

const (
	k8sSocket             = "/var/lib/kubelet/pod-resources/kubelet.sock"
	furiosaResourcePrefix = "furiosa.ai"
	furiosaDRADriverName  = "npu.furiosa.ai"
)

type podInfo struct {
	Name          string
	Namespace     string
	ContainerName string
	AllocatedPE   []int
	CoreLabel     string
}

type podContainerID struct {
	namespace string
	pod       string
	container string
}

// deviceWiseCache maps a device identifier to pods for device wise metrics.
type deviceWiseCache map[string][]podInfo

type coreToPodInfo map[int][]podInfo

// coreWiseCache maps a device identifier and core to pods for core wise metrics.
type coreWiseCache map[string]coreToPodInfo

type KubeResourcesMapper interface {
	TransformDeviceMetrics(metrics MetricContainer, coreWiseMetric bool) MetricContainer
}

type kubeResourcesMapper struct {
	enabled bool
	sync.RWMutex
	deviceWiseCache
	coreWiseCache
}

var _ KubeResourcesMapper = (*kubeResourcesMapper)(nil)

func NewKubeResourcesMapper(ctx context.Context, enabled bool) (KubeResourcesMapper, chan<- struct{}, error) {
	syncChan := make(chan struct{}, 1)

	mapper := &kubeResourcesMapper{
		enabled:         enabled,
		deviceWiseCache: make(deviceWiseCache),
	}

	go func() {
		for {
			select {
			case <-syncChan:
				mapper.syncPodInfoCache()
			case <-ctx.Done():
				return
			}
		}
	}()

	return mapper, syncChan, nil
}

func (k *kubeResourcesMapper) syncPodInfoCache() {
	if !k.enabled {
		return
	}

	deviceWise, coreWise, err := buildMultiWiseCache()
	if err != nil {
		fmt.Printf("failed to get kubernetes pod information cache: %v", err)
		return
	}

	k.Lock()
	defer k.Unlock()

	k.deviceWiseCache = deviceWise
	k.coreWiseCache = coreWise
}

func (k *kubeResourcesMapper) TransformDeviceMetrics(metrics MetricContainer, coreWiseMetric bool) MetricContainer {
	if !k.enabled {
		return metrics
	}

	transformed := make(MetricContainer, 0)

	k.RLock()
	defer k.RUnlock()

	for _, metric := range metrics {
		deviceUUID, _ := metric[uuid].(string)
		deviceName, _ := metric[device].(string)
		draKey := draDeviceKey(deviceName)
		var podInfos []podInfo

		if coreWiseMetric {
			// handle core wise metrics like utilization and performance counter
			coreValue, coreFound := metric[core].(string)
			if !coreFound {
				transformed = append(transformed, metric)
				continue
			}

			coreIdx, err := strconv.Atoi(coreValue)
			if err != nil {
				transformed = append(transformed, metric)
				continue
			}
			podInfos = append(podInfos, k.coreWiseCache[deviceUUID][coreIdx]...)
			podInfos = append(podInfos, k.coreWiseCache[draKey][coreIdx]...)
		} else {
			podInfos = append(podInfos, k.deviceWiseCache[deviceUUID]...)
			podInfos = append(podInfos, k.deviceWiseCache[draKey]...)
		}
		if len(podInfos) == 0 {
			transformed = append(transformed, metric)
			continue
		}

		seenContainers := make(map[podContainerID]bool)
		for _, podInformation := range podInfos {
			containerID := podContainerID{
				namespace: podInformation.Namespace,
				pod:       podInformation.Name,
				container: podInformation.ContainerName,
			}
			if seenContainers[containerID] {
				continue
			}
			seenContainers[containerID] = true

			copied := deepCopyMetric(metric)
			copied[kubernetesNamespace] = podInformation.Namespace
			copied[kubernetesPod] = podInformation.Name
			copied[kubernetesContainer] = podInformation.ContainerName
			// Device-wise metrics use a core range; core-wise metrics retain their core index.
			if !coreWiseMetric {
				copied[core] = podInformation.CoreLabel
			}
			transformed = append(transformed, copied)
		}
	}

	return transformed
}

// draDeviceKey separates DRA's SMI device names from device-plugin UUIDs.
func draDeviceKey(deviceName string) string {
	return furiosaDRADriverName + "/" + deviceName
}

func buildMultiWiseCache() (deviceWiseCache, coreWiseCache, error) {
	_, err := os.Stat(k8sSocket)
	if os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("kubelet socket '%s' does not exist", k8sSocket)
	}

	c, cleanup, err := connectToServer()
	if err != nil {
		return nil, nil, err
	}
	defer cleanup()

	devicePods, err := listPods(c)
	if err != nil {
		return nil, nil, err
	}

	deviceWise, coreWise := buildMultiWiseCacheFromPodResources(devicePods)
	return deviceWise, coreWise, nil
}

func buildMultiWiseCacheFromPodResources(devicePods *podResourcesAPI.ListPodResourcesResponse) (deviceWiseCache, coreWiseCache) {
	deviceWise := make(deviceWiseCache)
	coreWise := make(coreWiseCache)

	for _, podResource := range devicePods.GetPodResources() {
		for _, containerResource := range podResource.GetContainers() {
			deviceKeys := make(map[string]bool)
			for _, containerDevice := range containerResource.GetDevices() {
				resource := containerDevice.GetResourceName()
				if !strings.HasPrefix(resource, furiosaResourcePrefix) {
					continue
				}

				for _, deviceID := range containerDevice.GetDeviceIds() {
					if deviceID == "" {
						continue
					}

					deviceKeys[deviceID] = true
				}
			}

			for _, dynamicResource := range containerResource.GetDynamicResources() {
				for _, claimResource := range dynamicResource.GetClaimResources() {
					if claimResource.GetDriverName() != furiosaDRADriverName {
						continue
					}

					deviceName := claimResource.GetDeviceName()
					if deviceName == "" {
						continue
					}

					deviceKeys[draDeviceKey(deviceName)] = true
				}
			}

			for deviceKey := range deviceKeys {
				// fixme: use resource spec information from SMI or furiosaDevice
				allocatedPE := getAllocatedPE()
				podInformation := podInfo{
					Name:          podResource.GetName(),
					Namespace:     podResource.GetNamespace(),
					ContainerName: containerResource.GetName(),
					AllocatedPE:   allocatedPE,
					CoreLabel:     "0-7",
				}

				// build device wise cache
				deviceWise[deviceKey] = append(deviceWise[deviceKey], podInformation)

				// build core wise cache
				if _, ok := coreWise[deviceKey]; !ok {
					coreWise[deviceKey] = make(coreToPodInfo)
				}
				for _, coreIdx := range allocatedPE {
					coreWise[deviceKey][coreIdx] = append(coreWise[deviceKey][coreIdx], podInformation)
				}
			}
		}
	}

	return deviceWise, coreWise
}

func connectToServer() (*grpc.ClientConn, func(), error) {
	resolver.SetDefaultScheme("passthrough")

	conn, err := grpc.NewClient(k8sSocket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "unix", addr)
		}))

	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to connect to '%s'; err: %w", k8sSocket, err)
	}

	return conn, func() {
		_ = conn.Close()
	}, nil
}

func listPods(conn *grpc.ClientConn) (*podResourcesAPI.ListPodResourcesResponse, error) {
	client := podResourcesAPI.NewPodResourcesListerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.List(ctx, &podResourcesAPI.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod resources; err: %w", err)
	}

	return resp, nil
}

func getAllocatedPE() []int {
	return []int{0, 1, 2, 3, 4, 5, 6, 7}
}
