// Copyright (c) 2020 Red Hat, Inc.

package detach_destroy

import (
	. "github.com/onsi/ginkgo"
	"github.com/open-cluster-management/cluster-lifecycle-e2e/pkg/utils"
)

var _ = Describe("Cluster-lifecycle: ", func() {
	utils.DestroyCluster("aws", "OpenShift", cloudProviders)
})

var _ = Describe("Cluster-lifecycle: ", func() {
	utils.DestroyCluster("azure", "OpenShift", cloudProviders)
})

var _ = Describe("Cluster-lifecycle: ", func() {
	utils.DestroyCluster("gcp", "OpenShift", cloudProviders)
})
