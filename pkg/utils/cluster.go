package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	corev1 "k8s.io/api/core/v1"

	"github.com/stolostron/applier/pkg/applier"
	"github.com/stolostron/applier/pkg/templateprocessor"
	"github.com/stolostron/cluster-lifecycle-e2e/pkg/appliers"
	"github.com/stolostron/cluster-lifecycle-e2e/pkg/clients"
	libgooptions "github.com/stolostron/library-e2e-go/pkg/options"
	libgocrdv1 "github.com/stolostron/library-go/pkg/apis/meta/v1/crd"
	libgodeploymentv1 "github.com/stolostron/library-go/pkg/apis/meta/v1/deployment"
	libgounstructuredv1 "github.com/stolostron/library-go/pkg/apis/meta/v1/unstructured"
)

// list of manifestwork name for addon crs
var managedClusteraddOns = []string{
	"application-manager",
	"cert-policy-controller",
	"iam-policy-controller",
	"cert-policy-controller",
	"governance-policy-framework",
	"search-collector",
	"work-manager",
}

var (
	eventuallyTimeout  = 600
	eventuallyInterval = 10
)

const (
	NeedInvestigate              = "[need investigate]"
	KnownIssueTag                = "[known issue]"
	DetachKnownIssueLink         = "https://github.com/stolostron/cluster-lifecycle-e2e/blob/main/doc/e2eFailedAnalysis.md#klusterlet-crd-can-not-be-deleted"
	QuotaLimitTag                = "[quota limit]"
	ProvisionQuotaLimitErrorLink = "https://github.com/stolostron/cluster-lifecycle-e2e/blob/main/doc/e2eFailedAnalysis.md#quota-limit-in-awsazuregcp"
	UnknownError                 = "[unknown error]"
	ProvisionUnknownErrorLink    = "https://github.com/stolostron/cluster-lifecycle-e2e/blob/main/doc/e2eFailedAnalysis.md#cloud-providerawsgcpazure-bug-or-ocp-installer-bug"

	UnknownErrorLink = "https://github.com/stolostron/cluster-lifecycle-e2e/blob/main/doc/e2eFailedAnalysis.md#unknown-error"
	// List key word about quota limit which can not be identified in clusterdeployment.

	// failed to fetch dependency of \"Cluster\": failed to generate asset \"Platform Quota Check\": error(MissingQuota): compute.googleapis.com/firewalls is not available in global because the required number of resources (6) is more than remaining quota of 0\n,",
	gcpQuotaLimitMsg = "more than remaining quota"
)

func WaitClusterImported(hubClientDynamic dynamic.Interface, clusterName string) {
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
		return GenerateErrorMsg(UnknownError, UnknownErrorLink, "Import cluster fail, Cluster status in unknown", "Import cluster fail, Cluster status in unknown")
	}
}

