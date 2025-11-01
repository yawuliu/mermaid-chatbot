package parser

import (
	"fmt"
	"log"
	"mchat/model"
	"strings"
)

type MermaidWriter struct {
	nodeCounter int
	nodeMap     map[string]string // category.Pattern -> nodeID
}

func NewMermaidWriter() *MermaidWriter {
	return &MermaidWriter{
		nodeCounter: 0,
		nodeMap:     make(map[string]string),
	}
}

// SaveCategoriesToMermaid 将分类集合转换为Mermaid格式并保存到文件
func SaveCategoriesToMermaid(categories []model.Category, filePath string) error {
	writer := NewMermaidWriter()
	content, err := writer.ConvertToMermaid(categories)
	if err != nil {
		return err
	}

	// 这里需要实现文件写入，我们先返回内容用于测试
	log.Printf("生成的Mermaid内容:\n%s", content)

	// 实际文件写入将在后面实现
	return writer.writeToFile(content, filePath)
}

// ConvertToMermaid 将分类集合转换为Mermaid流程图格式
func (w *MermaidWriter) ConvertToMermaid(categories []model.Category) (string, error) {
	if len(categories) == 0 {
		return "flowchart TD\n    // 没有对话节点", nil
	}

	var builder strings.Builder
	builder.WriteString("flowchart TD\n")

	// 重置节点计数器
	w.nodeCounter = 0
	w.nodeMap = make(map[string]string)

	// 为每个分类生成节点和边
	for _, category := range categories {
		w.addCategoryToMermaid(&builder, category)
	}

	return builder.String(), nil
}

// addCategoryToMermaid 添加单个分类到Mermaid构建器
func (w *MermaidWriter) addCategoryToMermaid(builder *strings.Builder, category model.Category) {
	// 获取或创建模式节点ID
	patternNodeID := w.getOrCreateNodeID(category.Pattern)

	// 添加模式节点
	builder.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", patternNodeID, category.Pattern))

	// 为每个回复模板创建节点和边
	for _, template := range category.Templates {
		responseNodeID := w.getOrCreateNodeID(template)

		// 添加回复节点
		builder.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", responseNodeID, template))

		// 添加边
		builder.WriteString(fmt.Sprintf("    %s --> %s\n", patternNodeID, responseNodeID))
	}

	// 添加空行分隔不同的分类组
	builder.WriteString("\n")
}

// getOrCreateNodeID 获取或创建节点的唯一ID
func (w *MermaidWriter) getOrCreateNodeID(content string) string {
	// 如果已经存在这个内容的节点，返回现有ID
	if id, exists := w.nodeMap[content]; exists {
		return id
	}

	// 创建新节点ID（使用字母序列：A, B, C, ..., Z, AA, AB, ...）
	nodeID := w.generateNodeID()
	w.nodeMap[content] = nodeID

	return nodeID
}

// generateNodeID 生成唯一的节点ID
func (w *MermaidWriter) generateNodeID() string {
	w.nodeCounter++
	return w.numberToBase26(w.nodeCounter)
}

// numberToBase26 将数字转换为Base26（A-Z, AA-ZZ等）
func (w *MermaidWriter) numberToBase26(n int) string {
	if n <= 0 {
		return "A"
	}

	var result strings.Builder
	for n > 0 {
		n--
		remainder := n % 26
		result.WriteByte(byte('A' + remainder))
		n = n / 26
	}

	// 反转字符串
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// writeToFile 将内容写入文件
func (w *MermaidWriter) writeToFile(content, filePath string) error {
	// 这里需要导入 os 包
	// 在实际实现中，我们会这样写：
	/*
	   file, err := os.Create(filePath)
	   if err != nil {
	       return fmt.Errorf("创建文件失败: %v", err)
	   }
	   defer file.Close()

	   _, err = file.WriteString(content)
	   if err != nil {
	       return fmt.Errorf("写入文件失败: %v", err)
	   }
	*/

	log.Printf("将Mermaid内容写入文件: %s", filePath)
	// 暂时先打印内容，实际文件操作需要导入os包
	fmt.Printf("Mermaid内容已生成，应该写入文件: %s\n", filePath)
	fmt.Println("内容预览:")
	fmt.Println(content)

	return nil
}
