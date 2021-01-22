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
	libgocrdv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/crd"
	libgodeploymentv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/deployment"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
					if strings.HasPrefix(name.(string), cloud+"-ginkgo") {
						clusterName = name.(string)
						break
					}
				}
			}
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
					[]string{"managedcluster-import-controller"})
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

		When(fmt.Sprintf("the detach of the cluster %s is requested, wait for the effective detach", clusterName), func() {
			waitDetached(hubClientDynamic, clusterName)
		})

		When(fmt.Sprintf("Detached, delete the clusterDeployment %s", clusterName), func() {
			klog.V(1).Infof("Cluster %s: Deleting the clusterDeployment for cluster %s", clusterName, clusterName)
			gvr := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeployments"}
			Expect(hubClientDynamic.Resource(gvr).Namespace(clusterName).Delete(context.TODO(), clusterName, metav1.DeleteOptions{})).Should(BeNil())
		})

		When(fmt.Sprintf("Wait clusterDeployment %s to be deleted", clusterName), func() {
			waitDetroyed(hubClientDynamic, clusterName)
		})

		When(fmt.Sprintf("Wait namespace %s to be deleted", clusterName), func() {
			waitNamespaceDeleted(hubClient, clusterName)
		})

		klog.V(1).Infof("========================= End Test destroy cluster %s ===============================", clusterName)

	})

}