func CreateCluster(cloud, vendor, cloudProviders string) {
	var clusterNameObj *libgooptions.ClusterName
	var clusterName string
	var err error
	var imageRefName string
	var hubAppliers *appliers.HubAppliers
	var hubClients *clients.HubClients

	BeforeEach(func() {
		hubClients = clients.GetHubClients()
		hubAppliers = appliers.GetHubAppliers(hubClients)
		if cloudProviders != "" && !isRequestedCloudProvider(cloud, cloudProviders) {
			Skip(fmt.Sprintf("Cloud provider %s skipped", cloud))
		}
		clusterNameObj, err = libgooptions.NewClusterName(cloud)
		Expect(err).To(BeNil())
		clusterName = clusterNameObj.String()
		if cloud == "baremetal" {
			clusterName = libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.ClusterName
		}
		klog.V(1).Infof(`========================= Start Test create cluster %s 
with image %s ===============================`, clusterName, imageRefName)
		SetDefaultEventuallyTimeout(10 * time.Minute)
		SetDefaultEventuallyPollingInterval(10 * time.Second)
	})

	AfterEach(func() {

	})

	It(fmt.Sprintf("[P1][Sev1][cluster-lifecycle] Create cluster %s on %s with vendor %s (cluster/g1/create-cluster)", clusterName, cloud, vendor), func() {
		By("Checking the minimal requirements", func() {
			klog.V(1).Infof("Cluster %s: Checking the minimal requirements", clusterName)
			Eventually(func() bool {
				klog.V(2).Infof("Cluster %s: Check CRDs", clusterName)
				has, missing, _ := libgocrdv1.HasCRDs(hubClients.APIExtensionClient,
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
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
					"multicluster-engine",
					[]string{
						"managedcluster-import-controller-v2",
					})
				if !has {
					klog.Errorf("Cluster %s: Missing deployments\n%#v", clusterName, missing)
				}
				return has
			}).Should(BeTrue())

			Eventually(func() bool {
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
					"open-cluster-management",
					[]string{
						"klusterlet-addon-controller-v2",
					})
				if !has {
					klog.Errorf("Cluster %s: Missing deployments\n%#v", clusterName, missing)
				}
				return has
			}).Should(BeTrue())
			Eventually(func() bool {
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
					"open-cluster-management-hub",
					[]string{"cluster-manager-registration-controller"})
				if !has {
					klog.Errorf("Cluster %s: Missing deployments\n%#v", clusterName, missing)
				}
				return has
			}).Should(BeTrue())
			Eventually(func() bool {
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
					"hive",
					[]string{"hive-controllers"})
				if !has {
					klog.Errorf("Missing deployments\n%#v", missing)
				}
				return has
			}).Should(BeTrue())

		})

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

		By("Creating the needed resources", func() {
			klog.V(1).Infof("Cluster %s: Creating the needed resources", clusterName)
			pullSecret := &corev1.Secret{}
			Expect(hubClients.ClientClient.Get(context.TODO(),
				types.NamespacedName{
					Name:      "pull-secret",
					Namespace: "openshift-config",
				},
				pullSecret)).To(BeNil())
			values := struct {
				ManagedClusterName          string
				ManagedClusterCloud         string
				ManagedClusterVendor        string
				ManagedClusterSSHPrivateKey string
				ManagedClusterPullSecret    string
			}{
				ManagedClusterName:          clusterName,
				ManagedClusterCloud:         cloud,
				ManagedClusterVendor:        vendor,
				ManagedClusterSSHPrivateKey: libgooptions.TestOptions.Options.CloudConnection.SSHPrivateKey,
				ManagedClusterPullSecret:    string(pullSecret.Data[".dockerconfigjson"]),
			}
			Expect(hubAppliers.CreateApplier.CreateOrUpdateInPath(".",
				[]string{
					"install_config_secret_cr.yaml",
					"cluster_deployment_cr.yaml",
					"managed_cluster_cr.yaml",
					"klusterlet_addon_config_cr.yaml",
					"clusterimageset_cr.yaml",
				},
				false,
				values)).To(BeNil())

			if cloud != "baremetal" {
				klog.V(1).Infof("Cluster %s: Creating the %s cred secret", clusterName, cloud)
				Expect(createCredentialsSecret(hubAppliers.CreateApplier, clusterName, cloud)).To(BeNil())
			}

			klog.V(1).Infof("Cluster %s: Creating install config secret", clusterName)
			Expect(createInstallConfig(hubAppliers.CreateApplier, hubAppliers.CreateTemplateProcessor, clusterName, cloud)).To(BeNil())

			// imageRefName = libgooptions.TestOptions.ManagedClusters.ImageSetRefName

			if libgooptions.TestOptions.Options.OCPReleaseVersion != "" && cloud != "baremetal" {
				imageRefName, err = createClusterImageSet(hubAppliers.CreateApplier, clusterNameObj, libgooptions.TestOptions.Options.OCPReleaseVersion)
				Expect(err).To(BeNil())
				// imageRefName = libgooptions.TestOptions.Options.OCPReleaseVersion
			} else {
				gvr := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterimagesets"}
				imagesetsList, err := hubClients.DynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
				Expect(err).To(BeNil())

				for _, imageset := range imagesetsList.Items {
					if metadata, ok := imageset.Object["metadata"]; ok {
						if m, ok := metadata.(map[string]interface{}); ok {
							if name, ok := m["name"]; ok {
								strName := fmt.Sprintf("%v", name)
								klog.V(1).Infof("Cluster %s: Add imageset: %s", clusterName, strName)
								if len(imageRefName) == 0 {
									imageRefName = strName
									continue
								}
								// get the max version to deploy
								if compareImageVersion(imageRefName, strName) < 0 {
									imageRefName = strName
								}
							}
						}
					}
				}
			}
			if libgooptions.TestOptions.Options.OCPReleaseVersion != "" && cloud == "baremetal" {
				imageRefName = libgooptions.TestOptions.Options.OCPReleaseVersion
			}
			Expect(imageRefName).NotTo(Equal(""))
		})

		By("creating the clusterDeployment", func() {
			var region string
			if cloud != "baremetal" {
				region, err = libgooptions.GetRegion(cloud)
				Expect(err).To(BeNil())
			}
			baseDomain, err := libgooptions.GetBaseDomain(cloud)
			Expect(err).To(BeNil())
			values := struct {
				ManagedClusterName          string
				ManagedClusterCloud         string
				ManagedClusterRegion        string
				ManagedClusterVendor        string
				ManagedClusterBaseDomain    string
				ManagedClusterImageRefName  string
				ManagedClusterBaseDomainRGN string
				SSHKnownHosts               []string
				Hosts                       []libgooptions.Hosts
			}{
				ManagedClusterName:       clusterName,
				ManagedClusterCloud:      cloud,
				ManagedClusterRegion:     region,
				ManagedClusterVendor:     vendor,
				ManagedClusterBaseDomain: baseDomain,
				// TODO: parametrize the image
				ManagedClusterImageRefName:  imageRefName,
				ManagedClusterBaseDomainRGN: libgooptions.TestOptions.Options.CloudConnection.APIKeys.Azure.BaseDomainRGN,
				SSHKnownHosts:               libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.SSHKnownHostsList,
				Hosts:                       libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.Hosts,
			}
			klog.V(1).Infof("Cluster %s: Creating the clusterDeployment", clusterName)
			Expect(hubAppliers.CreateApplier.CreateOrUpdateResource("cluster_deployment_cr.yaml", values)).To(BeNil())
		})

		By("Attaching the cluster by creating the managedCluster and klusterletaddonconfig", func() {
			createManagedCluster(hubAppliers.CreateApplier, clusterName, cloud, vendor)
			createKlusterletAddonConfig(hubAppliers.CreateApplier, clusterName, cloud, vendor)
		})

		When("Import launched, wait for cluster to be installed", func() {
			gvr := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeployments"}
			Eventually(func() error {
				klog.V(1).Infof("Cluster %s: Wait %s to be installed...", clusterName, clusterName)
				clusterDeployment, err := hubClients.DynamicClient.Resource(gvr).Namespace(clusterName).Get(context.TODO(), clusterName, metav1.GetOptions{})
				if err == nil {
					if si, ok := clusterDeployment.Object["status"]; ok {
						s := si.(map[string]interface{})
						if ti, ok := s["installedTimestamp"]; ok && ti != nil {
							return nil
						}
					}
					var condition map[string]interface{}
					condition, err = libgounstructuredv1.GetConditionByType(clusterDeployment, "ProvisionFailed")
					if err == nil {
						if v, ok := condition["status"]; ok && v == string(metav1.ConditionTrue) {
							if strings.HasSuffix(condition["reason"].(string), "LimitExceeded") {
								return GenerateErrorMsg(QuotaLimitTag, ProvisionQuotaLimitErrorLink, condition["reason"].(string), condition["message"].(string))
							}

							if strings.Contains(condition["message"].(string), gcpQuotaLimitMsg) {
								return GenerateErrorMsg(QuotaLimitTag, ProvisionQuotaLimitErrorLink, condition["reason"].(string), condition["message"].(string))
							}

							if condition["reason"].(string) == "UnknownError" {
								return GenerateErrorMsg(UnknownError, ProvisionUnknownErrorLink, condition["reason"].(string), condition["message"].(string))
							}
							return GenerateErrorMsg("", "", condition["reason"].(string), condition["message"].(string))
						}
					}
					return fmt.Errorf("Failed to get provision result.")
				} else {
					klog.V(4).Info(err)
				}

				return err
			}, 5400, 60).Should(BeNil())
		})

		When(fmt.Sprintf("Import launched, wait for cluster %s to be ready", clusterName), func() {
			waitClusterImported(hubClients.DynamicClient, clusterName)
		})

		if cloud != "baremetal" {
			When("Imported, validate...", func() {
				validateClusterImported(hubClients.DynamicClient, hubClients.KubeClient, clusterName)
			})
		}

		klog.V(1).Infof("Cluster %s: Wait 3 min to settle", clusterName)
		time.Sleep(3 * time.Minute)

		if cloud != "baremetal" {
			When(fmt.Sprintf("Import launched, wait for Add-Ons %s to be available", clusterName), func() {
				WaitClusterAdddonsAvailable(hubClients.DynamicClient, clusterName)
			})
		}

		klog.V(1).Infof("========================= End Test create cluster %s ===============================", clusterName)

	})

}

