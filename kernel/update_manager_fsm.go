// kernel/update_manager_fsm.go
package kernel

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

// FSMManager 基于FSM的更新管理器
type FSMManager struct {
	userFSMs map[string]*FSM // userID -> FSM
	kernel   *Kernel
}

// NewFSMManager 创建新的FSM管理器
func NewFSMManager(kernel *Kernel) *FSMManager {
	return &FSMManager{
		userFSMs: make(map[string]*FSM),
		kernel:   kernel,
	}
}

// ProcessUpdateRequest 处理更新请求
func (m *FSMManager) ProcessUpdateRequest(userID, input string) string {
	// 获取或创建FSM
	fsm, exists := m.userFSMs[userID]
	if !exists {
		session := &UpdateSession{
			UserID:     userID,
			State:      StateIdle,
			Context:    make(map[string]string),
			LastActive: time.Now(),
		}
		fsm = m.createFSM(session)
		m.userFSMs[userID] = fsm
	}

	fsm.session.LastActive = time.Now()

	// 处理取消
	if strings.ToLower(input) == "取消" || strings.ToLower(input) == "cancel" {
		delete(m.userFSMs, userID)
		return "已取消更新操作。"
	}

	// 根据当前状态和输入确定事件
	event := m.determineEvent(fsm.GetCurrentState(), input)
	if event == "" {
		return "无效的操作，请输入有效选项或输入\"取消\"退出。"
	}

	// 处理事件
	response, err := fsm.HandleEvent(event, input)
	if err != nil {
		log.Printf("处理事件失败: %v", err)
		return "处理操作时发生错误，请重新开始。"
	}

	// 检查是否完成操作
	if fsm.GetCurrentState() == StateIdle {
		delete(m.userFSMs, userID)
	}

	return response
}

// determineEvent 修复事件检测逻辑
func (m *FSMManager) determineEvent(currentState State, input string) Event {
	switch currentState {
	case StateIdle:
		if m.isUpdateCommand(input) {
			return EventStartUpdate
		} else if m.isAddCommand(input) {
			return EventStartAdd
		} else if m.isDeleteCommand(input) {
			return EventStartDelete
		}

	case StateWaitingForNodeSelection:
		if _, err := strconv.Atoi(strings.TrimSpace(input)); err == nil {
			return EventNodeSelected
		}

	case StateWaitingForUpdateType:
		// 根据不同的选项返回不同的事件
		switch input {
		case "1":
			return EventUpdateTypeSelected // 修改模式
		case "2":
			return EventUpdateResponseSelected // 修改回复
		case "3":
			return EventUpdateTypeSelected // 添加回复
		case "4":
			return EventUpdateResponseSelected // 删除回复
		}

	case StateWaitingForContentInput:
		return EventContentProvided

	case StateWaitingForConfirmation:
		if strings.ToLower(input) == "确认" {
			return EventConfirmationYes
		} else {
			return EventConfirmationNo
		}

	case StateWaitingForResponseSelect, StateWaitingForResponseDelete:
		if _, err := strconv.Atoi(strings.TrimSpace(input)); err == nil {
			return EventResponseSelected
		}
	}

	return ""
}

// 命令检测辅助函数
func (m *FSMManager) isUpdateCommand(input string) bool {
	updateKeywords := []string{"更新", "修改", "编辑", "update", "modify", "edit"}
	return m.containsKeywords(input, updateKeywords)
}

func (m *FSMManager) isAddCommand(input string) bool {
	addKeywords := []string{"添加", "增加", "新建", "add", "create", "new"}
	return m.containsKeywords(input, addKeywords)
}

func (m *FSMManager) isDeleteCommand(input string) bool {
	deleteKeywords := []string{"删除", "移除", "delete", "remove"}
	return m.containsKeywords(input, deleteKeywords)
}

func (m *FSMManager) containsKeywords(input string, keywords []string) bool {
	inputLower := strings.ToLower(input)
	for _, keyword := range keywords {
		if strings.Contains(inputLower, keyword) {
			return true
		}
	}
	return false
}

// HasActiveSession 检查是否有活跃会话
func (m *FSMManager) HasActiveSession(userID string) bool {
	fsm, exists := m.userFSMs[userID]
	if !exists {
		return false
	}
	return fsm.GetCurrentState() != StateIdle
}

// 添加调试信息
func (m *FSMManager) DebugSessions() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("当前活跃会话数: %d\n", len(m.userFSMs)))
	for userID, fsm := range m.userFSMs {
		builder.WriteString(fmt.Sprintf("用户: %s, 状态: %v, 最后活动: %v\n",
			userID, fsm.session.State, fsm.session.LastActive))
	}
	return builder.String()
}
