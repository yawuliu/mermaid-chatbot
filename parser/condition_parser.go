// parser/condition_parser.go
package parser

import (
	"log"
	"regexp"
	"strings"

	"mchat/model"
)

// ConditionParser 条件解析器
type ConditionParser struct {
	debug bool
}

func (p *ConditionParser) isComment(line string) bool {
	return strings.HasPrefix(line, "%%")
}

func (p *ConditionParser) isEmpty(line string) bool {
	return line == ""
}
func (p *ConditionParser) isEdge(line string) bool {
	return strings.Contains(line, "-->")
}

// ParseConditionalMermaid 解析带条件的Mermaid内容
func ParseConditionalMermaid(content string) ([]model.ConditionalCategory, error) {
	parser := &ConditionParser{debug: true}
	return parser.Parse(content)
}

// Parse 解析Mermaid条件流程图
func (cp *ConditionParser) Parse(content string) ([]model.ConditionalCategory, error) {
	lines := strings.Split(content, "\n")
	graph := NewGraph()

	inFlowchart := false
	lineNumber := 0

	for _, line := range lines {
		lineNumber++
		line = strings.TrimSpace(line)

		if cp.isComment(line) || cp.isEmpty(line) {
			continue
		}

		if strings.HasPrefix(line, "flowchart") || strings.HasPrefix(line, "graph") {
			inFlowchart = true
			log.Printf("检测到流程图开始")
			continue
		}

		if !inFlowchart {
			continue
		}

		log.Printf("行 %d: %s", lineNumber, line)

		// 改进的解析顺序
		if cp.isEdgeLine(line) {
			log.Printf("  识别为边")
			cp.parseConditionEdge(line, graph)
		} else if cp.isPurePatternNode(line) {
			log.Printf("  识别为独立模式节点")
			cp.parsePatternNode(line, graph)
		} else if cp.isPureConditionNode(line) {
			log.Printf("  识别为独立条件节点")
			cp.parseConditionNode(line, graph)
		} else {
			log.Printf("  无法识别的行类型")
		}
	}

	// 解析完成后打印图状态
	cp.printGraphState(graph)
	// 构建前的图状态日志

	// 构建条件分类
	categories := cp.buildConditionalCategories(graph)
	log.Printf("构建的分类数: %d", len(categories))

	return categories, nil
}

// printGraphState 打印图的状态用于调试
func (cp *ConditionParser) printGraphState(graph *Graph) {
	log.Printf("=== 图状态 ===")
	log.Printf("节点总数: %d", len(graph.Nodes))

	for id, node := range graph.Nodes {
		nodeType := "未知"
		if node.IsPattern {
			nodeType = "模式"
		} else if node.IsCondition {
			nodeType = "条件"
		}

		log.Printf("  节点 %s: 类型=%s, 文本=%s", id, nodeType, node.Text)
		if node.IsCondition {
			log.Printf("    条件类型: %s, 条件名称: %s", node.ConditionType, node.ConditionName)
		}
	}

	log.Printf("边总数: %d", len(graph.Edges))
	for from, toList := range graph.Edges {
		for _, to := range toList {
			conditionValue := graph.EdgeConditions[from][to]
			log.Printf("  边: %s -> %s (条件值: '%s')", from, to, conditionValue)
		}
	}
	log.Printf("==============")
}

// isConditionNode 检查是否为条件节点
func (cp *ConditionParser) isConditionNode(line string) bool {
	return strings.Contains(line, "Condition:") && strings.Contains(line, "{")
}

// isPatternNode 检查是否为模式节点
func (cp *ConditionParser) isPatternNode(line string) bool {
	pattern := `\w+\[[^\]]+\]`
	matched, _ := regexp.MatchString(pattern, line)
	return matched && !cp.isConditionNode(line)
}

// isPurePatternNode 检查是否为纯粹的模式节点（不包含条件语法）
func (cp *ConditionParser) isPurePatternNode(line string) bool {
	// 匹配: A[text] 且不包含箭头
	pattern := `^\s*\w+\[[^\]]+\]\s*$`
	matched, _ := regexp.MatchString(pattern, line)
	return matched
}

// isPureConditionNode 检查是否为纯粹的条件节点
func (cp *ConditionParser) isPureConditionNode(line string) bool {
	// 匹配: B{Condition: ...} 且不包含箭头
	pattern := `^\s*\w+\{Condition:[^}]+\}\s*$`
	matched, _ := regexp.MatchString(pattern, line)
	return matched
}

// isEdgeLine 检查是否为边定义行
func (cp *ConditionParser) isEdgeLine(line string) bool {
	// 检查是否包含箭头
	if !strings.Contains(line, "-->") {
		return false
	}

	// 排除明显的独立节点
	if cp.isPurePatternNode(line) || cp.isPureConditionNode(line) {
		return false
	}

	return true
}