func GenerateErrorMsg(tag, solution, reason, errmsg string) error {
	return fmt.Errorf("Tag: %v, "+
		"Possible Solution: %v, "+
		"Reason: %v, "+
		"Error message: %v,",
		tag, solution, reason, errmsg)
}

// compareImageVersion returns an integer comparing two strings lexicographically.
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
// imageVersion format like img4.6.3-x86-64-appsub
func compareImageVersion(a, b string) int {
	strA := strings.ToLower(a)
	strB := strings.ToLower(b)

	switch {
	case strings.Compare(strA, strB) == 0:
		return 0
	case strings.HasPrefix(strA, "img") && !strings.HasPrefix(strB, "img"):
		return 1
	case !strings.HasPrefix(strA, "img") && strings.HasPrefix(strB, "img"):
		return -1
	}

	subStrA := strings.Split(strA, "-")
	subStrB := strings.Split(strB, "-")

	versionA := strings.Trim(subStrA[0], "img")
	versionB := strings.Trim(subStrB[0], "img")

	if strings.Compare(versionA, versionB) == 0 {
		return 0
	}

	subVersionA := strings.Split(versionA, ".")
	subVersionB := strings.Split(versionB, ".")

	lenA := len(subVersionA)
	lenB := len(subVersionB)
	for i := 0; i < lenA && i < lenB; i++ {
		switch {
		case len(subVersionA[i]) > len(subVersionB[i]):
			return 1
		case len(subVersionA[i]) < len(subVersionB[i]):
			return -1
		case len(subVersionA[i]) == len(subVersionB[i]):
			rst := strings.Compare(subVersionA[i], subVersionB[i])
			if rst != 0 {
				return rst
			}
		}
	}
	return 0
}

