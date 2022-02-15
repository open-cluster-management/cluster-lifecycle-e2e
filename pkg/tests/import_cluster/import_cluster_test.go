// Copyright (c) 2020 Red Hat, Inc.

package import_cluster

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stolostron/applier/pkg/applier"
	"github.com/stolostron/applier/pkg/templateprocessor"
	"github.com/stolostron/cluster-lifecycle-e2e/pkg/appliers"
	"github.com/stolostron/cluster-lifecycle-e2e/pkg/clients"
	"github.com/stolostron/cluster-lifecycle-e2e/pkg/utils"
	libgooptions "github.com/stolostron/library-e2e-go/pkg/options"
	libgocrdv1 "github.com/stolostron/library-go/pkg/apis/meta/v1/crd"
	libgodeploymentv1 "github.com/stolostron/library-go/pkg/apis/meta/v1/deployment"
	libgoclient "github.com/stolostron/library-go/pkg/client"

	"k8s.io/klog"
)

const (
	_v1APIExtensionKubeMinVersion = "v1.16.0"
)

var v1APIExtensionMinVersion = version.MustParseGeneric(_v1APIExtensionKubeMinVersion)

var _ = Describe("Cluster-lifecycle: [P1][Sev1][cluster-lifecycle] Import cluster", func() {
	var hubClients *clients.HubClients

	var err error
	var managedClusterClient client.Client

	BeforeEach(func() {
		hubClients = clients.GetHubClients()
	})

	It("Given a list of clusters to import (cluster/g0/import-service-resources)", func() {
		hubApplier := appliers.GetHubAppliers(hubClients)
		// for clusterName, clusterKubeconfig := range managedClustersForManualImport {
		for _, managedCluster := range libgooptions.TestOptions.Options.ManagedClusters {
			var clusterName = managedCluster.Name
			klog.V(1).Infof("========================= Test cluster import cluster %s ===============================", clusterName)
			managedClusterClient, err = libgoclient.NewDefaultClient(managedCluster.KubeConfig, client.Options{})
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
					"multicluster-engine",
					[]string{
						"managedcluster-import-controller-v2",
					})
				return err
			}).Should(BeNil())

			Eventually(func() error {
				_, _, err := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
					"open-cluster-management",
					[]string{
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

			By("creating the namespace in which the cluster will be imported", func() {
				// Create the cluster NS on master
				klog.V(1).Infof("Cluster %s: Creating the namespace in which the cluster will be imported", clusterName)
				namespaces := hubClients.KubeClient.CoreV1().Namespaces()
				_, err := namespaces.Get(context.TODO(), clusterName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) {
						Expect(namespaces.Create(context.TODO(), &corev1.Namespace{
							ObjectMeta: metav1.ObjectMeta{
								Name: clusterName,
							},
						}, metav1.CreateOptions{})).NotTo(BeNil())
						Expect(namespaces.Get(context.TODO(), clusterName, metav1.GetOptions{})).NotTo(BeNil())
					} else {
						Fail(err.Error())
					}
				}
			})

			By("creating the managedCluster and klusterletaddonconfig", func() {
				klog.V(1).Infof("Cluster %s: Creating the managedCluster and klusterletaddonconfig", clusterName)
				values := struct {
					ManagedClusterName string
				}{
					ManagedClusterName: clusterName,
				}
				Expect(hubApplier.ImportApplier.CreateOrUpdateInPath(".",
					nil,
					false,
					values)).To(BeNil())
			})
			time.Sleep(10 * time.Second)

			var importSecret *corev1.Secret
			When("the managedcluster is created, wait for import secret", func() {
				var err error
				Eventually(func() error {
					klog.V(1).Infof("Cluster %s: Wait import secret %s...", clusterName, clusterName)
					importSecret, err = hubClients.KubeClient.CoreV1().Secrets(clusterName).Get(context.TODO(), clusterName+"-import", metav1.GetOptions{})
					if err != nil {
						klog.V(1).Infof("Cluster %s: %s", clusterName, err)
					}
					return err
				}).Should(BeNil())
				klog.V(1).Infof("Cluster %s: bootstrap import secret %s created", clusterName, clusterName+"-import")
			})

			By("Launching the manual import", func() {
				klog.V(1).Infof("Cluster %s: Apply the crds.yaml", clusterName)
				isV1, err := isAPIExtensionV1(managedCluster.KubeConfig)
				Expect(err).To(BeNil())
				var importStringReader *templateprocessor.YamlStringReader
				if isV1 {
					klog.V(5).Infof("Cluster %s: importSecret.Data[v1]: %s\n", clusterName, importSecret.Data["crdsv1.yaml"])
					importStringReader = templateprocessor.NewYamlStringReader(string(importSecret.Data["crdsv1.yaml"]), templateprocessor.KubernetesYamlsDelimiter)
				} else {
					klog.V(5).Infof("Cluster %s: importSecret.Data[v1beta1]: %s\n", clusterName, importSecret.Data["crdsv1beta1.yaml"])
					importStringReader = templateprocessor.NewYamlStringReader(string(importSecret.Data["crdsv1beta1.yaml"]), templateprocessor.KubernetesYamlsDelimiter)
				}
				managedClusterApplier, err := applier.NewApplier(importStringReader, &templateprocessor.Options{}, managedClusterClient, nil, nil, nil)
				Expect(err).To(BeNil())
				Expect(managedClusterApplier.CreateOrUpdateInPath(".", nil, false, nil)).NotTo(HaveOccurred())
				// Wait 2 sec to make sure the CRDs are effective. The UI does the same.
				time.Sleep(2 * time.Second)
				klog.V(1).Infof("Cluster %s: Apply the import.yaml", clusterName)
				klog.V(5).Infof("Cluster %s: importSecret.Data[import.yaml]: %s\n", clusterName, importSecret.Data["import.yaml"])
				importStringReader = templateprocessor.NewYamlStringReader(string(importSecret.Data["import.yaml"]), templateprocessor.KubernetesYamlsDelimiter)
				managedClusterApplier, err = applier.NewApplier(importStringReader, &templateprocessor.Options{}, managedClusterClient, nil, nil, nil)
				Expect(err).To(BeNil())
				Expect(managedClusterApplier.CreateOrUpdateInPath(".", nil, false, nil)).NotTo(HaveOccurred())
			})

			time.Sleep(1 * time.Minute)

			When(fmt.Sprintf("Import launched, wait for cluster %s to be ready", clusterName), func() {
				utils.WaitClusterImported(hubClients.DynamicClient, clusterName)
			})

			time.Sleep(3 * time.Minute)
			When(fmt.Sprintf("Cluster %s ready, wait manifestWorks to be applied", clusterName), func() {
				checkManifestWorksApplied(hubClients.DynamicClient, clusterName)
			})

			klog.V(1).Infof("Cluster %s: Wait 3 min to settle", clusterName)
			time.Sleep(3 * time.Minute)

			When(fmt.Sprintf("Import launched, wait for Add-Ons %s to be available", clusterName), func() {
				utils.WaitClusterAdddonsAvailable(hubClients.DynamicClient, clusterName)
			})

		}

	})

})

func isAPIExtensionV1(kubeConfig string) (bool, error) {

	config, err := clientcmd.LoadFromFile(kubeConfig)
	if err != nil {
		return false, err
	}

	rconfig, err := clientcmd.NewDefaultClientConfig(
		*config,
		&clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		return false, err
	}

	kubeClient, err := kubernetes.NewForConfig(rconfig)
	if err != nil {
		return false, err
	}

	// Search the kubernestes version by connecting to the managed cluster
	kubeVersion, err := kubeClient.ServerVersion()
	if err != nil {
		return false, err
	}
	isV1, err := v1APIExtensionMinVersion.Compare(kubeVersion.String())
	if err != nil {
		return false, err
	}
	klog.V(4).Infof("isV1: %t", isV1 == -1)
	return isV1 == -1, nil
}
