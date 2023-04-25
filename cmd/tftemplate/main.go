package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"tftemplate/models"

	"github.com/AlecAivazis/survey/v2"
)

//go:embed resource.go.tmpl
var resourceTemplateFile string

//go:embed resource_test.go.tmpl
var resourceTestTemplateFile string

var resourceQS = []*survey.Question{
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
}

func main() {
	resourceInput := struct {
		API      string
		Resource string
		Locality string
	}{}
	err := survey.Ask(resourceQS, &resourceInput)
	if err != nil {
		log.Fatalln(err)
	}
	resourceData := models.NewResourceTemplate(resourceInput.API, resourceInput.Resource, resourceInput.Locality)

	resourceTemplate, err := template.New("resource").Parse(resourceTemplateFile)
	if err != nil {
		log.Fatalln(err)
	}
	resourceTestTemplate, err := template.New("resource").Parse(resourceTestTemplateFile)
	if err != nil {
		log.Fatalln(err)
	}
	resourceFile, err := os.Create(fmt.Sprintf("../../scaleway/resource_%s.go", resourceData.ResourceHCL))
	if err != nil {
		log.Fatalln(err)
	}
	defer resourceFile.Close()
	err = resourceTemplate.Execute(resourceFile, resourceData)
	if err != nil {
		log.Println(err)
		return
	}

	resourceTestFile, err := os.Create(fmt.Sprintf("../../scaleway/resource_%s_test.go", resourceData.ResourceHCL))
	if err != nil {
		log.Println(err)
		return
	}
	defer resourceTestFile.Close()
	err = resourceTestTemplate.Execute(resourceTestFile, resourceData)
	if err != nil {
		log.Println(err)
		return
	}
}
