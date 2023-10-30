package main

import (
	"log"
	"os"
	"text/template"
	"tftemplate/models"
)

type TerraformTemplate struct {
	// Target file for output
	FileName string
	// TemplateFile is a Go template as string
	TemplateFile string
	// Template is a Go template, will be created from TemplateFile if nil
	Template *template.Template
	// Skip template generation if true
	Skip bool
	// Append template output to target if true
	Append bool
}

func executeTemplate(tmpl *TerraformTemplate, data models.ResourceTemplate) error {
	var outputFile *os.File
	var err error

	if tmpl.Append {
		outputFile, err = os.OpenFile(tmpl.FileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		outputFile, err = os.Create(tmpl.FileName)
	}
	if err != nil {
		log.Fatalln(err)
	}
	defer outputFile.Close()

	err = tmpl.Template.Execute(outputFile, data)
	if err != nil {
		return err
	}

	return nil
}
