// parser/mermaid_parser.go
package parser

import (
	"fmt"
	"log"
	"mchat/model"
	"regexp"
	"strings"
)

type MermaidParser struct {
	debug bool
}

func ParseMermaid(content string) ([]model.Category, error) {
	parser := &MermaidParser{debug: true}
	return parser.Parse(content)
}

// ParseMermaidWithSource 解析Mermaid内容并设置来源（供外部调用）
func ParseMermaidWithSource(content, sourceFile string) ([]model.Category, error) {
	categories, err := ParseMermaid(content)
	if err != nil {
		return nil, err
	}

	// 设置来源文件
	for i := range categories {
		categories[i].SourceFile = sourceFile
	}

	return categories, nil
}

func (p *MermaidParser) Parse(content string) ([]model.Category, error) {
	lines := strings.Split(content, "\n")
	graph := NewGraph()

	inFlowchart := false
	lineCount := 0

	if p.debug {
		log.Printf("开始解析Mermaid内容，共 %d 行", len(lines))
	}

	for _, line := range lines {
		lineCount++
		line = strings.TrimSpace(line)

		if p.debug {
			log.Printf("行 %d: '%s'", lineCount, line)
		}

		// 跳过空行和注释
		if p.isComment(line) || p.isEmpty(line) {
			if p.debug {
				log.Printf("  跳过注释或空行")
			}
			continue
		}

		// 检测流程图开始
		if strings.HasPrefix(line, "flowchart") || strings.HasPrefix(line, "graph") {
			inFlowchart = true
			if p.debug {
				log.Printf("  检测到流程图开始")
			}
			continue
		}

		if !inFlowchart {
			if p.debug {
				log.Printf("  不在流程图范围内，跳过")
			}
			continue
		}

		// 解析节点和边
		if p.containsNode(line) {
			if p.debug {
				log.Printf("  包含节点定义")
			}
			p.parseNode(line, graph)
		}
		if p.containsEdge(line) {
			if p.debug {
				log.Printf("  包含边定义")
			}
			p.parseEdge(line, graph)
		}
	}

	if p.debug {
		log.Printf("解析完成，找到 %d 个节点", len(graph.Nodes))
		for id, node := range graph.Nodes {
			log.Printf("  节点 %s: %s", id, node.Text)
		}
		log.Printf("找到 %d 条边", len(graph.Edges))
		for from, toList := range graph.Edges {
			log.Printf("  边 %s -> %v", from, toList)
		}
	}

	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("未在文件中找到有效的节点")
	}

	return p.buildCategories(graph), nil
}

// 更新节点检测逻辑
func (p *MermaidParser) containsNode(line string) bool {
	nodePattern := `\w+\[[^\]]+\]`
	matched, _ := regexp.MatchString(nodePattern, line)
	return matched
}

// 更新边检测逻辑
func (p *MermaidParser) containsEdge(line string) bool {
	return strings.Contains(line, "-->")
}

// 增强节点解析
func (p *MermaidParser) parseNode(line string, graph *Graph) {
	// 匹配类似: A[你好] 或 A[P:你好]
	re := regexp.MustCompile(`(\w+)\[([^\]]+)\]`)
	matches := re.FindAllStringSubmatch(line, -1)

	for _, match := range matches {
		if len(match) == 3 {
			nodeID := match[1]
			nodeText := match[2]
			if p.debug {
				log.Printf("    解析到节点: ID=%s, Text=%s", nodeID, nodeText)
			}
			graph.AddNode(nodeID, nodeText)
		}
	}
}

// 增强边解析
func (p *MermaidParser) parseEdge(line string, graph *Graph) {
	// 匹配多种边格式:
	// A --> B
	// A[你好] --> B[你好！]
	// A --> B[你好！]
	re := regexp.MustCompile(`(\w+)(?:\[[^\]]*\])?\s*-->\s*(\w+)(?:\[[^\]]*\])?`)
	matches := re.FindAllStringSubmatch(line, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			from := match[1]
			to := match[2]
			if p.debug {
				log.Printf("    解析到边: %s --> %s", from, to)
			}
			graph.AddEdge(from, to)

			// 如果边定义中包含了节点文本，也添加到图中
			if strings.Contains(line, from+"[") {
				p.parseNode(line, graph)
			}
		}
	}
}

// 其余方法保持不变...
func (p *MermaidParser) isComment(line string) bool {
	return strings.HasPrefix(line, "%%")
}

func (p *MermaidParser) isEmpty(line string) bool {
	return line == ""
}

func (p *MermaidParser) buildCategories(graph *Graph) []model.Category {
	categories := make([]model.Category, 0)

	// 找出所有起始节点（没有入边的节点）
	startNodes := graph.FindStartNodes()

	if p.debug {
		log.Printf("构建类别，起始节点: %v", startNodes)
	}

	for _, startNode := range startNodes {
		// 对于每个起始节点，找到所有直接连接的响应节点
		responses := graph.GetResponses(startNode)
		if len(responses) > 0 {
			category := model.NewCategory(
				graph.Nodes[startNode].Text,
				responses,
			)
			if p.debug {
				log.Printf("  创建类别: Pattern='%s', Responses=%v", category.Pattern, category.Templates)
			}
			categories = append(categories, category)
		}
	}

	return categories
}
