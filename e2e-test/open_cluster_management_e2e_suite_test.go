// Copyright (c) 2020 Red Hat, Inc.

// +build e2e

package e2e

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	libgocmd "github.com/open-cluster-management/library-e2e-go/pkg/cmd"
	libgooptions "github.com/open-cluster-management/library-e2e-go/pkg/options"
	libgocrdv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/crd"
	libgounstructuredv1 "github.com/open-cluster-management/library-go/pkg/apis/meta/v1/unstructured"
	libgoapplier "github.com/open-cluster-management/library-go/pkg/applier"
	libgoclient "github.com/open-cluster-management/library-go/pkg/client"
	"github.com/open-cluster-management/library-go/pkg/templateprocessor"
)

const (
	importClusterScenario                    = "import"
	selfImportClusterScenario                = "self_import"
	createClusterScenario                    = "create"
	openClusterManagementAgentNamespace      = "open-cluster-management-agent"
	openClusterManagementAgentAddonNamespace = "open-cluster-management-agent-addon"
	klusterletCRDName                        = "klusterlet"
	manifestWorkNamePostfix                  = "-klusterlet"
	manifestWorkCRDSPostfix                  = "-crds"
)

// list of manifestwork name for addon crs
var managedClusteraddOns = []string{
	"application-manager",
	"cert-policy-controller",
	"iam-policy-controller",
	"policy-controller",
	"search-collector",
	"work-manager",
}

var cloudProviders string
var ocpImageRelease string
var reportFile string
var kubeconfig string
var baseDomain string
var kubeadminUser string
var kubeadminCredential string

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)

	libgocmd.InitFlags(nil)

	flag.StringVar(&cloudProviders, "cloud-providers", "",
		"A comma separated list of cloud providers (ie: aws,azure) "+
			"If set only these cloud providers will be tested")
	flag.StringVar(&ocpImageRelease, "ocp-image-release", "",
		"If set this image will be use to create an imageSet reference instead of the one in options.yaml")

	//flag.StringVar(&reportFile, "report-file", "/results/result", "Provide the path to where the junit results will be printed.")

}

func TestOpenClusterManagementE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("%s-%d.xml", "/results/result", config.GinkgoConfig.ParallelNode))
	//junitReporter := reporters.NewJUnitReporter(libgocmd.End2End.ReportFile)
	RunSpecsWithDefaultAndCustomReporters(t, "OpenClusterManagementE2E Suite", []Reporter{junitReporter})
}

var hubClientClient client.Client
var hubClient kubernetes.Interface
var hubClientDynamic dynamic.Interface
var hubClientAPIExtension clientset.Interface
var createTemplateProcessor *templateprocessor.TemplateProcessor
var hubCreateApplier *libgoapplier.Applier
var importYamlReader templateprocessor.TemplateReader
var hubImportApplier *libgoapplier.Applier
var hubSelfImportApplier *libgoapplier.Applier

var _ = BeforeSuite(func() {
	var err error
	Expect(initVars()).To(BeNil())
	//kubeconfig := libgooptions.GetHubKubeConfig(filepath.Join(libgooptions.TestOptions.Hub.ConfigDir, "kube"), libgooptions.TestOptions.Hub.KubeConfigPath)
	hubClient, err = libgoclient.NewKubeClient(libgooptions.TestOptions.Options.Hub.MasterURL, libgooptions.TestOptions.Options.Hub.KubeConfig, libgooptions.TestOptions.Options.Hub.KubeContext)
	Expect(err).To(BeNil())
	hubClientDynamic, err = libgoclient.NewKubeClientDynamic(libgooptions.TestOptions.Options.Hub.MasterURL, libgooptions.TestOptions.Options.Hub.KubeConfig, libgooptions.TestOptions.Options.Hub.KubeContext)
	Expect(err).To(BeNil())
	hubClientAPIExtension, err = libgoclient.NewKubeClientAPIExtension(libgooptions.TestOptions.Options.Hub.MasterURL, libgooptions.TestOptions.Options.Hub.KubeConfig, libgooptions.TestOptions.Options.Hub.KubeContext)
	Expect(err).To(BeNil())
	hubClientClient, err = libgoclient.NewClient(libgooptions.TestOptions.Options.Hub.MasterURL, libgooptions.TestOptions.Options.Hub.KubeConfig, libgooptions.TestOptions.Options.Hub.KubeContext, client.Options{})
	Expect(err).To(BeNil())
	createYamlReader := templateprocessor.NewYamlFileReader(filepath.Join("resources/hub", createClusterScenario))
	createTemplateProcessor, err = templateprocessor.NewTemplateProcessor(createYamlReader, &templateprocessor.Options{})
	Expect(err).To(BeNil())
	hubCreateApplier, err = libgoapplier.NewApplier(createYamlReader, &templateprocessor.Options{}, hubClientClient, nil, nil, nil, nil)
	Expect(err).To(BeNil())
	importYamlReader = templateprocessor.NewYamlFileReader(filepath.Join("resources/hub", importClusterScenario))
	hubImportApplier, err = libgoapplier.NewApplier(importYamlReader, &templateprocessor.Options{}, hubClientClient, nil, nil, nil, nil)
	Expect(err).To(BeNil())
	selfImportYamlReader := templateprocessor.NewYamlFileReader(filepath.Join("resources/hub", selfImportClusterScenario))
	hubSelfImportApplier, err = libgoapplier.NewApplier(selfImportYamlReader, &templateprocessor.Options{}, hubClientClient, nil, nil, nil, nil)
	Expect(err).To(BeNil())
})