func isRequestedCloudProvider(cloud, cloudProviders string) bool {
	cloudProviderstSlice := strings.Split(cloudProviders, ",")
	klog.V(5).Infof("cloudProviderSlice %v", cloudProviderstSlice)
	for _, cloudProvider := range cloudProviderstSlice {
		cp := strings.TrimSpace(cloudProvider)
		klog.V(5).Infof("cloudProvider %s", cp)
		if cp == cloud {
			return true
		}
	}
	return false
}

func createCredentialsSecret(hubCreateApplier *applier.Applier, clusterName, cloud string) error {
	switch cloud {
	case "aws":
		cloudCredSecretValues := struct {
			ManagedClusterName string
			AWSAccessKeyID     string
			AWSSecretAccessKey string
		}{
			ManagedClusterName: clusterName,
			AWSAccessKeyID:     libgooptions.TestOptions.Options.CloudConnection.APIKeys.AWS.AWSAccessKeyID,
			AWSSecretAccessKey: libgooptions.TestOptions.Options.CloudConnection.APIKeys.AWS.AWSAccessSecret,
		}
		return hubCreateApplier.CreateOrUpdateResource(filepath.Join(cloud, "creds_secret_cr.yaml"), cloudCredSecretValues)
	case "azure":
		cloudCredSecretValues := struct {
			ManagedClusterName           string
			ManagedClusterClientId       string
			ManagedClusterClientSecret   string
			ManagedClusterTenantId       string
			ManagedClusterSubscriptionId string
		}{
			ManagedClusterName:           clusterName,
			ManagedClusterClientId:       libgooptions.TestOptions.Options.CloudConnection.APIKeys.Azure.ClientID,
			ManagedClusterClientSecret:   libgooptions.TestOptions.Options.CloudConnection.APIKeys.Azure.ClientSecret,
			ManagedClusterTenantId:       libgooptions.TestOptions.Options.CloudConnection.APIKeys.Azure.TenantID,
			ManagedClusterSubscriptionId: libgooptions.TestOptions.Options.CloudConnection.APIKeys.Azure.SubscriptionID,
		}
		return hubCreateApplier.CreateOrUpdateResource(filepath.Join(cloud, "creds_secret_cr.yaml"), cloudCredSecretValues)
	case "gcp":
		cloudCredSecretValues := struct {
			ManagedClusterName      string
			GCPOSServiceAccountJson string
		}{
			ManagedClusterName:      clusterName,
			GCPOSServiceAccountJson: libgooptions.TestOptions.Options.CloudConnection.APIKeys.GCP.ServiceAccountJSONKey,
		}
		return hubCreateApplier.CreateOrUpdateResource(filepath.Join(cloud, "creds_secret_cr.yaml"), cloudCredSecretValues)
	// case "baremetal":
	// 	return fmt.Println("baremetal")
	default:
		return fmt.Errorf("unsupporter cloud %s", cloud)
	}
}

