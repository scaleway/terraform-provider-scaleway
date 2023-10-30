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
			Message: "Generate helpers ? Will override scaleway/helpers_{api}.go",
			Default: false,
		},
	},
	{
		Name: "waiters",
		Prompt: &survey.Confirm{
			Message: "Generate waiters ? Will be added to scaleway/helpers_{api}.go",
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
	}{}
	err := survey.Ask(resourceQS, &resourceInput)
	if err != nil {
		log.Fatalln(err)
	}
	resourceData := models.NewResourceTemplate(resourceInput.API, resourceInput.Resource, resourceInput.Locality)

	templates := []*TerraformTemplate{
		{
			FileName:     fmt.Sprintf("../../scaleway/resource_%s.go", resourceData.ResourceHCL),
			TemplateFile: resourceTemplateFile,
			Skip:         !contains(resourceInput.Targets, "resource"),
		},
		{
			FileName:     fmt.Sprintf("../../scaleway/resource_%s_test.go", resourceData.ResourceHCL),
			TemplateFile: resourceTestTemplateFile,
			Skip:         !contains(resourceInput.Targets, "resource"),
		},
		{
			FileName:     fmt.Sprintf("../../scaleway/data_source_%s.go", resourceData.ResourceHCL),
			TemplateFile: datasourceTemplateFile,
			Skip:         !contains(resourceInput.Targets, "datasource"),
		},
		{
			FileName:     fmt.Sprintf("../../scaleway/data_source_%s_test.go", resourceData.ResourceHCL),
			TemplateFile: datasourceTestTemplateFile,
			Skip:         !contains(resourceInput.Targets, "datasource"),
		},
		{
			FileName:     fmt.Sprintf("../../scaleway/helpers_%s.go", resourceData.API),
			TemplateFile: resourceHelpersTemplateFile,
			Skip:         !resourceInput.Helpers,
		},
		{
			FileName:     fmt.Sprintf("../../scaleway/helpers_%s.go", resourceData.API),
			TemplateFile: resourceWaitersTemplateFile,
			Skip:         !resourceInput.Waiters,
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
