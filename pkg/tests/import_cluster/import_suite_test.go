package import_cluster

import (
	"context"
	"flag"
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	libgocmd "github.com/open-cluster-management/library-e2e-go/pkg/cmd"
	libgounstructuredv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

const (
	klusterletCRDName       = "klusterlet"
	manifestWorkNamePostfix = "-klusterlet"
	manifestWorkCRDSPostfix = "-crds"
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

func TestImport(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Import Suite")
}

func checkManifestWorksApplied(hubClientDynamic dynamic.Interface, clusterName string) {
	manifestWorkCRDsName := clusterName + manifestWorkNamePostfix + manifestWorkCRDSPostfix
	By(fmt.Sprintf("Checking manfestwork %s to be applied", manifestWorkCRDsName), func() {
		klog.V(1).Infof("Cluster %s: Checking manfestwork %s to be applied", clusterName, manifestWorkCRDsName)
		Eventually(func() error {
			klog.V(1).Infof("Cluster %s: Wait manifestwork %s to be applied...", clusterName, manifestWorkCRDsName)
			gvr := schema.GroupVersionResource{Group: "work.open-cluster-management.io", Version: "v1", Resource: "manifestworks"}
			mwcrd, err := hubClientDynamic.Resource(gvr).Namespace(clusterName).Get(context.TODO(), manifestWorkCRDsName, metav1.GetOptions{})
			if err != nil {
				klog.V(4).Infof("Cluster %s: %s", clusterName, err)
				return err
			}

			var condition map[string]interface{}
			condition, err = libgounstructuredv1.GetConditionByType(mwcrd, "Applied")
			if err != nil {
				klog.V(4).Infof("Cluster %s: %s", clusterName, err)
				return err
			}
			klog.V(4).Info(condition)
			if v, ok := condition["status"]; ok && v == string(metav1.ConditionTrue) {
				return nil
			}
			err = fmt.Errorf("Cluster %s: status not found or not true", clusterName)
			klog.V(4).Infof("Cluster %s: %s", clusterName, err)
			return err
		}).Should(BeNil())
		klog.V(1).Infof("Cluster %s: manifestwork %s applied", clusterName, manifestWorkCRDsName)
	})

	manifestWorkYAMLsName := clusterName + manifestWorkNamePostfix
	By(fmt.Sprintf("Checking manfestwork %s to be applied", manifestWorkYAMLsName), func() {
		klog.V(1).Infof("Cluster %s: Checking manfestwork %s to be applied", clusterName, manifestWorkYAMLsName)
		Eventually(func() error {
			klog.V(1).Infof("Cluster %s: Wait manifestwork %s to be applied...", clusterName, manifestWorkYAMLsName)
			gvr := schema.GroupVersionResource{Group: "work.open-cluster-management.io", Version: "v1", Resource: "manifestworks"}
			mwyaml, err := hubClientDynamic.Resource(gvr).Namespace(clusterName).Get(context.TODO(), manifestWorkYAMLsName, metav1.GetOptions{})
			if err != nil {
				klog.V(4).Info(err)
				return err
			}
			var condition map[string]interface{}
			condition, err = libgounstructuredv1.GetConditionByType(mwyaml, "Applied")
			if err != nil {
				return err
			}
			if v, ok := condition["status"]; ok && v == string(metav1.ConditionTrue) {
				return nil
			}
			return fmt.Errorf("Cluster %s: status not found or not true", clusterName)
		}).Should(BeNil())
		klog.V(1).Infof("Cluster %s: manifestwork %s applied", clusterName, manifestWorkYAMLsName)
	})
}