func createInstallConfig(hubCreateApplier *applier.Applier,
	createTemplateProcessor *templateprocessor.TemplateProcessor,
	clusterName,
	cloud string) error {
	baseDomain, err := libgooptions.GetBaseDomain(cloud)
	if err != nil {
		return err
	}

	var region string
	if cloud != "baremetal" {
		region, err = libgooptions.GetRegion(cloud)
		if err != nil {
			return err
		}
	}

	var b []byte
	switch cloud {
	case "aws":
		installConfigValues := struct {
			ManagedClusterName         string
			ManagedClusterBaseDomain   string
			ManagedClusterRegion       string
			ManagedClusterSSHPublicKey string
		}{
			ManagedClusterName:         clusterName,
			ManagedClusterBaseDomain:   baseDomain,
			ManagedClusterRegion:       region,
			ManagedClusterSSHPublicKey: libgooptions.TestOptions.Options.CloudConnection.SSHPublicKey,
		}
		b, err = createTemplateProcessor.TemplateResource(filepath.Join(cloud, "install_config.yaml"), installConfigValues)
	case "azure":
		installConfigValues := struct {
			ManagedClusterName          string
			ManagedClusterBaseDomain    string
			ManagedClusterBaseDomainRGN string
			ManagedClusterRegion        string
			ManagedClusterSSHPublicKey  string
		}{
			ManagedClusterName:          clusterName,
			ManagedClusterBaseDomain:    baseDomain,
			ManagedClusterBaseDomainRGN: libgooptions.TestOptions.Options.CloudConnection.APIKeys.Azure.BaseDomainRGN,
			ManagedClusterRegion:        region,
			ManagedClusterSSHPublicKey:  libgooptions.TestOptions.Options.CloudConnection.SSHPublicKey,
		}
		b, err = createTemplateProcessor.TemplateResource(filepath.Join(cloud, "install_config.yaml"), installConfigValues)
	case "gcp":
		installConfigValues := struct {
			ManagedClusterName         string
			ManagedClusterBaseDomain   string
			ManagedClusterProjectID    string
			ManagedClusterRegion       string
			ManagedClusterSSHPublicKey string
		}{
			ManagedClusterName:         clusterName,
			ManagedClusterBaseDomain:   baseDomain,
			ManagedClusterProjectID:    libgooptions.TestOptions.Options.CloudConnection.APIKeys.GCP.ProjectID,
			ManagedClusterRegion:       region,
			ManagedClusterSSHPublicKey: libgooptions.TestOptions.Options.CloudConnection.SSHPublicKey,
		}
		b, err = createTemplateProcessor.TemplateResource(filepath.Join(cloud, "install_config.yaml"), installConfigValues)
	case "baremetal":
		installConfigValues := struct {
			ManagedClusterName             string
			ManagedClusterBaseDomain       string
			LibvirtURI                     string
			ProvisioningNetworkCIDR        string
			ProvisioningNetworkInterface   string
			ProvisioningBridge             string
			ExternalBridge                 string
			APIVIP                         string
			IngressVIP                     string
			ManagedClusterBootstrapOSImage string
			ManagedClusterClusterOSImage   string
			ManagedClusterSSHPublicKey     string
			ManagedClusterTrustBundle      string
			ImageRegistryMirror            string
			Hosts                          []libgooptions.Hosts
		}{
			ManagedClusterName:             clusterName,
			ManagedClusterBaseDomain:       baseDomain,
			LibvirtURI:                     libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.LibvirtURI,
			ProvisioningNetworkCIDR:        libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.ProvisioningNetworkCIDR,
			ProvisioningNetworkInterface:   libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.ProvisioningNetworkInterface,
			ProvisioningBridge:             libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.ProvisioningBridge,
			ExternalBridge:                 libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.ExternalBridge,
			APIVIP:                         libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.APIVIP,
			IngressVIP:                     libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.IngressVIP,
			ManagedClusterBootstrapOSImage: libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.BootstrapOSImage,
			ManagedClusterClusterOSImage:   libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.ClusterOSImage,
			ManagedClusterSSHPublicKey:     libgooptions.TestOptions.Options.CloudConnection.SSHPublicKey,
			ManagedClusterTrustBundle:      libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.TrustBundle,
			ImageRegistryMirror:            libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.ImageRegistryMirror,
			Hosts:                          libgooptions.TestOptions.Options.CloudConnection.APIKeys.BareMetal.Hosts,
		}

		b, err = createTemplateProcessor.TemplateResource(filepath.Join(cloud, "install_config.yaml"), installConfigValues)
	default:
		return fmt.Errorf("unsupporter cloud %s", cloud)

	}
	if err != nil {
		return err
	}
	installConfigSecretValues := struct {
		ManagedClusterName          string
		ManagedClusterInstallConfig string
	}{
		ManagedClusterName:          clusterName,
		ManagedClusterInstallConfig: base64.StdEncoding.EncodeToString(b),
	}
	return hubCreateApplier.CreateOrUpdateResource("install_config_secret_cr.yaml", installConfigSecretValues)
}

