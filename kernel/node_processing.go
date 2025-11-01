// node_processing.go - 节点处理逻辑
package kernel

import (
	"fmt"
	"log"
	"mchat/model"
	"strconv"
	"strings"
	"time"
)

// 处理节点选择
func (m *CorpusUpdateManager) handleNodeSelection(session *UpdateSession, input string) string {
	categories := m.kernel.GetAllCategories()

	// 调试信息
	log.Printf("处理节点选择，输入: %s, 可用分类数: %d", input, len(categories))

	index, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || index < 1 || index > len(categories) {
		return fmt.Sprintf("请输入有效的序号 (1-%d)，或输入\"取消\"退出。", len(categories))
	}

	selectedCategory := categories[index-1]
	session.TargetNode = selectedCategory.Pattern
	session.Context["selected_index"] = strconv.Itoa(index - 1)
	session.LastActive = time.Now() // 更新活动时间

	operation := session.Context["operation"]
	log.Printf("节点选择操作: %s, 目标节点: %s", operation, session.TargetNode)

	switch operation {
	case "update":
		session.State = StateWaitingForUpdateType
		return m.showUpdateOptions(session, selectedCategory)
	case "delete":
		session.State = StateWaitingForConfirmation
		return fmt.Sprintf("您确定要删除以下对话吗？\n\n用户说: \"%s\"\n机器人回复: %v\n\n请输入 \"确认\" 或 \"取消\"",
			selectedCategory.Pattern, selectedCategory.Templates)
	default:
		return "操作类型错误，请重新开始。"
	}
}

// 显示更新选项
func (m *CorpusUpdateManager) showUpdateOptions(session *UpdateSession, category model.Category) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("您选择了对话：\n用户说: \"%s\"\n", category.Pattern))
	builder.WriteString(fmt.Sprintf("当前回复: %v\n\n", category.Templates))
	builder.WriteString("请选择要修改的内容：\n")
	builder.WriteString("1. 修改用户说的话（模式）\n")
	builder.WriteString("2. 修改机器人的回复\n")
	builder.WriteString("3. 添加新的回复选项\n")
	builder.WriteString("4. 删除某个回复选项\n")

	session.State = StateWaitingForUpdateType
	return builder.String()
}

// 处理更新类型选择
func (m *CorpusUpdateManager) handleUpdateTypeSelection(session *UpdateSession, input string) string {
	switch input {
	case "1":
		session.State = StateWaitingForNewContent
		session.Context["update_type"] = "pattern"
		return fmt.Sprintf("当前用户说的话是：\"%s\"\n请输入新的内容：", session.TargetNode)
	case "2":
		session.State = StateWaitingForResponseSelection // 修正：使用新状态
		session.Context["update_type"] = "update_response"
		return m.showResponseOptionsForUpdate(session)
	case "3":
		session.State = StateWaitingForNewContent
		session.Context["update_type"] = "add_response"
		return "请输入要添加的新回复："
	case "4":
		session.State = StateWaitingForResponseDeletion // 修正：使用新状态
		session.Context["update_type"] = "delete_response"
		return m.showResponseOptionsForDeletion(session)
	default:
		return "请输入有效的选项 (1-4)，或输入\"取消\"退出。"
	}
}

// 显示回复选项供更新
func (m *CorpusUpdateManager) showResponseOptionsForUpdate(session *UpdateSession) string {
	categories := m.kernel.GetAllCategories()
	index, _ := strconv.Atoi(session.Context["selected_index"])
	category := categories[index]

	var builder strings.Builder
	builder.WriteString("请选择要修改的回复（输入序号）：\n")

	for i, response := range category.Templates {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, response))
	}

	// 保存回复列表到上下文，用于后续处理
	session.Context["response_list"] = strings.Join(category.Templates, "|||")

	//session.State = StateWaitingForNodeSelection
	//session.Context["update_type"] = "update_response"
	//session.Context["response_list"] = strings.Join(category.Templates, "|||")

	return builder.String()
}

// 显示回复选项供删除
func (m *CorpusUpdateManager) showResponseOptionsForDeletion(session *UpdateSession) string {
	categories := m.kernel.GetAllCategories()
	index, _ := strconv.Atoi(session.Context["selected_index"])
	category := categories[index]

	if len(category.Templates) <= 1 {
		return "该对话只有一个回复，不能删除。请选择其他操作。"
	}

	var builder strings.Builder
	builder.WriteString("请选择要删除的回复（输入序号）：\n")

	for i, response := range category.Templates {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, response))
	}

	session.State = StateWaitingForNodeSelection
	session.Context["update_type"] = "delete_response"
	session.Context["response_list"] = strings.Join(category.Templates, "|||")

	return builder.String()
}