var _ = AfterSuite(func() {
})

func initVars() error {

	err := libgooptions.LoadOptions(libgocmd.End2End.OptionsFile)
	if err != nil {
		klog.Errorf("--options error: %v", err)
		return err
	}

	o, _ := yaml.Marshal(libgooptions.TestOptions)
	//Expect(err).NotTo(HaveOccurred())

	if libgooptions.TestOptions.Options.Hub.KubeConfig == "" {
		if kubeconfig == "" {
			kubeconfig = os.Getenv("KUBECONFIG")
		}
		libgooptions.TestOptions.Options.Hub.KubeConfig = kubeconfig
	}

	if libgooptions.TestOptions.Options.Hub.BaseDomain != "" {
		baseDomain = libgooptions.TestOptions.Options.Hub.BaseDomain

		if libgooptions.TestOptions.Options.Hub.MasterURL == "" {
			libgooptions.TestOptions.Options.Hub.MasterURL = fmt.Sprintf("https://api.%s.%s:6443", libgooptions.TestOptions.Options.Hub.Name, libgooptions.TestOptions.Options.Hub.BaseDomain)
		}
	} else {
		Expect(baseDomain).NotTo(BeEmpty(), "The `baseDomain` is required.")
		libgooptions.TestOptions.Options.Hub.BaseDomain = baseDomain
		libgooptions.TestOptions.Options.Hub.MasterURL = fmt.Sprintf("https://api.%s.%s:6443", libgooptions.TestOptions.Options.Hub.Name, baseDomain)
	}

	if libgooptions.TestOptions.Options.Hub.User != "" {
		kubeadminUser = libgooptions.TestOptions.Options.Hub.User
	}
	if libgooptions.TestOptions.Options.Hub.Password != "" {
		kubeadminCredential = libgooptions.TestOptions.Options.Hub.Password
	}

	if libgooptions.TestOptions.Options.ManagedClusters != nil && len(libgooptions.TestOptions.Options.ManagedClusters) > 0 {
		for i, mc := range libgooptions.TestOptions.Options.ManagedClusters {
			if mc.MasterURL == "" {
				libgooptions.TestOptions.Options.ManagedClusters[i].MasterURL = fmt.Sprintf("https://api.%s:6443", mc.BaseDomain)
			}
			if mc.KubeConfig == "" {
				libgooptions.TestOptions.Options.ManagedClusters[i].KubeConfig = os.Getenv("IMPORT_KUBECONFIG")
			}
		}
	}

	return nil
}

func waitClusterImported(hubClientDynamic dynamic.Interface, clusterName string) {
	Eventually(func() error {
		klog.V(1).Infof("Cluster %s: Wait %s to be imported...", clusterName, clusterName)
		return checkClusterImported(hubClientDynamic, clusterName)
	}).Should(BeNil())
	klog.V(1).Infof("Cluster %s: imported", clusterName)
}

