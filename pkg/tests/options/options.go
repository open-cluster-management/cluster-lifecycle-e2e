package options

import (
	"fmt"
	"os"

	"k8s.io/klog"

	"github.com/onsi/gomega"

	libgocmd "github.com/stolostron/library-e2e-go/pkg/cmd"
	libgooptions "github.com/stolostron/library-e2e-go/pkg/options"
)

var BaseDomain string
var KubeadminUser string
var KubeadminCredential string
var ClusterName string

func InitVars() error {

	err := libgooptions.LoadOptions(libgocmd.End2End.OptionsFile)
	if err != nil {
		klog.Errorf("--options error: %v", err)
		return err
	}

	if libgooptions.TestOptions.Options.Hub.KubeConfig == "" {
		libgooptions.TestOptions.Options.Hub.KubeConfig = os.Getenv("KUBECONFIG")
	}

	if libgooptions.TestOptions.Options.Hub.Name != "" {
		ClusterName = libgooptions.TestOptions.Options.Hub.Name
	}

	if libgooptions.TestOptions.Options.Hub.BaseDomain != "" {
		BaseDomain = libgooptions.TestOptions.Options.Hub.BaseDomain

		if libgooptions.TestOptions.Options.Hub.ApiServerURL == "" {
			libgooptions.TestOptions.Options.Hub.ApiServerURL = fmt.Sprintf("https://api.%s.%s:6443", libgooptions.TestOptions.Options.Hub.Name, libgooptions.TestOptions.Options.Hub.BaseDomain)
		}
	} else {
		gomega.Expect(BaseDomain).NotTo(gomega.BeEmpty(), "The `baseDomain` is required.")
		libgooptions.TestOptions.Options.Hub.BaseDomain = BaseDomain
		libgooptions.TestOptions.Options.Hub.ApiServerURL = fmt.Sprintf("https://api.%s.%s:6443", libgooptions.TestOptions.Options.Hub.Name, BaseDomain)
	}

	if libgooptions.TestOptions.Options.Hub.User != "" {
		KubeadminUser = libgooptions.TestOptions.Options.Hub.User
	}
	if libgooptions.TestOptions.Options.Hub.Password != "" {
		KubeadminCredential = libgooptions.TestOptions.Options.Hub.Password
	}

	if libgooptions.TestOptions.Options.ManagedClusters != nil && len(libgooptions.TestOptions.Options.ManagedClusters) > 0 {
		for i, mc := range libgooptions.TestOptions.Options.ManagedClusters {
			if mc.ApiServerURL == "" {
				libgooptions.TestOptions.Options.ManagedClusters[i].ApiServerURL = fmt.Sprintf("https://api.%s:6443", mc.BaseDomain)
			}
			if mc.KubeConfig == "" {
				libgooptions.TestOptions.Options.ManagedClusters[i].KubeConfig = os.Getenv("IMPORT_KUBECONFIG")
			}
		}
	}

	return nil
}
