// parser/mermaid_parser.go
package parser

import (
	"bufio"
	"mchat/models"
	"regexp"
	"strings"
)

type MermaidParser struct {
	nodeRegex     *regexp.Regexp
	edgeRegex     *regexp.Regexp
	classDefRegex *regexp.Regexp
	classRegex    *regexp.Regexp
	debug         bool
}

func NewMermaidParser() *MermaidParser {
	return &MermaidParser{
		nodeRegex:     regexp.MustCompile(`(\w+)\["([^"]+)"\]`),
		edgeRegex:     regexp.MustCompile(`(\w+)\s*-->\s*(\||\{)?\s*(\w+)`),
		classDefRegex: regexp.MustCompile(`classDef\s+(\w+)\s+fill:([^,]+),stroke:([^,]+)`),
		classRegex:    regexp.MustCompile(`class\s+([\w+,]+)\s+(\w+)`),
	}
}

func (p *MermaidParser) Parse(content string) (*models.Graph, error) {
	graph := &models.Graph{
		Nodes: []models.Node{},
		Edges: []models.Edge{},
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过注释和空行
		if strings.HasPrefix(line, "%%") || line == "" {
			continue
		}

		// 解析节点
		if matches := p.nodeRegex.FindStringSubmatch(line); matches != nil {
			nodeID := matches[1]
			nodeContent := matches[2]
			node := p.parseNode(nodeID, nodeContent)
			graph.Nodes = append(graph.Nodes, node)
			continue
		}

		// 解析边
		if matches := p.edgeRegex.FindStringSubmatch(line); matches != nil {
			sourceID := matches[1]
			targetID := matches[3]
			edge := models.Edge{
				SourceID: sourceID,
				TargetID: targetID,
				Label:    strings.TrimSpace(matches[2]),
			}
			graph.Edges = append(graph.Edges, edge)
		}
	}

	return graph, nil
}

func (p *MermaidParser) parseNode(id, content string) models.Node {
	node := models.Node{
		ID: id,
		Metadata: models.Metadata{
			Properties: make(map[string]interface{}),
		},
	}

	// 解析节点类型和内容
	parts := strings.Split(content, "<br/>")
	if len(parts) > 0 {
		p.parseNodeHeader(parts[0], &node)
	}

	if len(parts) > 1 {
		p.parseNodeBody(parts[1:], &node)
	}

	return node
}

func (p *MermaidParser) parseNodeHeader(header string, node *models.Node) {
	// 根据ID前缀判断类型
	switch {
	case strings.HasPrefix(node.ID, "Q_"):
		node.Type = models.NodeQuestion
		node.PatternID = strings.TrimPrefix(node.ID, "Q_")
	case strings.HasPrefix(node.ID, "A_"):
		node.Type = models.NodeAnswer
		node.ResponseID = strings.TrimPrefix(node.ID, "A_")
	case strings.HasPrefix(node.ID, "C_"):
		node.Type = models.NodeCondition
	case strings.HasPrefix(node.ID, "S_"):
		node.Type = models.NodeVariable
	case strings.HasPrefix(node.ID, "T_"):
		node.Type = models.NodeTopic
	default:
		node.Type = models.NodeQuestion // 默认
	}

	// 解析头部内容
	if strings.Contains(header, ":") {
		parts := strings.SplitN(header, ":", 2)
		node.Metadata.Pattern = strings.TrimSpace(parts[1])
	} else {
		node.Metadata.Pattern = header
	}
}

func (p *MermaidParser) parseNodeBody(body []string, node *models.Node) {
	for _, line := range body {
		switch node.Type {
		case models.NodeAnswer:
			node.Metadata.Response = line
		case models.NodeVariable:
			if strings.HasPrefix(line, "set:") {
				p.parseSetAction(line, node)
			}
		case models.NodeCondition:
			node.Metadata.Condition = line
		case models.NodeTopic:
			if strings.HasPrefix(line, "topic:") {
				node.Metadata.Topic = strings.TrimPrefix(line, "topic:")
			}
		}
	}
}

func (p *MermaidParser) parseSetAction(line string, node *models.Node) {
	// 解析 set:user_name=$1 格式
	setPart := strings.TrimPrefix(line, "set:")
	parts := strings.SplitN(setPart, "=", 2)
	if len(parts) == 2 {
		node.Metadata.Variable = parts[0]
		node.Metadata.Value = parts[1]
	}
}
