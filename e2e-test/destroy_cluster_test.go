// Copyright (c) 2020 Red Hat, Inc.

// +build e2e

package e2e

import (
	"context"
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	libgooptions "github.com/open-cluster-management/library-e2e-go/pkg/options"
	libgocrdv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/crd"
	libgodeploymentv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/deployment"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var _ = Describe("Cluster-lifecycle: ", func() {
	destroyCluster("aws", "OpenShift")
})

var _ = Describe("Cluster-lifecycle: ", func() {
	destroyCluster("azure", "OpenShift")
})

var _ = Describe("Cluster-lifecycle: ", func() {
	destroyCluster("gcp", "OpenShift")
})

var _ = Describe("Cluster-lifecycle: ", func() {
	destroyCluster("baremetal", "OpenShift")
})

func destroyCluster(cloud, vendor string) {
	// var clusterNameObj *libgooptions.ClusterName
	var clusterName string
	//var err error
	//var imageRefName string

	BeforeEach(func() {
		if cloudProviders != "" && !isRequestedCloudProvider(cloud) {
			Skip(fmt.Sprintf("Cloud provider %s skipped", cloud))
		}

		gvrClusterDeployment := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeployments"}
		clusterDeploymentList, err := hubClientDynamic.Resource(gvrClusterDeployment).List(context.TODO(), metav1.ListOptions{})
		Expect(err).To(BeNil())

		for _, cd := range clusterDeploymentList.Items {
			if metadata, ok := cd.Object["metadata"]; ok {
				meta := metadata.(map[string]interface{})
				if name, ok := meta["name"]; ok {
					if strings.HasPrefix(name.(string), cloud+"-"+libgooptions.GetOwner()) {
						clusterName = name.(string)
						break
					}
				}
			}
		}
		if len(clusterName) == 0 {
			Fail(fmt.Sprintf("No cluster for Cloud provider %s to delete", cloud))
		}

		if cloud == "baremetal" {
			clusterName = libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.ClusterName
		}

		klog.V(1).Infof(`========================= Start Test destroy cluster %s  ===============================`, clusterName)
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultEventuallyPollingInterval(10 * time.Second)
	})

	AfterEach(func() {

	})

	It(fmt.Sprintf("[P1][Sev1][cluster-lifecycle] Destroy cluster %s on %s with vendor %s (cluster/g1/destroy-cluster)", clusterName, cloud, vendor), func() {
		By("Checking the minimal requirements", func() {
			klog.V(1).Infof("Cluster %s: Checking the minimal requirements", clusterName)
			Eventually(func() bool {
				klog.V(2).Infof("Cluster %s: Check CRDs", clusterName)
				has, missing, _ := libgocrdv1.HasCRDs(hubClientAPIExtension,
					[]string{
						"managedclusters.cluster.open-cluster-management.io",
						"clusterdeployments.hive.openshift.io",
						"syncsets.hive.openshift.io",
					})
				if !has {
					klog.Errorf("Cluster %s: Missing CRDs\n%#v", clusterName, missing)
				}
				return has
			}).Should(BeTrue())

			Eventually(func() bool {
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClient,
					"open-cluster-management",
					[]string{"managedcluster-import-controller-v2"})
				if !has {
					klog.Errorf("Cluster %s: Missing deployments\n%#v", clusterName, missing)
				}
				return has
			}).Should(BeTrue())
			Eventually(func() bool {
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClient,
					"open-cluster-management-hub",
					[]string{"cluster-manager-registration-controller"})
				if !has {
					klog.Errorf("Cluster %s: Missing deployments\n%#v", clusterName, missing)
				}
				return has
			}).Should(BeTrue())
			Eventually(func() bool {
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClient,
					"hive",
					[]string{"hive-controllers"})
				if !has {
					klog.Errorf("Missing deployments\n%#v", missing)
				}
				return has
			}).Should(BeTrue())

		})

		By(fmt.Sprintf("Detaching the %s CR on the hub", clusterName), func() {
			klog.V(1).Infof("Cluster %s: Detaching the %s CR on the hub", clusterName, clusterName)
			gvr := schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1", Resource: "managedclusters"}
			Expect(hubClientDynamic.Resource(gvr).Delete(context.TODO(), clusterName, metav1.DeleteOptions{})).Should(BeNil())
		})

		// if cloud != "baremetal" {
		// 	When(fmt.Sprintf("the detach of the cluster %s is requested, wait for the effective detach", clusterName), func() {
		// 		waitDetached(hubClientDynamic, clusterName)
		// 	})
		// }

		When(fmt.Sprintf("Detached, delete the clusterDeployment %s", clusterName), func() {
			klog.V(1).Infof("Cluster %s: Deleting the clusterDeployment for cluster %s", clusterName, clusterName)
			gvr := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeployments"}
			Expect(hubClientDynamic.Resource(gvr).Namespace(clusterName).Delete(context.TODO(), clusterName, metav1.DeleteOptions{})).Should(BeNil())
		})

		When(fmt.Sprintf("Wait clusterDeployment %s to be deleted", clusterName), func() {
			waitDetroyed(hubClientDynamic, clusterName)
		})

		When(fmt.Sprintf("Wait namespace %s to be deleted", clusterName), func() {
			waitNamespaceDeleted(hubClient, hubClientDynamic, hubClientDiscovery, clusterName)
		})

		klog.V(1).Infof("========================= End Test destroy cluster %s ===============================", clusterName)

	})

}

func waitNamespaceDeleted(
	hubClient kubernetes.Interface,
	hubClientDynamic dynamic.Interface,
	hubClientDiscovery *discovery.DiscoveryClient,
	clusterName string) {
	By(fmt.Sprintf("Checking the deletion of the %s namespace on the hub", clusterName), func() {
		klog.V(1).Infof("Cluster %s: Checking the deletion of the %s namespace on the hub", clusterName, clusterName)
		Eventually(func() bool {
			klog.V(1).Infof("Cluster %s: Wait %s namespace deletion...", clusterName, clusterName)
			_, err := hubClient.CoreV1().Namespaces().Get(context.TODO(), clusterName, metav1.GetOptions{})
			if err != nil {
				klog.V(1).Info(err)
				return errors.IsNotFound(err)
			}
			err = printLeftOver(hubClientDynamic, hubClientDiscovery, clusterName)
			if err != nil {
				klog.Error(err)
			}
			return false
		}, 3600, 60).Should(BeTrue())
		klog.V(1).Infof("Cluster %s: %s namespace deleted", clusterName, clusterName)
	})
}
