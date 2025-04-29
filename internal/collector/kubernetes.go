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

	podResourcesAPI "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
)

const (
	k8sSocket                         = "/var/lib/kubelet/pod-resources/kubelet.sock"
	furiosaResourcePrefix             = "furiosa.ai"
	furiosaPartitionedResourcePattern = "_cores_"
)

type podInfo struct {
	Name          string
	Namespace     string
	ContainerName string
	AllocatedPE   []int
	CoreLabel     string
}

// deviceWiseCache maps uuid to pod information for device wise metrics
type deviceWiseCache map[string][]podInfo

type coreToPodInfo map[int]podInfo

// coreWiseCache maps uuid to core to pod name for core wise metrics
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
		uuidValue, uuidFound := metric[uuid].(string)
		if !uuidFound {
			transformed = append(transformed, metric)
			continue
		}

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

			podInformation, found := k.coreWiseCache[uuidValue][coreIdx]
			if !found {
				transformed = append(transformed, metric)
				continue
			}

			copied := deepCopyMetric(metric)
			copied[kubernetesNamespace] = podInformation.Namespace
			copied[kubernetesPod] = podInformation.Name
			copied[kubernetesContainer] = podInformation.ContainerName
			transformed = append(transformed, copied)

		} else {
			// handle device wise metrics
			podInfoSlice, podInfoSliceFound := k.deviceWiseCache[uuidValue]
			if !podInfoSliceFound {
				transformed = append(transformed, metric)
				continue
			}

			if len(podInfoSlice) == 1 && len(podInfoSlice[0].AllocatedPE) == 8 {
				// exclusive allocation case
				copied := deepCopyMetric(metric)
				copied[kubernetesNamespace] = podInfoSlice[0].Namespace
				copied[kubernetesPod] = podInfoSlice[0].Name
				copied[kubernetesContainer] = podInfoSlice[0].ContainerName
				copied[core] = podInfoSlice[0].CoreLabel
				transformed = append(transformed, copied)
			} else {
				// partitioned allocation case, preserve origin metric and duplicate the metric for each pod
				transformed = append(transformed, metric)
				for _, podInformation := range podInfoSlice {
					duplicated := deepCopyMetric(metric)
					duplicated[kubernetesNamespace] = podInformation.Namespace
					duplicated[kubernetesPod] = podInformation.Name
					duplicated[kubernetesContainer] = podInformation.ContainerName
					duplicated[core] = podInformation.CoreLabel

					transformed = append(transformed, duplicated)
				}
			}
		}
	}

	return transformed
}

func buildMultiWiseCache() (deviceWiseCache, coreWiseCache, error) {
	deviceWise := make(deviceWiseCache)
	coreWise := make(coreWiseCache)

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

	for _, podResource := range devicePods.GetPodResources() {
		for _, containerResource := range podResource.GetContainers() {
			for _, containerDevice := range containerResource.GetDevices() {

				resource := containerDevice.GetResourceName()
				if !strings.HasPrefix(resource, furiosaResourcePrefix) {
					continue
				}

				for _, deviceID := range containerDevice.GetDeviceIds() {
					deviceUUID := strings.Split(deviceID, furiosaPartitionedResourcePattern)[0]
					allocatedPE := getAllocatedPEfromDeviceID(deviceID)

					podInformation := podInfo{
						Name:          podResource.GetName(),
						Namespace:     podResource.GetNamespace(),
						ContainerName: containerResource.GetName(),
						AllocatedPE:   allocatedPE,
						CoreLabel:     getCoreLabelfromDeviceID(deviceID),
					}

					// build device wise cache
					deviceWise[deviceUUID] = append(deviceWise[deviceUUID], podInformation)

					// build core wise cache
					if _, ok := coreWise[deviceUUID]; !ok {
						coreWise[deviceUUID] = make(coreToPodInfo)
					}

					for _, coreIdx := range allocatedPE {
						coreWise[deviceUUID][coreIdx] = podInformation
					}
				}
			}
		}
	}

	return deviceWise, coreWise, nil
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

	return conn, func() { conn.Close() }, nil
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

func getAllocatedPEfromDeviceID(deviceID string) []int {
	if !strings.Contains(deviceID, furiosaPartitionedResourcePattern) {
		return []int{0, 1, 2, 3, 4, 5, 6, 7} // TODO(jongchan): warboy case?
	} else {
		cores := strings.Split(deviceID, furiosaPartitionedResourcePattern)[1]
		coreRange := strings.Split(cores, "-")
		if len(coreRange) == 1 {
			n, _ := strconv.Atoi(coreRange[0])
			return []int{n}
		} else {
			n, _ := strconv.Atoi(coreRange[0])
			m, _ := strconv.Atoi(coreRange[1])
			allocatedPE := []int{}
			for i := n; i <= m; i++ {
				allocatedPE = append(allocatedPE, i)
			}
			return allocatedPE
		}
	}
}

func getCoreLabelfromDeviceID(deviceID string) string {
	if !strings.Contains(deviceID, furiosaPartitionedResourcePattern) {
		return "0-7" // TODO(jongchan): warboy case?
	} else {
		return strings.Split(deviceID, furiosaPartitionedResourcePattern)[1]
	}
}
