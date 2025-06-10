package templates

import (
	"os"
	"strings"
)


type EnvironmentTemplateData struct {
	EnvironmentVariables map[string]string
}
func NewEnvironmentTemplateData(envVars map[string]string) *EnvironmentTemplateData {
	return &EnvironmentTemplateData{
		EnvironmentVariables: envVars,
	}
}

func EnvironmentVariableTemplate(template string) string {
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			// fmt.Printf("Adding env var: %s=%s\n", parts[0], parts[1])
			envVars[parts[0]] = parts[1]
		}
	}
	result := NewTemplate("templated-envvar", template, NewEnvironmentTemplateData(envVars))
	rendered, err := result.Render()
	if err != nil {
		return template
	}
	return rendered
}
