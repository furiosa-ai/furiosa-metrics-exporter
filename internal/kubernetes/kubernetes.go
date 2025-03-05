package kubernetes

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
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

type PodInfo struct {
	Name        string
	Namespace   string
	AllocatedPE []int
	CoreLabel   string
}

func GetDeviceMap() (map[string][]PodInfo, error) {
	return generateDeviceMap()
}

func generateDeviceMap() (map[string][]PodInfo, error) {
	deviceMap := make(map[string][]PodInfo)

	_, err := os.Stat(k8sSocket)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("kubelet socket '%s' does not exist", k8sSocket)
	}

	c, cleanup, err := connectToServer()

	if err != nil {
		return nil, err
	}
	defer cleanup()

	devicePods, err := listPods(c)

	if err != nil {
		return nil, err
	}

	for _, pod := range devicePods.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, device := range container.GetDevices() {

				resource := device.GetResourceName()
				if !strings.HasPrefix(resource, furiosaResourcePrefix) { // Need to check device type? (e.g., warboy, rngd)
					continue
				}

				podName := pod.GetName()
				podNamespace := pod.GetNamespace()

				for _, deviceID := range device.GetDeviceIds() {
					podInfo := PodInfo{
						Name:        podName,
						Namespace:   podNamespace,
						AllocatedPE: getAllocatedPE(deviceID),
						CoreLabel:   getCoreLabel(deviceID),
					}

					uuid := strings.Split(deviceID, furiosaPartitionedResourcePattern)[0]
					deviceMap[uuid] = append(deviceMap[uuid], podInfo)
				}
			}
		}
	}

	return deviceMap, nil
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

func getAllocatedPE(deviceID string) []int {
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

func getCoreLabel(deviceID string) string {
	if !strings.Contains(deviceID, furiosaPartitionedResourcePattern) {
		return "0-7" // TODO(jongchan): warboy case?
	} else {
		return strings.Split(deviceID, furiosaPartitionedResourcePattern)[1]
	}
}
