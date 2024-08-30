package e2e

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenericBeforeSuiteFunc() Framework {
	frk, err := NewFrameworkWithDefaultNamespace()
	Expect(err).To(BeNil())

	list, err := frk.GetClientSet().CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	Expect(err).To(BeNil())
	Expect(len(list.Items)).Should(BeNumerically(">=", 1))

	return frk
}

func GenericRunTestSuiteFunc(t *testing.T, suitSpec string) {
	RegisterFailHandler(Fail)
	RunSpecs(t, suitSpec)
}
