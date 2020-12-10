// Copyright (c) 2020 Red Hat, Inc.

// +build e2e

package e2e

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	libgocrdv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/crd"
	libgodeploymentv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/deployment"

	"k8s.io/klog"
)

var _ = Describe("Cluster-lifecycle: [P1][Sev1][cluster-lifecycle] Check local-cluster imported", func() {

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(15 * time.Minute)
		SetDefaultEventuallyPollingInterval(10 * time.Second)
	})

	It("Check if local-cluster is imported on hub", func() {
		clusterName := "local-cluster"
		klog.V(1).Infof("========================= Test cluster import hub %s ===============================", clusterName)
		Eventually(func() bool {
			klog.V(1).Infof("Cluster %s: Check CRDs", clusterName)
			has, _, _ := libgocrdv1.HasCRDs(hubClientAPIExtension,
				[]string{
					"managedclusters.cluster.open-cluster-management.io",
					"manifestworks.work.open-cluster-management.io",
				})
			return has
		}).Should(BeTrue())

		Eventually(func() error {
			_, _, err := libgodeploymentv1.HasDeploymentsInNamespace(hubClient,
				"open-cluster-management",
				[]string{
					"managedcluster-import-controller",
					"klusterlet-addon-controller",
				})
			return err
		}).Should(BeNil())

		Eventually(func() error {
			_, _, err := libgodeploymentv1.HasDeploymentsInNamespace(hubClient,
				"open-cluster-management-hub",
				[]string{"cluster-manager-registration-controller"})
			return err
		}).Should(BeNil())

		By("Checking namespace local-cluster is present in which the cluster is imported", func() {
			namespaces := hubClient.CoreV1().Namespaces()
			_, err := namespaces.Get(context.TODO(), clusterName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			klog.V(1).Infof("Cluster %s: Namespace %s is present", clusterName, clusterName)
		})

		By("Checking the managedCluster resource is present on hub", func() {
			gvr := schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1", Resource: "managedclusters"}
			_, err := hubClientDynamic.Resource(gvr).Get(context.TODO(), clusterName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			klog.V(1).Infof("Cluster %s: ManagedCluster resourec %s is present", clusterName, clusterName)
		})

		When(fmt.Sprintf("Checking cluster %s to be ready", clusterName), func() {
			waitClusterImported(hubClientDynamic, clusterName)
		})

		When(fmt.Sprintf("Cluster %s ready, wait manifestWorks to be applied", clusterName), func() {
			checkManifestWorksApplied(hubClientDynamic, clusterName)
		})

		When(fmt.Sprintf("Import launched, wait for Add-Ons %s to be available", clusterName), func() {
			waitClusterAdddonsAvailable(hubClientDynamic, clusterName)
		})

	})

})
