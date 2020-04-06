package frontend

import (
	"errors"
	"html/template"
	"io/ioutil"
	"path"
	"strings"
)

type TemplateManager struct {
	templatesPath string
	templates     map[string]*template.Template
}

func (templateManager *TemplateManager) Init(templatesPath string) {
	templateManager.templates = make(map[string]*template.Template)
	templateManager.templatesPath = templatesPath
	templateFiles, err := ioutil.ReadDir(templatesPath)
	if err != nil {
		panic("Template manager initialization error: " + err.Error())
	}
	for _, templateFile := range templateFiles {
		if !strings.HasSuffix(templateFile.Name(), ".html") {
			continue
		}
		templateFilename := path.Join(templatesPath, templateFile.Name())
		tmpl, err := template.ParseFiles(templateFilename)
		if err != nil {
			panic("Template manager initialization error: " + err.Error())
		}
		templateManager.templates[templateFilename[0:len(templateFilename)-5]] = tmpl
	}
}

func (templateManager *TemplateManager) GetTemplate(name string) (*template.Template, error) {
	tmpl, ok := templateManager.templates[path.Join(templateManager.templatesPath, name)]
	if !ok {
		return nil, errors.New("can not get template, no such file: '" + name + "'")
	}
	return tmpl, nil
}
