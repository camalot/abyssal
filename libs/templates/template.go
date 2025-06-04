package templates

import (
	"bytes"
	"text/template"
)

type Template struct {
	Data     interface{}
	Name     string
	Template string
}

func NewTemplate(name string, templateString string, data interface{}) *Template {
  return &Template{
    Name:     name,
    Template: templateString,
    Data:     data,
  }
}

func (t *Template) Render() (string, error) {
	tmpl, err := template.New(t.Name).Parse(t.Template)
	if err != nil {
		return "", err
	}

	var renderedString bytes.Buffer
	err = tmpl.Execute(&renderedString, t.Data)
	if err != nil {
		return "", err
	}
	return renderedString.String(), nil
}
