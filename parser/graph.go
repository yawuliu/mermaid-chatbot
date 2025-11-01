// parser/graph.go
package parser

import "log"

type Node struct {
	ID   string
	Text string
}

type Graph struct {
	Nodes map[string]*Node
	Edges map[string][]string // from -> to nodes
}

func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
		Edges: make(map[string][]string),
	}
}

func (g *Graph) AddNode(id, text string) {
	// 如果节点已存在，更新文本；否则创建新节点
	if existing, exists := g.Nodes[id]; exists {
		log.Printf("警告: 节点 %s 已存在，文本从 '%s' 更新为 '%s'", id, existing.Text, text)
		existing.Text = text
	} else {
		g.Nodes[id] = &Node{
			ID:   id,
			Text: text,
		}
	}
}

func (g *Graph) AddEdge(from, to string) {
	// 确保from和to节点存在（即使还没有文本）
	if _, exists := g.Nodes[from]; !exists {
		g.Nodes[from] = &Node{ID: from, Text: from} // 临时文本
	}
	if _, exists := g.Nodes[to]; !exists {
		g.Nodes[to] = &Node{ID: to, Text: to} // 临时文本
	}

	g.Edges[from] = append(g.Edges[from], to)
}

func (g *Graph) FindStartNodes() []string {
	startNodes := make([]string, 0)
	hasIncoming := make(map[string]bool)

	// 标记所有有入边的节点
	for _, targets := range g.Edges {
		for _, target := range targets {
			hasIncoming[target] = true
		}
	}

	// 起始节点是没有入边的节点，并且有出边
	for nodeID := range g.Nodes {
		if !hasIncoming[nodeID] && len(g.Edges[nodeID]) > 0 {
			startNodes = append(startNodes, nodeID)
		}
	}

	return startNodes
}

func (g *Graph) GetResponses(startNode string) []string {
	responses := make([]string, 0)

	if targets, exists := g.Edges[startNode]; exists {
		for _, target := range targets {
			if node, exists := g.Nodes[target]; exists {
				responses = append(responses, node.Text)
			}
		}
	}

	return responses
}