func createKlusterletAddonConfig(hubCreateApplier *applier.Applier, clusterName, cloud, vendor string) {
	By("creating the klusterletaddonconfig", func() {
		values := struct {
			ManagedClusterName   string
			ManagedClusterCloud  string
			ManagedClusterVendor string
		}{
			ManagedClusterName:   clusterName,
			ManagedClusterCloud:  cloud,
			ManagedClusterVendor: vendor,
		}
		klog.V(1).Infof("Cluster %s: Creating the klusterletaddonconfig", clusterName)
		Expect(hubCreateApplier.CreateOrUpdateResource("klusterlet_addon_config_cr.yaml", values)).To(BeNil())
	})
}

func createClusterImageSet(hubCreateApplier *applier.Applier, clusterNameObj *libgooptions.ClusterName, ocpImageRelease string) (string, error) {
	ocpImageReleaseSlice := strings.Split(ocpImageRelease, ":")
	if len(ocpImageReleaseSlice) != 2 {
		return "", fmt.Errorf("OCPImageRelease malformed: %s (no tag)", ocpImageRelease)
	}
	normalizedOCPImageRelease := strings.ReplaceAll(ocpImageReleaseSlice[1], "_", "-")
	normalizedOCPImageRelease = strings.ToLower(normalizedOCPImageRelease)
	clusterImageSetName := fmt.Sprintf("%s-%s", normalizedOCPImageRelease, clusterNameObj.GetUID())
	values := struct {
		ClusterImageSetName string
		OCPReleaseImage     string
	}{
		ClusterImageSetName: clusterImageSetName,
		OCPReleaseImage:     ocpImageRelease,
	}
	klog.V(1).Infof("Cluster %s: Creating the imageSetName %s", clusterNameObj, clusterImageSetName)
	err := hubCreateApplier.CreateOrUpdateResource("clusterimageset_cr.yaml", values)
	if err != nil {
		return "", err
	}
	return clusterImageSetName, nil
}

