package collector

type fakeKubeResourcesMapper struct{}

var _ KubeResourcesMapper = (*fakeKubeResourcesMapper)(nil)

func NewFakeKubeResourcesMapper() KubeResourcesMapper {
	return &fakeKubeResourcesMapper{}
}

func (k *fakeKubeResourcesMapper) TransformDeviceMetrics(metrics MetricContainer, _ bool) MetricContainer {
	return metrics
}
