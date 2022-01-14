package clients

import (
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/onsi/gomega"

	"github.com/stolostron/cluster-lifecycle-e2e/pkg/tests/options"
	libgooptions "github.com/stolostron/library-e2e-go/pkg/options"
	libgoclient "github.com/stolostron/library-go/pkg/client"
	libgoconfig "github.com/stolostron/library-go/pkg/config"
)

type HubClients struct {
	RestConfig         *rest.Config
	ClientClient       client.Client
	KubeClient         kubernetes.Interface
	DynamicClient      dynamic.Interface
	DiscoveryClient    *discovery.DiscoveryClient
	APIExtensionClient clientset.Interface
}

func GetHubClients() (hubClients *HubClients) {
	hubClients = &HubClients{}
	var err error
	gomega.Expect(options.InitVars()).To(gomega.BeNil())
	hubClients.RestConfig, err = libgoconfig.LoadConfig(libgooptions.TestOptions.Options.Hub.ApiServerURL, libgooptions.TestOptions.Options.Hub.KubeConfig, libgooptions.TestOptions.Options.Hub.KubeContext)
	gomega.Expect(err).To(gomega.BeNil())
	hubClients.KubeClient, err = libgoclient.NewKubeClient(libgooptions.TestOptions.Options.Hub.ApiServerURL, libgooptions.TestOptions.Options.Hub.KubeConfig, libgooptions.TestOptions.Options.Hub.KubeContext)
	gomega.Expect(err).To(gomega.BeNil())
	hubClients.DynamicClient, err = libgoclient.NewKubeClientDynamic(libgooptions.TestOptions.Options.Hub.ApiServerURL, libgooptions.TestOptions.Options.Hub.KubeConfig, libgooptions.TestOptions.Options.Hub.KubeContext)
	gomega.Expect(err).To(gomega.BeNil())
	hubClients.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(hubClients.RestConfig)
	gomega.Expect(err).To(gomega.BeNil())
	hubClients.APIExtensionClient, err = libgoclient.NewKubeClientAPIExtension(libgooptions.TestOptions.Options.Hub.ApiServerURL, libgooptions.TestOptions.Options.Hub.KubeConfig, libgooptions.TestOptions.Options.Hub.KubeContext)
	gomega.Expect(err).To(gomega.BeNil())
	hubClients.ClientClient, err = libgoclient.NewClient(libgooptions.TestOptions.Options.Hub.ApiServerURL, libgooptions.TestOptions.Options.Hub.KubeConfig, libgooptions.TestOptions.Options.Hub.KubeContext, client.Options{})
	gomega.Expect(err).To(gomega.BeNil())
	return
}