func createManagedCluster(hubCreateApplier *applier.Applier, clusterName, cloud, vendor string) {
	By("creating the managedCluster and klusterletaddonconfig", func() {
		values := struct {
			ManagedClusterName   string
			ManagedClusterCloud  string
			ManagedClusterVendor string
		}{
			ManagedClusterName:   clusterName,
			ManagedClusterCloud:  cloud,
			ManagedClusterVendor: vendor,
		}
		klog.V(1).Infof("Cluster %s: Creating the managedCluster", clusterName)
		Expect(hubCreateApplier.CreateOrUpdateResource("managed_cluster_cr.yaml", values)).To(BeNil())
	})
}

func waitDetroyed(hubClientDynamic dynamic.Interface, clusterName string) {
	By(fmt.Sprintf("Checking the deletion of the %s clusterDeployment on the hub", clusterName), func() {
		klog.V(1).Infof("Cluster %s: Checking the deletion of the %s clusterDeployment on the hub", clusterName, clusterName)
		gvr := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeployments"}
		Eventually(func() bool {
			klog.V(1).Infof("Cluster %s: Wait %s clusterDeployment deletion...", clusterName, clusterName)
			_, err := hubClientDynamic.Resource(gvr).Namespace(clusterName).Get(context.TODO(), clusterName, metav1.GetOptions{})
			if err != nil {
				klog.V(1).Info(err)
				return errors.IsNotFound(err)
			}
			return false
		}, 3600, 60).Should(BeTrue())
		klog.V(1).Infof("Cluster %s: %s clusterDeployment deleted", clusterName, clusterName)
	})
}

