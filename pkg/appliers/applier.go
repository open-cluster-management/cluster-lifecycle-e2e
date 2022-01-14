package appliers

import (
	"path/filepath"

	"github.com/onsi/gomega"

	"github.com/stolostron/applier/pkg/applier"
	"github.com/stolostron/applier/pkg/templateprocessor"
	"github.com/stolostron/cluster-lifecycle-e2e/pkg/clients"
)

const (
	importClusterScenario     = "import"
	selfImportClusterScenario = "self_import"
	createClusterScenario     = "create"
)

type HubAppliers struct {
	CreateTemplateProcessor *templateprocessor.TemplateProcessor
	CreateApplier           *applier.Applier
	ImportYamlReader        templateprocessor.TemplateReader
	ImportApplier           *applier.Applier
	SelfImportApplier       *applier.Applier
}

func GetHubAppliers(hubClient *clients.HubClients) (hubAppliers *HubAppliers) {
	var err error
	hubAppliers = &HubAppliers{}
	createYamlReader := templateprocessor.NewYamlFileReader(filepath.Join("../resources/hub", createClusterScenario))
	hubAppliers.CreateTemplateProcessor, err = templateprocessor.NewTemplateProcessor(createYamlReader, &templateprocessor.Options{})
	gomega.Expect(err).To(gomega.BeNil())
	hubAppliers.CreateApplier, err = applier.NewApplier(createYamlReader, &templateprocessor.Options{}, hubClient.ClientClient, nil, nil, nil)
	gomega.Expect(err).To(gomega.BeNil())
	hubAppliers.ImportYamlReader = templateprocessor.NewYamlFileReader(filepath.Join("../resources/hub", importClusterScenario))
	hubAppliers.ImportApplier, err = applier.NewApplier(hubAppliers.ImportYamlReader, &templateprocessor.Options{}, hubClient.ClientClient, nil, nil, nil)
	gomega.Expect(err).To(gomega.BeNil())
	selfImportYamlReader := templateprocessor.NewYamlFileReader(filepath.Join("../resources/hub", selfImportClusterScenario))
	hubAppliers.SelfImportApplier, err = applier.NewApplier(selfImportYamlReader, &templateprocessor.Options{}, hubClient.ClientClient, nil, nil, nil)
	gomega.Expect(err).To(gomega.BeNil())
	return
}
