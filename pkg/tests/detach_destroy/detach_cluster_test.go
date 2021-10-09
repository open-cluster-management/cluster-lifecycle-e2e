// Copyright (c) 2020 Red Hat, Inc.

package detach_destroy

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/open-cluster-management/cluster-lifecycle-e2e/pkg/clients"
	"github.com/open-cluster-management/cluster-lifecycle-e2e/pkg/utils"
	libgooptions "github.com/open-cluster-management/library-e2e-go/pkg/options"
	libgocrdv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/crd"
	libgodeploymentv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/deployment"
	libgoclient "github.com/open-cluster-management/library-go/pkg/client"
	libgoconfig "github.com/open-cluster-management/library-go/pkg/config"

	"k8s.io/klog"
)

var _ = Describe("Cluster-lifecycle: [P1][Sev1][cluster-lifecycle] Detach cluster", func() {

	var err error
	var managedClusterKubeClient kubernetes.Interface
	var managedClusterDynamicClient dynamic.Interface
	var managedClusterDiscoveryClient *discovery.DiscoveryClient
	var hubClients *clients.HubClients

	BeforeEach(func() {
		hubClients = clients.GetHubClients()
	})

	It("Given a list of clusters to detach (cluster/g0/detach-service-resources)", func() {
		for _, managedCluster := range libgooptions.TestOptions.Options.ManagedClusters {
			var clusterName = managedCluster.Name
			klog.V(1).Infof("========================= Test cluster detach cluster %s ===============================", managedCluster.Name)
			managedClusterKubeClient, err = libgoclient.NewDefaultKubeClient(managedCluster.KubeConfig)
			Expect(err).To(BeNil())
			managedClusterDynamicClient, err = libgoclient.NewDefaultKubeClientDynamic(managedCluster.KubeConfig)
			Expect(err).To(BeNil())
			managedClusterRestConfig, err := libgoconfig.LoadConfig("", managedCluster.KubeConfig, "")
			Expect(err).To(BeNil())
			managedClusterDiscoveryClient, err = discovery.NewDiscoveryClientForConfig(managedClusterRestConfig)
			Expect(err).To(BeNil())
			Eventually(func() bool {
				klog.V(1).Infof("Cluster %s: Check CRDs", clusterName)
				has, _, _ := libgocrdv1.HasCRDs(hubClients.APIExtensionClient,
					[]string{
						"managedclusters.cluster.open-cluster-management.io",
						"manifestworks.work.open-cluster-management.io",
						"klusterletaddonconfigs.agent.open-cluster-management.io",
					})
				return has
			}).Should(BeTrue())

			Eventually(func() error {
				_, _, err := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
					"open-cluster-management",
					[]string{
						"managedcluster-import-controller-v2",
						"klusterlet-addon-controller-v2",
					})
				return err
			}).Should(BeNil())

			Eventually(func() error {
				_, _, err := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
					"open-cluster-management-hub",
					[]string{"cluster-manager-registration-controller"})
				return err
			}).Should(BeNil())

			By(fmt.Sprintf("Detaching the %s CR on the hub", clusterName), func() {
				klog.V(1).Infof("Cluster %s: Detaching the %s CR on the hub", clusterName, clusterName)
				gvr := schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1", Resource: "managedclusters"}
				Expect(hubClients.DynamicClient.Resource(gvr).Delete(context.TODO(), clusterName, metav1.DeleteOptions{})).Should(BeNil())
			})

			klog.V(1).Infof("Cluster %s: Wait 20 min for cluster to go in Unknown state", clusterName)
			time.Sleep(20 * time.Minute)

			When(fmt.Sprintf("the detach of the cluster %s is requested, wait for the effective detach", clusterName), func() {
				waitDetached(hubClients.DynamicClient, clusterName)
			})

			When("the namespace is deleted, check if managed cluster is well cleaned", func() {
				By(fmt.Sprintf("Checking if the %s namespace is deleted", openClusterManagementAgentAddonNamespace), func() {
					klog.V(1).Infof("Cluster %s: Checking if the %s is deleted", clusterName, openClusterManagementAgentAddonNamespace)
					Eventually(func() bool {
						klog.V(1).Infof("Cluster %s: Wait %s namespace deletion...", clusterName, openClusterManagementAgentAddonNamespace)
						_, err := managedClusterKubeClient.CoreV1().Namespaces().Get(context.TODO(), openClusterManagementAgentAddonNamespace, metav1.GetOptions{})
						if err != nil {
							klog.V(1).Infof("Cluster %s: %s", clusterName, err)
							return errors.IsNotFound(err)
						}
						err = utils.PrintLeftOver(managedClusterDynamicClient, managedClusterDiscoveryClient, openClusterManagementAgentAddonNamespace)
						if err != nil {
							klog.Error(err)
						}
						return false
					}).Should(BeTrue())
				})
				By(fmt.Sprintf("Checking if the %s namespace is deleted", openClusterManagementAgentNamespace), func() {
					klog.V(1).Infof("Cluster %s: Checking if the %s is deleted", clusterName, openClusterManagementAgentNamespace)
					Eventually(func() bool {
						klog.V(1).Infof("Cluster %s: Wait %s namespace deletion...", clusterName, openClusterManagementAgentNamespace)
						_, err := managedClusterKubeClient.CoreV1().Namespaces().Get(context.TODO(), openClusterManagementAgentNamespace, metav1.GetOptions{})
						if err != nil {
							klog.V(1).Infof("Cluster %s: %s", clusterName, err)
							return errors.IsNotFound(err)
						}
						err = utils.PrintLeftOver(managedClusterDynamicClient, managedClusterDiscoveryClient, openClusterManagementAgentNamespace)
						if err != nil {
							klog.Error(err)
						}
						return false
					}).Should(BeTrue())
				})
				By(fmt.Sprintf("Checking if the %s crd is deleted", klusterletCRDName), func() {
					klog.V(1).Infof("Cluster %s: Checking if the %s crd is deleted", clusterName, klusterletCRDName)
					gvr := schema.GroupVersionResource{Group: "operator.open-cluster-management.io", Version: "v1", Resource: "klusterlets"}
					Eventually(func() bool {
						klog.V(1).Infof("Cluster %s: Wait %s crd deletion...", clusterName, klusterletCRDName)
						_, err := managedClusterDynamicClient.Resource(gvr).Get(context.TODO(), klusterletCRDName, metav1.GetOptions{})
						if err != nil {
							klog.V(1).Infof("Cluster %s: %s", clusterName, err)
							return errors.IsNotFound(err)
						}
						return false
					}).Should(BeTrue())
				})
			})

			When("the deletion of the cluster is done, wait for the namespace deletion", func() {
				By(fmt.Sprintf("Checking the deletion of the %s namespace on the hub", clusterName), func() {
					klog.V(1).Infof("Cluster %s: Checking the deletion of the %s namespace on the hub", clusterName, clusterName)
					Eventually(func() bool {
						klog.V(1).Infof("Cluster %s: Wait %s namespace deletion...", clusterName, clusterName)
						_, err := hubClients.KubeClient.CoreV1().Namespaces().Get(context.TODO(), clusterName, metav1.GetOptions{})
						if err != nil {
							klog.V(1).Infof("Cluster %s: %s", clusterName, err)
							return errors.IsNotFound(err)
						}
						err = utils.PrintLeftOver(hubClients.DynamicClient, hubClients.DiscoveryClient, clusterName)
						if err != nil {
							klog.Error(err)
						}
						return false
					}).Should(BeTrue())
					klog.V(1).Infof("Cluster %s: %s namespace deleted", clusterName, clusterName)
				})
			})

		}

	})

})

func waitDetached(hubClientDynamic dynamic.Interface, clusterName string) {
	By(fmt.Sprintf("Checking the deletion of the %s managedCluster on the hub", clusterName), func() {
		klog.V(1).Infof("Cluster %s: Checking the deletion of the %s managedCluster on the hub", clusterName, clusterName)
		gvr := schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1", Resource: "managedclusters"}
		Eventually(func() bool {
			klog.V(1).Infof("Cluster %s: Wait %s managedCluster deletion...", clusterName, clusterName)
			_, err := hubClientDynamic.Resource(gvr).Get(context.TODO(), clusterName, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				return true
			}

			if err != nil {
				klog.Errorf("Cluster %s: Failed to get managed cluster: %v", clusterName, err)
				return false
			}

			return false
		}, 60*time.Second, 1*time.Second).Should(BeTrue())
		klog.V(1).Infof("Cluster %s: %s managedCluster deleted", clusterName, clusterName)
	})
}
