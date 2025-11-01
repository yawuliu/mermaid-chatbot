// parser/advanced_mermaid_writer.go
package parser

import (
	"fmt"
	"mchat/model"
	"sort"
	"strings"
)

// AdvancedMermaidWriter 提供更智能的Mermaid生成，避免重复节点
type AdvancedMermaidWriter struct {
	nodeCounter int
	contentToID map[string]string
	usedIDs     map[string]bool
}

func NewAdvancedMermaidWriter() *AdvancedMermaidWriter {
	return &AdvancedMermaidWriter{
		nodeCounter: 0,
		contentToID: make(map[string]string),
		usedIDs:     make(map[string]bool),
	}
}

// ConvertToOptimizedMermaid 生成优化的Mermaid，合并相同内容的节点
func (w *AdvancedMermaidWriter) ConvertToOptimizedMermaid(categories []model.Category) (string, error) {
	if len(categories) == 0 {
		return "flowchart TD\n    // 没有对话节点", nil
	}

	var builder strings.Builder
	builder.WriteString("flowchart TD\n")

	// 重置状态
	w.nodeCounter = 0
	w.contentToID = make(map[string]string)
	w.usedIDs = make(map[string]bool)

	// 首先收集所有节点
	nodes := w.collectAllNodes(categories)

	// 生成节点定义
	w.generateNodeDefinitions(&builder, nodes)

	// 生成边
	w.generateEdges(&builder, categories)

	return builder.String(), nil
}

// collectAllNodes 收集所有唯一的节点内容
func (w *AdvancedMermaidWriter) collectAllNodes(categories []model.Category) []string {
	nodeSet := make(map[string]bool)

	for _, category := range categories {
		nodeSet[category.Pattern] = true
		for _, template := range category.Templates {
			nodeSet[template] = true
		}
	}

	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}

	// 排序以确保生成稳定的输出
	sort.Strings(nodes)
	return nodes
}

// cleanNodeText 清理节点文本，移除多余的引号并确保格式正确
func (w *AdvancedMermaidWriter) cleanNodeText(text string) string {
	if text == "" {
		return text
	}

	// 移除首尾的多余引号
	cleaned := strings.Trim(text, `"`)

	// 如果文本包含特殊字符或空格，需要引号，但要确保不重复
	if w.needsQuotes(cleaned) {
		// 确保不会变成 """文本""" 这样的形式
		if !strings.HasPrefix(cleaned, `"`) || !strings.HasSuffix(cleaned, `"`) {
			cleaned = `"` + cleaned + `"`
		}
	}

	return cleaned
}

// needsQuotes 检查文本是否需要引号
func (w *AdvancedMermaidWriter) needsQuotes(text string) bool {
	if text == "" {
		return false
	}

	// 如果包含以下字符，需要引号
	specialChars := []string{" ", "-", ">", "<", "[", "]", "(", ")", "{", "}", ",", ".", "!", "?"}
	for _, char := range specialChars {
		if strings.Contains(text, char) {
			return true
		}
	}

	// 如果是纯英文数字，可能不需要引号
	if w.isSimpleText(text) {
		return false
	}

	// 中文文本通常需要引号
	return w.containsChinese(text)
}

// isSimpleText 检查是否为简单的英文数字文本
func (w *AdvancedMermaidWriter) isSimpleText(text string) bool {
	for _, r := range text {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}
	return true
}

// containsChinese 检查是否包含中文字符
func (w *AdvancedMermaidWriter) containsChinese(text string) bool {
	for _, r := range text {
		if r >= '\u4e00' && r <= '\u9fff' {
			return true
		}
	}
	return false
}

// generateNodeDefinitions 生成所有节点的定义
func (w *AdvancedMermaidWriter) generateNodeDefinitions(builder *strings.Builder, nodes []string) {
	for _, content := range nodes {
		nodeID := w.getNodeID(content)
		cleanedContent := w.cleanNodeText(content)
		builder.WriteString(fmt.Sprintf("    %s[%s]\n", nodeID, cleanedContent))
	}
	builder.WriteString("\n")
}

// generateEdges 生成所有边
func (w *AdvancedMermaidWriter) generateEdges(builder *strings.Builder, categories []model.Category) {
	for _, category := range categories {
		patternID := w.getNodeID(category.Pattern)

		for _, template := range category.Templates {
			responseID := w.getNodeID(template)
			builder.WriteString(fmt.Sprintf("    %s --> %s\n", patternID, responseID))
		}
	}
}

// getNodeID 获取节点的ID（相同内容使用相同ID）
func (w *AdvancedMermaidWriter) getNodeID(content string) string {
	if id, exists := w.contentToID[content]; exists {
		return id
	}

	// 生成新ID
	newID := w.generateMeaningfulID(content)
	w.contentToID[content] = newID
	w.usedIDs[newID] = true

	return newID
}

// generateMeaningfulID 生成有意义的节点ID
func (w *AdvancedMermaidWriter) generateMeaningfulID(content string) string {
	// 尝试使用内容的首字母和长度生成ID
	if len(content) == 0 {
		return w.generateFallbackID()
	}

	// 使用前几个字符生成基础ID
	baseID := ""
	for _, char := range content {
		if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') {
			baseID += string(char)
			if len(baseID) >= 3 {
				break
			}
		}
	}

	if baseID == "" {
		return w.generateFallbackID()
	}

	// 确保ID唯一
	candidate := strings.ToUpper(baseID)
	if !w.usedIDs[candidate] {
		return candidate
	}

	// 如果冲突，添加数字后缀
	for i := 1; ; i++ {
		candidate = fmt.Sprintf("%s%d", strings.ToUpper(baseID), i)
		if !w.usedIDs[candidate] {
			return candidate
		}
	}
}

// generateFallbackID 生成回退ID
func (w *AdvancedMermaidWriter) generateFallbackID() string {
	w.nodeCounter++
	return fmt.Sprintf("N%d", w.nodeCounter)
}
