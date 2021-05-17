// Copyright (c) 2020 Red Hat, Inc.

package metrics

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/open-cluster-management/cluster-lifecycle-e2e/pkg/clients"
	libgocmd "github.com/open-cluster-management/library-e2e-go/pkg/cmd"
	"k8s.io/klog"
)

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)

	libgocmd.InitFlags(nil)
}

var _ = BeforeSuite(func() {
	hubClients = clients.GetHubClients()
})

func TestMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics Suite")
}
