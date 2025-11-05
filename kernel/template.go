package kernel

import (
	"mchat/models"
	"regexp"
	"strings"
)

type TemplateEngine struct {
	variables map[string]string
}

func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		variables: make(map[string]string),
	}
}

func (te *TemplateEngine) Render(template *models.Template, captures map[string]string, ctx *models.Context) string {
	result := template.Content

	// 1. 替换变量引用 <get var>
	result = te.replaceGetVariables(result, ctx)

	// 2. 替换捕获组 $1, $2
	result = te.replaceCaptures(result, captures)

	// 3. 执行动作
	for _, action := range template.Actions {
		te.executeAction(action, ctx)
	}

	return result
}

func (te *TemplateEngine) replaceGetVariables(template string, ctx *models.Context) string {
	getRegex := regexp.MustCompile(`<get\s+(\w+)>`)
	return getRegex.ReplaceAllStringFunc(template, func(match string) string {
		varName := getRegex.FindStringSubmatch(match)[1]
		if value, exists := ctx.Variables[varName]; exists {
			return value
		}
		return match // 保持原样
	})
}

func (te *TemplateEngine) replaceCaptures(template string, captures map[string]string) string {
	result := template
	for key, value := range captures {
		placeholder := "$" + key
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

func (te *TemplateEngine) executeAction(action models.Action, ctx *models.Context) {
	switch action.Type {
	case "set":
		ctx.Variables[action.Key] = action.Value
	case "topic":
		ctx.CurrentTopic = action.Value
	}
}
