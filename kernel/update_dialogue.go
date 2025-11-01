// update_dialogue.go - 对话状态处理
package kernel

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// HasActiveSession 检查用户是否有活跃的会话
func (m *CorpusUpdateManager) HasActiveSession(userID string) bool {
	session, exists := m.sessions[userID]
	if !exists {
		return false
	}

	// 检查会话是否过期
	if time.Since(session.LastActive) > m.sessionTTL {
		delete(m.sessions, userID) // 清理过期会话
		return false
	}

	return true
}

// 添加调试信息
func (m *CorpusUpdateManager) DebugSessions() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("当前活跃会话数: %d\n", len(m.sessions)))
	for userID, session := range m.sessions {
		builder.WriteString(fmt.Sprintf("用户: %s, 状态: %d, 最后活动: %v\n",
			userID, session.State, session.LastActive))
	}
	return builder.String()
}

// 处理更新相关的用户输入
func (m *CorpusUpdateManager) ProcessUpdateRequest(userID, input string) string {
	session, exists := m.sessions[userID]
	if !exists {
		session = &UpdateSession{
			UserID:     userID,
			State:      StateIdle,
			Context:    make(map[string]string),
			LastActive: time.Now(),
		}
		m.sessions[userID] = session
	}

	// 更新最后活动时间
	session.LastActive = time.Now()

	oldState := session.State
	log.Printf("更新请求 - 用户: %s, 输入: %s, 旧状态: %d", userID, input, oldState)

	// 检查是否要取消操作
	if strings.ToLower(input) == "取消" || strings.ToLower(input) == "cancel" {
		delete(m.sessions, userID)
		return "已取消更新操作。"
	}

	// 根据当前状态处理输入
	var response string
	// 根据当前状态处理输入
	switch session.State {
	case StateIdle:
		response = m.handleIdleState(session, input)
	case StateWaitingForNodeSelection:
		response = m.handleNodeSelection(session, input)
	case StateWaitingForUpdateType:
		response = m.handleUpdateTypeSelection(session, input)
	case StateWaitingForNewContent:
		response = m.handleNewContent(session, input)
	case StateWaitingForConfirmation:
		response = m.handleConfirmation(session, input)
	case StateAddingNewNode:
		response = m.handleNewNodeFlow(session, input)
	case StateWaitingForNewNodePattern:
		response = m.handleNewNodePattern(session, input)
	case StateWaitingForNewNodeResponse:
		response = m.handleNewNodeResponse(session, input)
	case StateWaitingForResponseSelection: // 新增：处理回复选择
		response = m.handleResponseSelection(session, input)
	case StateWaitingForResponseDeletion: // 新增：处理回复删除选择
		response = m.handleResponseDeletion(session, input)
	default:
		response = "系统状态错误，请重新开始。"
	}

	log.Printf("状态转换 - 用户: %s, 从状态 %d 到 %d", userID, oldState, session.State)
	return response
}

// 空闲状态 - 检测更新意图
func (m *CorpusUpdateManager) handleIdleState(session *UpdateSession, input string) string {
	updateKeywords := []string{"更新", "修改", "编辑", "update", "modify", "edit"}
	addKeywords := []string{"添加", "增加", "新建", "add", "create", "new"}
	deleteKeywords := []string{"删除", "移除", "delete", "remove"}

	inputLower := strings.ToLower(input)

	// 检查是否要添加新节点
	for _, keyword := range addKeywords {
		if strings.Contains(inputLower, keyword) {
			session.State = StateAddingNewNode
			return "您想要添加新的对话节点。请先输入用户会说的话（模式）："
		}
	}

	// 检查是否要更新现有节点
	for _, keyword := range updateKeywords {
		if strings.Contains(inputLower, keyword) {
			return m.showAvailableNodes(session)
		}
	}

	// 检查是否要删除节点
	for _, keyword := range deleteKeywords {
		if strings.Contains(inputLower, keyword) {
			session.State = StateWaitingForNodeSelection
			session.Context["operation"] = "delete"
			return m.showAvailableNodesForDeletion(session)
		}
	}

	return "如果您想要修改语料，可以说：\n• \"更新语料\" - 修改现有对话\n• \"添加对话\" - 创建新的对话\n• \"删除对话\" - 移除现有对话"
}

// 显示可用的节点供选择
func (m *CorpusUpdateManager) showAvailableNodes(session *UpdateSession) string {
	categories := m.kernel.GetAllCategories()
	if len(categories) == 0 {
		return "当前没有可用的对话节点。"
	}

	session.State = StateWaitingForNodeSelection
	session.Context["operation"] = "update"

	var builder strings.Builder
	builder.WriteString("请选择要修改的对话节点（输入序号）：\n\n")

	for i, category := range categories {
		builder.WriteString(fmt.Sprintf("%d. 用户说: \"%s\"\n", i+1, category.Pattern))
		builder.WriteString(fmt.Sprintf("   机器人回复: %v\n\n", category.Templates))
	}

	builder.WriteString("输入 \"取消\" 可以随时取消操作。")
	return builder.String()
}

// 显示可删除的节点
func (m *CorpusUpdateManager) showAvailableNodesForDeletion(session *UpdateSession) string {
	categories := m.kernel.GetAllCategories()
	if len(categories) == 0 {
		return "当前没有可删除的对话节点。"
	}

	var builder strings.Builder
	builder.WriteString("请选择要删除的对话节点（输入序号）：\n\n")

	for i, category := range categories {
		builder.WriteString(fmt.Sprintf("%d. 用户说: \"%s\"\n", i+1, category.Pattern))
		builder.WriteString(fmt.Sprintf("   机器人回复: %v\n\n", category.Templates))
	}

	builder.WriteString("输入 \"取消\" 可以随时取消操作。")
	return builder.String()
}

func (m *CorpusUpdateManager) handleResponseSelection(session *UpdateSession, input string) string {
	// 解析回复列表
	responseList := strings.Split(session.Context["response_list"], "|||")

	index, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || index < 1 || index > len(responseList) {
		return fmt.Sprintf("请输入有效的回复序号 (1-%d)，或输入\"取消\"退出。", len(responseList))
	}

	// 保存选择的回复索引
	session.Context["response_index"] = strconv.Itoa(index - 1)
	session.Context["old_response"] = responseList[index-1]

	// 进入等待新内容状态
	session.State = StateWaitingForNewContent
	return fmt.Sprintf("当前回复内容为：\"%s\"\n请输入新的回复内容：", responseList[index-1])
}

func (m *CorpusUpdateManager) handleResponseDeletion(session *UpdateSession, input string) string {
	// 解析回复列表
	responseList := strings.Split(session.Context["response_list"], "|||")

	index, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || index < 1 || index > len(responseList) {
		return fmt.Sprintf("请输入有效的回复序号 (1-%d)，或输入\"取消\"退出。", len(responseList))
	}

	// 检查是否要删除最后一个回复
	if len(responseList) <= 1 {
		return "不能删除最后一个回复。请选择其他操作。"
	}

	// 保存选择的回复索引
	session.Context["response_index"] = strconv.Itoa(index - 1)

	// 进入确认状态
	session.State = StateWaitingForConfirmation
	return fmt.Sprintf("您确定要删除以下回复吗？\n\n回复：\"%s\"\n\n请输入 \"确认\" 或 \"取消\"",
		responseList[index-1])
}
