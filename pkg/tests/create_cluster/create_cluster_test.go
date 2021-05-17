// Copyright (c) 2020 Red Hat, Inc.

package create_cluster

import (
	. "github.com/onsi/ginkgo"
	"github.com/open-cluster-management/cluster-lifecycle-e2e/pkg/utils"
)

var _ = Describe("Cluster-lifecycle: ", func() {
	utils.CreateCluster("aws", "OpenShift", cloudProviders)
})

var _ = Describe("Cluster-lifecycle: ", func() {
	utils.CreateCluster("azure", "OpenShift", cloudProviders)
})

var _ = Describe("Cluster-lifecycle: ", func() {
	utils.CreateCluster("gcp", "OpenShift", cloudProviders)
})