// parseConditionNode 解析条件节点
func (cp *ConditionParser) parseConditionNode(line string, graph *Graph) {
	// 匹配: B{Condition: predicate time}
	re := regexp.MustCompile(`(\w+)\{Condition:\s*(\w+)\s+(\w+)\}`)
	matches := re.FindStringSubmatch(line)

	if len(matches) == 4 {
		nodeID := matches[1]
		conditionType := matches[2] // predicate 或 variable
		conditionName := matches[3]

		graph.AddNode(nodeID, "")
		graph.Nodes[nodeID].IsCondition = true
		graph.Nodes[nodeID].ConditionType = conditionType
		graph.Nodes[nodeID].ConditionName = conditionName

		if cp.debug {
			log.Printf("解析条件节点: %s, 类型: %s, 名称: %s",
				nodeID, conditionType, conditionName)
		}
	} else if len(matches) > 0 {
		if cp.debug {
			log.Printf("len(matches): %d", len(matches))
		}
	}
}

// parseConditionEdge 改进的边解析方法
func (cp *ConditionParser) parseConditionEdge(line string, graph *Graph) {
	// 尝试匹配带标签的边：A -->|value| B 或 A[pattern] -->|value| B[response]
	reLabeled := regexp.MustCompile(`(\w+)(?:\[[^\]]*\])?\s*-->\s*\|([^|]+)\|\s*(\w+)(?:\[[^\]]*\])?`)
	matches := reLabeled.FindStringSubmatch(line)
	if matches != nil {
		from := matches[1]
		conditionValue := strings.TrimSpace(matches[2])
		to := matches[3]
		graph.AddEdge(from, to)
		if graph.EdgeConditions[from] == nil {
			graph.EdgeConditions[from] = make(map[string]string)
		}
		graph.EdgeConditions[from][to] = conditionValue
		log.Printf("  解析带标签边: %s -> %s (条件: %s)", from, to, conditionValue)

		// 尝试从边中提取节点信息
		cp.extractNodesFromEdge(line, graph)
		return
	}

	// 新增：匹配无标签的边：A --> B 或 A[pattern] --> B{condition}
	reUnlabeled := regexp.MustCompile(`(\w+)(?:\[([^\]]*)\])?\s*-->\s*(\w+)(?:\{([^}]+)\})?`)
	matches = reUnlabeled.FindStringSubmatch(line)
	if matches != nil {
		from := matches[1]
		patternText := matches[2] // 可能为空
		to := matches[3]
		conditionText := matches[4] // 可能为空

		// 如果边中包含模式文本，添加到图中
		if patternText != "" {
			graph.AddNode(from, patternText)
			graph.Nodes[from].IsPattern = true
			log.Printf("  从边中提取模式节点: %s[%s]", from, patternText)
		} else {
			// 如果没有模式文本，确保节点存在
			if _, exists := graph.Nodes[from]; !exists {
				graph.AddNode(from, "")
			}
		}

		// 如果边中包含条件文本，添加到图中
		if conditionText != "" && strings.Contains(conditionText, "Condition:") {
			graph.AddNode(to, "")
			graph.Nodes[to].IsCondition = true
			// 解析条件类型和名称
			cp.parseConditionDetails(to, conditionText, graph)
			log.Printf("  从边中提取条件节点: %s{%s}", to, conditionText)
		} else {
			// 如果不是条件节点，确保节点存在
			if _, exists := graph.Nodes[to]; !exists {
				graph.AddNode(to, "")
			}
		}

		graph.AddEdge(from, to)
		// 对于无标签的边，设置条件值为空字符串
		if graph.EdgeConditions[from] == nil {
			graph.EdgeConditions[from] = make(map[string]string)
		}
		graph.EdgeConditions[from][to] = "" // 空字符串表示无条件连接
		log.Printf("  解析无标签边: %s -> %s", from, to)
		return
	}

	// 新增：更简单的边匹配，只提取节点ID
	reSimple := regexp.MustCompile(`(\w+)\s*-->\s*(\w+)`)
	matches = reSimple.FindStringSubmatch(line)
	if matches != nil {
		from := matches[1]
		to := matches[2]

		// 确保节点存在
		if _, exists := graph.Nodes[from]; !exists {
			graph.AddNode(from, "")
		}
		if _, exists := graph.Nodes[to]; !exists {
			graph.AddNode(to, "")
		}

		graph.AddEdge(from, to)
		if graph.EdgeConditions[from] == nil {
			graph.EdgeConditions[from] = make(map[string]string)
		}
		graph.EdgeConditions[from][to] = ""
		log.Printf("  解析简单边: %s -> %s", from, to)

		// 尝试从原始行中提取更多信息
		cp.extractNodesFromEdge(line, graph)
		return
	}

	log.Printf("  无法解析的边: %s", line)
}

