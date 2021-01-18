// Copyright (c) 2020 Red Hat, Inc.

// +build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	metricQueryURI = "api/v1/query?query="
	metricName     = "acm_managed_cluster_info"
	// metricName           = "alertmanager_alerts"
	prometheusServiceURL = "https://prometheus-k8s-openshift-monitoring.apps"
)

type queryResponse struct {
	Status string `json:"status"`
	Data   data   `json:"data"`
}

type data struct {
	ResultType string   `json:"resultType`
	Result     []result `json:"result"`
}

type result struct {
	Metric metric        `json:"metric"`
	Value  []interface{} `json:"value"`
}

type metric struct {
	Name     string `json:"__name__"`
	Job      string `json:"job"`
	Instance string `json:"instance"`
}

var prometheusQueryURL string

var _ = Describe("Cluster-lifecycle: [P2][Sev1][cluster-lifecycle] Check metrics", func() {
	BeforeEach(func() {
		prometheusQueryURL = fmt.Sprintf("%s.%s/%s", prometheusServiceURL, baseDomain, metricQueryURI)
		SetDefaultEventuallyTimeout(1 * time.Minute)
		SetDefaultEventuallyPollingInterval(10 * time.Second)
	})

	It("Check if local-cluster metrics are available  (cluster/g0/metrics)", func() {
		clusterName := "local-cluster"
		klog.V(1).Infof("========================= Test cluster metrics hub %s ===============================", clusterName)
		By(fmt.Sprintf("Checking cluster %s to be ready", clusterName), func() {
			waitClusterImported(hubClientDynamic, clusterName)
		})

		var clusterID string
		By("Getting the managed cluster info", func() {
			gvr := schema.GroupVersionResource{
				Group:    "internal.open-cluster-management.io",
				Version:  "v1beta1",
				Resource: "managedclusterinfos",
			}
			managedClusterInfo, err := hubClientDynamic.Resource(gvr).Namespace(clusterName).Get(context.TODO(), clusterName, metav1.GetOptions{})
			Expect(err).To(BeNil())
			if v, ok := managedClusterInfo.Object["status"]; ok {
				status := v.(map[string]interface{})
				klog.V(2).Infof("status found: %s", status)
				if v, ok := status["clusterID"]; ok {
					clusterID = v.(string)
					klog.V(2).Infof("cloudID found: %s", clusterID)
				}
			}
			Expect(clusterID).ShouldNot(Equal(""))
		})
		By("Getting metrics", func() {
			Eventually(func() error {
				query := "sum(acm_managed_cluster_info{hub_cluster_id=\"" +
					clusterID + "\",cluster_id=\"" + clusterID + "\"})"
				klog.V(1).Infof("Querying metric expression:%s", query)
				resp, b, err := getMetricsQuery(query)
				if err != nil {
					klog.V(2).Infof("err: %s", err)
					return err
				}
				if resp.StatusCode != http.StatusOK {
					klog.V(2).Infof("StatusCode: %d", resp.StatusCode)
					return err
				}
				klog.V(5).Infof("body:\n%s", b)
				qr := &queryResponse{}
				err = json.Unmarshal(b, qr)
				if err != nil {
					klog.V(2).Infof("err: %s", err)
					return err
				}
				if qr.Status != "success" {
					err = fmt.Errorf("Expected status success got %s", qr.Status)
					klog.V(2).Infof("err: %s", err)
					return err
				}
				if qr.Data.Result[0].Value[1].(string) != "1" {
					err = fmt.Errorf("Expected value 1 got %s", qr.Data.Result[0].Value[1].(string))
					klog.V(2).Infof("err: %s", err)
					return err
				}
				return nil
			}).Should(BeNil())
		})

	})

})

func getMetricsQuery(queryExpression string) (resp *http.Response, body []byte, err error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", prometheusQueryURL, queryExpression), nil)
	if err != nil {
		return
	}

	bearerToken := hubRestConfig.BearerToken
	req.Header.Add("Authorization", "Bearer "+bearerToken)

	client := http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	return
}