func checkClusterImported(hubClientDynamic dynamic.Interface, clusterName string) error {
	klog.V(1).Infof("Cluster %s: Check %s is imported...", clusterName, clusterName)
	gvr := schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1", Resource: "managedclusters"}
	managedCluster, err := hubClientDynamic.Resource(gvr).Get(context.TODO(), clusterName, metav1.GetOptions{})
	if err != nil {
		klog.V(4).Info(err)
		return err
	}
	var condition map[string]interface{}
	condition, err = libgounstructuredv1.GetConditionByType(managedCluster, "ManagedClusterConditionAvailable")
	if err != nil {
		return err
	}
	klog.V(4).Infof("Cluster %s: Condition %#v", clusterName, condition)
	if v, ok := condition["status"]; ok && v == string(metav1.ConditionTrue) {
		return nil
	} else {
		klog.V(4).Infof("Cluster %s: Current is not equal to \"%s\" but \"%v\"", clusterName, metav1.ConditionTrue, v)
		return fmt.Errorf("status is %s", v)
	}
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

func waitDetached(hubClientDynamic dynamic.Interface, clusterName string) {
	By(fmt.Sprintf("Checking the deletion of the %s managedCluster on the hub", clusterName), func() {
		klog.V(1).Infof("Cluster %s: Checking the deletion of the %s managedCluster on the hub", clusterName, clusterName)
		gvr := schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1", Resource: "managedclusters"}
		Eventually(func() bool {
			klog.V(1).Infof("Cluster %s: Wait %s managedCluster deletion...", clusterName, clusterName)
			_, err := hubClientDynamic.Resource(gvr).Get(context.TODO(), clusterName, metav1.GetOptions{})
			if err != nil {
				klog.V(4).Infof("Cluster %s: %s", clusterName, err)
				return errors.IsNotFound(err)
			}
			return false
		}).Should(BeTrue())
		klog.V(1).Infof("Cluster %s: %s managedCluster deleted", clusterName, clusterName)
	})
}

func validateClusterImported(hubClientDynamic dynamic.Interface, hubClient kubernetes.Interface, clusterName string) {
	gvr := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeployments"}
	clusterDeployment, err := hubClientDynamic.Resource(gvr).Namespace(clusterName).Get(context.TODO(), clusterName, metav1.GetOptions{})
	Expect(err).To(BeNil())
	var configSecretRef string
	if si, ok := clusterDeployment.Object["spec"]; ok {
		s := si.(map[string]interface{})
		if ci, ok := s["clusterMetadata"]; ok {
			c := ci.(map[string]interface{})
			if ai, ok := c["adminKubeconfigSecretRef"]; ok {
				a := ai.(map[string]interface{})
				if ni, ok := a["name"]; ok {
					configSecretRef = ni.(string)
				}
			}
		}
	}
	if configSecretRef == "" {
		Fail(fmt.Sprintf("adminKubeconfigSecretRef.name not found in clusterDeployment %s", clusterName))
	}
	s, err := hubClient.CoreV1().Secrets(clusterName).Get(context.TODO(), configSecretRef, metav1.GetOptions{})
	Expect(err).To(BeNil())
	config, err := clientcmd.Load(s.Data["kubeconfig"])
	Expect(err).To(BeNil())
	rconfig, err := clientcmd.NewDefaultClientConfig(
		*config,
		&clientcmd.ConfigOverrides{}).ClientConfig()
	Expect(err).To(BeNil())
	By("Checking if \"open-cluster-management-agent\" namespace on managed cluster exists", func() {
		clientset, err := kubernetes.NewForConfig(rconfig)
		Expect(err).To(BeNil())
		_, err = clientset.CoreV1().Namespaces().Get(context.TODO(), "open-cluster-management-agent", metav1.GetOptions{})
		Expect(err).To(BeNil())
		klog.V(1).Info("\"open-cluster-management-agent\" namespace on managed cluster exists")
	})
	By("Checking if \"klusterlet\" on managed cluster exits", func() {
		gvr := schema.GroupVersionResource{Group: "operator.open-cluster-management.io", Version: "v1", Resource: "klusterlets"}
		clientDynamic, err := dynamic.NewForConfig(rconfig)
		Expect(err).To(BeNil())
		_, err = clientDynamic.Resource(gvr).Get(context.TODO(), "klusterlet", metav1.GetOptions{})
		Expect(err).To(BeNil())
		klog.V(1).Info("klusterlet on managed cluster exists")
	})
}

func validateClusterDetached(hubClientDynamic dynamic.Interface, hubClient kubernetes.Interface, clusterName string) {
	gvr := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeployments"}
	clusterDeployment, err := hubClientDynamic.Resource(gvr).Namespace(clusterName).Get(context.TODO(), clusterName, metav1.GetOptions{})
	Expect(err).To(BeNil())
	var configSecretRef string
	if si, ok := clusterDeployment.Object["spec"]; ok {
		s := si.(map[string]interface{})
		if ci, ok := s["clusterMetadata"]; ok {
			c := ci.(map[string]interface{})
			if ai, ok := c["adminKubeconfigSecretRef"]; ok {
				a := ai.(map[string]interface{})
				if ni, ok := a["name"]; ok {
					configSecretRef = ni.(string)
				}
			}
		}
	}
	if configSecretRef == "" {
		Fail(fmt.Sprintf("adminKubeconfigSecretRef.name not found in clusterDeployment %s", clusterName))
	}
	s, err := hubClient.CoreV1().Secrets(clusterName).Get(context.TODO(), configSecretRef, metav1.GetOptions{})
	Expect(err).To(BeNil())
	config, err := clientcmd.Load(s.Data["kubeconfig"])
	Expect(err).To(BeNil())
	rconfig, err := clientcmd.NewDefaultClientConfig(
		*config,
		&clientcmd.ConfigOverrides{}).ClientConfig()
	Expect(err).To(BeNil())
	By("Checking if \"klusterlet\" on managed cluster is deleted", func() {
		gvr := schema.GroupVersionResource{Group: "operator.open-cluster-management.io", Version: "v1", Resource: "klusterlets"}
		clientDynamic, err := dynamic.NewForConfig(rconfig)
		Expect(err).To(BeNil())
		_, err = clientDynamic.Resource(gvr).Get(context.TODO(), "klusterlet", metav1.GetOptions{})
		deleted := false
		if err != nil {
			klog.V(4).Info(err)
			deleted = errors.IsNotFound(err)
		}
		Expect(deleted).To(BeTrue())
		klog.V(1).Info("klusterlet on managed cluster deleted")
	})
	By("Checking if \"klusterlet CRD\" on managed cluster is deleted", func() {
		clientset, err := clientset.NewForConfig(rconfig)
		Expect(err).To(BeNil())
		has, _, _ := libgocrdv1.HasCRDs(clientset,
			[]string{
				"klusterlets.operator.open-cluster-management.io",
			})
		Expect(has).To(BeFalse())
		klog.V(1).Info("klusterlet CRD on managed cluster deleted")
	})
	By("Checking if \"open-cluster-management-agent\" namespace on managed cluster is deleted", func() {
		clientset, err := kubernetes.NewForConfig(rconfig)
		Expect(err).To(BeNil())
		_, err = clientset.CoreV1().Namespaces().Get(context.TODO(), "open-cluster-management-agent", metav1.GetOptions{})
		deleted := false
		if err != nil {
			klog.V(4).Info(err)
			deleted = errors.IsNotFound(err)
		}
		Expect(deleted).To(BeTrue())
		klog.V(1).Info("\"open-cluster-management-agent\" namespace on managed cluster deleted")
	})
}

func waitClusterAdddonsAvailable(hubClientDynamic dynamic.Interface, clusterName string) {
	//gvr := schema.GroupVersionResource{Group: "addon.open-cluster-management.io", Version: "v1alpha1", Resource: "managedclusteraddons"}
	for _, addOnName := range managedClusteraddOns {
		if !(clusterName == "local-cluster" && addOnName == "search-collector") {
			Eventually(func() error {
				klog.V(1).Infof("Cluster %s: Checking Add-On %s is available...", clusterName, addOnName)
				return validateClusterAddOnAvailable(hubClientDynamic, clusterName, addOnName)
			}).Should(BeNil())
			klog.V(1).Infof("Cluster %s: all add-ons are available", clusterName)
		}
	}
}

func validateClusterAddOnAvailable(hubClientDynamic dynamic.Interface, clusterName string, addOnName string) error {

	gvr := schema.GroupVersionResource{Group: "addon.open-cluster-management.io", Version: "v1alpha1", Resource: "managedclusteraddons"}
	managedClusterAddon, err := hubClientDynamic.Resource(gvr).Namespace(clusterName).Get(context.TODO(), addOnName, metav1.GetOptions{})
	Expect(err).To(BeNil())

	var condition map[string]interface{}
	condition, err = libgounstructuredv1.GetConditionByType(managedClusterAddon, "Available")
	if err != nil {
		klog.V(4).Infof("Cluster %s - Add-On %s: %s", clusterName, addOnName, err)
		return err
	}
	klog.V(4).Info(condition)
	if v, ok := condition["status"]; ok && v == string(metav1.ConditionTrue) {
		klog.V(1).Infof("Cluster %s: Add-On %s is available...", clusterName, addOnName)
		return nil
	}
	err = fmt.Errorf("Cluster %s - Add-On %s: status not found or not true", clusterName, addOnName)
	klog.V(4).Infof("Cluster %s - Add-On %s: %s", clusterName, addOnName, err)
	return err

}
