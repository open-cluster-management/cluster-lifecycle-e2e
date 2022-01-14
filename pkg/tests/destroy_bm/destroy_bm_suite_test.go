package destroy_bm

import (
	"flag"
	"fmt"
	"testing"

	libgocmd "github.com/stolostron/library-e2e-go/pkg/cmd"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"k8s.io/klog"
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
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("%s-%d.xml", "/results/result-destroy-bm", config.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, "Destroy baremetal Suite", []Reporter{junitReporter})
}