// extractNodesFromEdge 从边定义中提取节点信息
func (cp *ConditionParser) extractNodesFromEdge(line string, graph *Graph) {
	// 提取模式节点：A[pattern]
	rePattern := regexp.MustCompile(`(\w+)\[([^\]]+)\]`)
	matches := rePattern.FindStringSubmatch(line)
	if matches != nil {
		nodeID := matches[1]
		pattern := matches[2]
		graph.AddNode(nodeID, pattern)
		graph.Nodes[nodeID].IsPattern = true
		log.Printf("    提取模式节点: %s[%s]", nodeID, pattern)
	}

	// 提取条件节点：B{Condition: ...}
	reCondition := regexp.MustCompile(`(\w+)\{Condition:\s*([^}]+)\}`)
	matches = reCondition.FindStringSubmatch(line)
	if matches != nil {
		nodeID := matches[1]
		conditionDetails := strings.TrimSpace(matches[2])
		graph.AddNode(nodeID, "")
		graph.Nodes[nodeID].IsCondition = true
		cp.parseConditionDetails(nodeID, conditionDetails, graph)
		log.Printf("    提取条件节点: %s{Condition: %s}", nodeID, conditionDetails)
	}
}

// parseConditionDetails 解析条件详细信息
func (cp *ConditionParser) parseConditionDetails(nodeID, conditionDetails string, graph *Graph) {
	// 解析条件类型和名称：predicate isanumber
	parts := strings.Fields(conditionDetails)
	if len(parts) >= 2 {
		graph.Nodes[nodeID].ConditionType = parts[0]
		graph.Nodes[nodeID].ConditionName = parts[1]
	} else if len(parts) == 1 {
		graph.Nodes[nodeID].ConditionType = parts[0]
		graph.Nodes[nodeID].ConditionName = "unknown"
	}
}

// parsePatternNode 解析模式节点
func (cp *ConditionParser) parsePatternNode(line string, graph *Graph) {
	re := regexp.MustCompile(`(\w+)\[([^\]]+)\]`)
	matches := re.FindStringSubmatch(line)

	if len(matches) == 3 {
		nodeID := matches[1]
		pattern := matches[2]

		graph.AddNode(nodeID, pattern)
		graph.Nodes[nodeID].IsPattern = true

		if cp.debug {
			log.Printf("解析模式节点: %s, 模式: %s", nodeID, pattern)
		}
	}
}

// buildConditionalCategories 构建条件分类
func (cp *ConditionParser) buildConditionalCategories(graph *Graph) []model.ConditionalCategory {
	var categories []model.ConditionalCategory

	// 找到所有模式节点（起始节点）
	for _, node := range graph.Nodes {
		if node.IsPattern {
			category := cp.buildCategoryFromPattern(node, graph)
			categories = append(categories, category)
		}
	}

	return categories
}

// buildCategoryFromPattern 从模式节点构建分类
func (cp *ConditionParser) buildCategoryFromPattern(patternNode *Node, graph *Graph) model.ConditionalCategory {
	category := model.ConditionalCategory{
		Pattern: patternNode.Text,
	}

	// 查找模式节点连接的条件节点
	for target, _ := range graph.EdgeConditions[patternNode.ID] { // conditionValue
		targetNode := graph.Nodes[target]
		if targetNode.IsCondition {
			// 构建条件回复
			conditionalResponses := cp.buildConditionalResponses(targetNode, graph)
			category.Conditions = append(category.Conditions, conditionalResponses...)
		}
	}

	return category
}

// buildConditionalResponses 构建条件回复
func (cp *ConditionParser) buildConditionalResponses(conditionNode *Node, graph *Graph) []model.ConditionalResponse {
	var responses []model.ConditionalResponse

	// 查找条件节点连接的所有回复节点
	for target, conditionValue := range graph.EdgeConditions[conditionNode.ID] {
		targetNode := graph.Nodes[target]

		condition := model.Condition{
			Type:  cp.mapConditionType(conditionNode.ConditionType),
			Name:  conditionNode.ConditionName,
			Value: conditionValue,
		}

		// 检查是否为默认分支
		if conditionValue == "*" {
			condition.IsDefault = true
		}

		response := model.ConditionalResponse{
			Condition: condition,
			Response:  targetNode.Text,
		}

		responses = append(responses, response)
	}

	return responses
}

// mapConditionType 映射条件类型
func (cp *ConditionParser) mapConditionType(typeStr string) model.ConditionType {
	switch typeStr {
	case "predicate":
		return model.ConditionPredicate
	case "variable":
		return model.ConditionVariable
	case "star":
		return model.ConditionStar
	default:
		return model.ConditionPredicate
	}
}