func waitClusterImported(hubClientDynamic dynamic.Interface, clusterName string) {
	Eventually(func() error {
		klog.V(1).Infof("Cluster %s: Wait %s to be imported...", clusterName, clusterName)
		return checkClusterImported(hubClientDynamic, clusterName)
	}, eventuallyTimeout, eventuallyInterval).Should(BeNil())
	klog.V(1).Infof("Cluster %s: imported", clusterName)
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

func WaitClusterAdddonsAvailable(hubClientDynamic dynamic.Interface, clusterName string) {
	// gvr := schema.GroupVersionResource{Group: "addon.open-cluster-management.io", Version: "v1alpha1", Resource: "managedclusteraddons"}
	for _, addOnName := range managedClusteraddOns {
		if !(clusterName == "local-cluster" && addOnName == "search-collector") {
			Eventually(func() error {
				klog.V(1).Infof("Cluster %s: Checking Add-On %s is available...", clusterName, addOnName)
				return validateClusterAddOnAvailable(hubClientDynamic, clusterName, addOnName)
			}, eventuallyTimeout, eventuallyInterval).Should(BeNil())
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
	err = fmt.Errorf("cluster %s - Add-On %s: status not found or not true", clusterName, addOnName)
	klog.V(4).Infof("Cluster %s - Add-On %s: %s", clusterName, addOnName, err)
	return err

}

func DestroyCluster(cloud, vendor, cloudProviders string) {
	var clusterName string
	var hubClients *clients.HubClients

	BeforeEach(func() {

		if cloudProviders != "" && !isRequestedCloudProvider(cloud, cloudProviders) {
			Skip(fmt.Sprintf("Cloud provider %s skipped", cloud))
		}

		hubClients = clients.GetHubClients()
		gvrClusterDeployment := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeployments"}
		clusterDeploymentList, err := hubClients.DynamicClient.Resource(gvrClusterDeployment).List(context.TODO(), metav1.ListOptions{})
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
				has, missing, _ := libgocrdv1.HasCRDs(hubClients.APIExtensionClient,
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
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
					"multicluster-engine",
					[]string{"managedcluster-import-controller-v2"})
				if !has {
					klog.Errorf("Cluster %s: Missing deployments\n%#v", clusterName, missing)
				}
				return has
			}).Should(BeTrue())
			Eventually(func() bool {
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
					"open-cluster-management-hub",
					[]string{"cluster-manager-registration-controller"})
				if !has {
					klog.Errorf("Cluster %s: Missing deployments\n%#v", clusterName, missing)
				}
				return has
			}).Should(BeTrue())
			Eventually(func() bool {
				has, missing, _ := libgodeploymentv1.HasDeploymentsInNamespace(hubClients.KubeClient,
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
			Expect(hubClients.DynamicClient.Resource(gvr).Delete(context.TODO(), clusterName, metav1.DeleteOptions{})).Should(BeNil())
		})

		When(fmt.Sprintf("Detached, delete the clusterDeployment %s", clusterName), func() {
			klog.V(1).Infof("Cluster %s: Deleting the clusterDeployment for cluster %s", clusterName, clusterName)
			gvr := schema.GroupVersionResource{Group: "hive.openshift.io", Version: "v1", Resource: "clusterdeployments"}
			Expect(hubClients.DynamicClient.Resource(gvr).Namespace(clusterName).Delete(context.TODO(), clusterName, metav1.DeleteOptions{})).Should(BeNil())
		})

		When(fmt.Sprintf("Wait clusterDeployment %s to be deleted", clusterName), func() {
			waitDetroyed(hubClients.DynamicClient, clusterName)
		})

		When(fmt.Sprintf("Wait namespace %s to be deleted", clusterName), func() {
			waitNamespaceDeleted(hubClients.KubeClient, hubClients.DynamicClient, hubClients.DiscoveryClient, clusterName)
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
			err = PrintLeftOver(hubClientDynamic, hubClientDiscovery, clusterName)
			if err != nil {
				klog.Error(err)
			}
			return false
		}, 3600, 60).Should(BeTrue())
		klog.V(1).Infof("Cluster %s: %s namespace deleted", clusterName, clusterName)
	})
}

func PrintLeftOver(dynamicClient dynamic.Interface, discoveryClient *discovery.DiscoveryClient, ns string) error {
	klog.Infof("==================== Left Over in %s ======================", ns)
	_, err := dynamicClient.Resource(schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}).Get(context.TODO(), ns, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		} else {
			return err
		}
	}
	apiResourceLists, err := discoveryClient.ServerPreferredNamespacedResources()
	if err != nil {
		return err
	}
	for _, apiResourceList := range apiResourceLists {
		for _, apiResource := range apiResourceList.APIResources {
			us, err := dynamicClient.Resource(
				schema.GroupVersionResource{
					Group:    apiResourceList.GroupVersionKind().Group,
					Version:  apiResourceList.GroupVersion,
					Resource: apiResource.Name,
				}).Namespace(ns).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					klog.Errorf("Group: %s, Version: %s, Resource: %s Error: %s",
						apiResourceList.GroupVersionKind().Group,
						apiResourceList.GroupVersion,
						apiResource.Name, err)
				}
				continue
			}
			for _, u := range us.Items {
				klog.Infof("Kind: %s, Name: %s", u.GetKind(), u.GetName())
			}
		}
	}
	return nil
}
