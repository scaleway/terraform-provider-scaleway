package main

import (
	_ "embed"
	"fmt"
	"log"
	"text/template"

	"tftemplate/models"

	"github.com/AlecAivazis/survey/v2"
)

var (
	//go:embed resource.go.tmpl
	resourceTemplateFile string
	//go:embed resource_test.go.tmpl
	resourceTestTemplateFile string
	//go:embed helpers.go.tmpl
	resourceHelpersTemplateFile string
	//go:embed waiters.go.tmpl
	resourceWaitersTemplateFile string
	//go:embed datasource.go.tmpl
	datasourceTemplateFile string
	//go:embed datasource_test.go.tmpl
	datasourceTestTemplateFile string
	//go:embed sweep_test.go.tmpl
	resourceSweepTemplateFile string
	//go:embed sweep.go.tmpl
	resourceSweepTestTemplateFile string
)

var resourceQS = []*survey.Question{
	{
		Name: "targets",
		Prompt: &survey.MultiSelect{
			Message: "Select targets to generate",
			Options: []string{"resource", "datasource"},
			Default: []string{"resource"},
		},
	},
	{
		Name:   "api",
		Prompt: &survey.Input{Message: "API name (function, instance, container)"},
	},
	{
		Name:      "resource",
		Prompt:    &survey.Input{Message: "Resource name (FunctionNamespace, InstanceServer)"},
		Validate:  survey.Required,
		Transform: survey.Title,
	},
	{
		Name: "locality",
		Prompt: &survey.Select{
			Options: []string{"zone", "region"},
			Default: "zone",
		},
	},
	{
		Name: "helpers",
		Prompt: &survey.Confirm{
			Message: "Generate helpers ? Will override ../../internal/services/{api}/helpers_{api}.go",
			Default: false,
		},
	},
	{
		Name: "waiters",
		Prompt: &survey.Confirm{
			Message: "Generate waiters ? Will be added to ../../internal/services/{api}/waiter.go",
			Default: true,
		},
	},
	{
		Name: "sweep",
		Prompt: &survey.Confirm{
			Message: "Generate sweeper ? Will be added to ../../internal/services/{api}/sweep.go",
			Default: true,
		},
	},
}

func contains[T comparable](slice []T, expected T) bool {
	for _, elem := range slice {
		if elem == expected {
			return true
		}
	}

	return false
}

func main() {
	resourceInput := struct {
		Targets  []string
		API      string
		Resource string
		Locality string
		Helpers  bool
		Waiters  bool
		Sweep    bool
	}{}
	err := survey.Ask(resourceQS, &resourceInput)
	if err != nil {
		log.Fatalln(err)
	}
	resourceData := models.NewResourceTemplate(resourceInput.API, resourceInput.Resource, resourceInput.Locality)
	resourceData.SupportWaiters = resourceInput.Waiters

	templates := []*TerraformTemplate{
		{
			FileName:     fmt.Sprintf("../../internal/services/%s/%s.go", resourceData.API, resourceData.ResourceHCL),
			TemplateFile: resourceTemplateFile,
			Skip:         !contains(resourceInput.Targets, "resource"),
		},
		{
			FileName:     fmt.Sprintf("../../internal/services/%s/%s_test.go", resourceData.API, resourceData.ResourceHCL),
			TemplateFile: resourceTestTemplateFile,
			Skip:         !contains(resourceInput.Targets, "resource"),
		},
		{
			FileName:     fmt.Sprintf("../../internal/services/%s/%s_data_source.go", resourceData.API, resourceData.ResourceHCL),
			TemplateFile: datasourceTemplateFile,
			Skip:         !contains(resourceInput.Targets, "datasource"),
		},
		{
			FileName:     fmt.Sprintf("../../internal/services/%s/data_source_%s_test.go", resourceData.API, resourceData.ResourceHCL),
			TemplateFile: datasourceTestTemplateFile,
			Skip:         !contains(resourceInput.Targets, "datasource"),
		},
		{
			FileName:     fmt.Sprintf("../../internal/services/%s/helpers_%s.go", resourceData.API, resourceData.API),
			TemplateFile: resourceHelpersTemplateFile,
			Skip:         !resourceInput.Helpers,
		},
		{
			FileName:     fmt.Sprintf("../../internal/services/%s/waiter.go", resourceData.API),
			TemplateFile: resourceWaitersTemplateFile,
			Skip:         !resourceInput.Waiters,
			Append:       true,
		},
		{
			FileName:     fmt.Sprintf("../../internal/services/%s/testfuncs/sweep.go", resourceData.API),
			TemplateFile: resourceSweepTemplateFile,
			Skip:         !contains(resourceInput.Targets, "resource"),
			Append:       true,
		},
		{
			FileName:     fmt.Sprintf("../../internal/services/%s/sweep_test.go", resourceData.API),
			TemplateFile: resourceSweepTestTemplateFile,
			Skip:         !contains(resourceInput.Targets, "resource"),
			Append:       true,
		},
	}

	for _, tmpl := range templates {
		if tmpl.Template == nil {
			tmpl.Template, err = template.New(tmpl.FileName).Parse(tmpl.TemplateFile)
			if err != nil {
				log.Fatalln("failed to template " + tmpl.FileName + ":" + err.Error())
			}
		}
	}

	for _, tmpl := range templates {
		if tmpl.Skip {
			continue
		}
		err := executeTemplate(tmpl, resourceData)
		if err != nil {
			log.Println(err)
		}
	}
}
