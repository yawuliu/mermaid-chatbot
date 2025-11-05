package kernel

import (
	"mchat/models"
	"regexp"
	"strings"
)

type Matcher struct {
	patterns  map[string]models.Pattern
	templates map[string]models.Template
	graph     *models.Graph
}

func NewMatcher(graph *models.Graph) *Matcher {
	m := &Matcher{
		patterns:  make(map[string]models.Pattern),
		templates: make(map[string]models.Template),
		graph:     graph,
	}
	m.extractPatternsAndTemplates()
	return m
}

func (m *Matcher) Match(input string, ctx *models.Context) (*models.Template, map[string]string) {
	// 1. 清理输入
	cleanedInput := strings.TrimSpace(input)

	// 2. 查找匹配的模式
	for _, pattern := range m.patterns {
		if matches, captures := m.matchPattern(pattern, cleanedInput); matches {
			// 3. 找到对应的回答模板
			if template, exists := m.templates[pattern.ID]; exists {
				return &template, captures
			}
		}
	}

	return nil, nil
}

func (m *Matcher) matchPattern(pattern models.Pattern, input string) (bool, map[string]string) {
	captures := make(map[string]string)

	switch pattern.Type {
	case models.MatchExact:
		for _, content := range pattern.Content {
			if input == content {
				return true, captures
			}
		}

	case models.MatchContains:
		for _, content := range pattern.Content {
			if strings.Contains(input, content) {
				return true, captures
			}
		}

	case models.MatchPattern:
		for _, content := range pattern.Content {
			// 将AIML模式转换为正则表达式
			regexPattern := m.convertAIMLPattern(content)
			if regexPattern.MatchString(input) {
				// 提取捕获组
				matches := regexPattern.FindStringSubmatch(input)
				for i, name := range regexPattern.SubexpNames() {
					if i > 0 && i <= len(matches) && name != "" {
						captures[name] = matches[i]
					}
				}
				return true, captures
			}
		}

	case models.MatchRegex:
		if pattern.Regex != nil && pattern.Regex.MatchString(input) {
			return true, captures
		}
	}

	return false, captures
}

func (m *Matcher) convertAIMLPattern(pattern string) *regexp.Regexp {
	// 将AIML模式转换为正则表达式
	// * -> (.*)
	// _ -> (\w+)
	regexPattern := strings.ReplaceAll(pattern, "*", `(.*)`)
	regexPattern = strings.ReplaceAll(regexPattern, "_", `(\w+)`)
	regexPattern = "^" + regexPattern + "$"

	return regexp.MustCompile(regexPattern)
}

func (m *Matcher) extractPatternsAndTemplates() {
	for _, node := range m.graph.Nodes {
		switch node.Type {
		case models.NodeQuestion:
			pattern := models.Pattern{
				ID:       node.PatternID,
				Type:     m.detectMatchType(node.Metadata.Pattern),
				Content:  []string{node.Metadata.Pattern},
				Priority: node.Metadata.Priority,
			}
			m.patterns[pattern.ID] = pattern

		case models.NodeAnswer:
			template := models.Template{
				ID:      node.ResponseID,
				Content: node.Metadata.Response,
			}
			m.templates[template.ID] = template
		}
	}
}

func (m *Matcher) detectMatchType(pattern string) models.MatchType {
	if strings.Contains(pattern, "*") || strings.Contains(pattern, "_") {
		return models.MatchPattern
	}
	if len(pattern) < 10 { // 简单启发式规则
		return models.MatchExact
	}
	return models.MatchContains
}
