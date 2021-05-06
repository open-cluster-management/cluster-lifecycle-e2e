package detach_destroy

import (
	"flag"
	"testing"

	libgocmd "github.com/open-cluster-management/library-e2e-go/pkg/cmd"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/klog"
)

const (
	klusterletCRDName                        = "klusterlet"
	openClusterManagementAgentNamespace      = "open-cluster-management-agent"
	openClusterManagementAgentAddonNamespace = "open-cluster-management-agent-addon"
)

var cloudProviders string

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)

	libgocmd.InitFlags(nil)

	flag.StringVar(&cloudProviders, "cloud-providers", "",
		"A comma separated list of cloud providers (ie: aws,azure) "+
			"If set only these cloud providers will be tested")

}

var _ = BeforeSuite(func() {
})

func TestDetachDestroy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DetachDestroy Suite")
}
