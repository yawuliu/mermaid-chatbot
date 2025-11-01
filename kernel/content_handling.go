// content_handling.go - 内容处理
package kernel

import (
	"fmt"
	"mchat/model"
	"strconv"
	"strings"
)

// 处理新内容输入
func (m *CorpusUpdateManager) handleNewContent(session *UpdateSession, input string) string {
	updateType := session.Context["update_type"]

	switch updateType {
	case "pattern":
		session.NewContent = input
		session.State = StateWaitingForConfirmation
		return fmt.Sprintf("确认将用户说的话从 \"%s\" 修改为 \"%s\"？\n请输入 \"确认\" 或 \"取消\"",
			session.TargetNode, input)

	case "add_response":
		// 直接添加到现有回复中
		categories := m.kernel.GetAllCategories()
		index, _ := strconv.Atoi(session.Context["selected_index"])
		category := categories[index]

		newTemplates := append(category.Templates, input)
		m.kernel.UpdateCategory(session.TargetNode, newTemplates)

		delete(m.sessions, session.UserID)
		return fmt.Sprintf("成功添加新的回复！\n现在当用户说 \"%s\" 时，机器人会回复：%v",
			session.TargetNode, newTemplates)

	case "update_response":
		// 更新特定回复
		responseIndex, _ := strconv.Atoi(session.Context["response_index"])
		responseList := strings.Split(session.Context["response_list"], "|||")

		if responseIndex < 0 || responseIndex >= len(responseList) {
			return "回复序号无效，请重新开始。"
		}

		oldResponse := responseList[responseIndex]
		responseList[responseIndex] = input

		m.kernel.UpdateCategory(session.TargetNode, responseList)

		delete(m.sessions, session.UserID)
		return fmt.Sprintf("成功更新回复！\n回复已从 \"%s\" 修改为 \"%s\"",
			oldResponse, input)

	default:
		return "未知的更新类型，请重新开始。"
	}
}

// 处理确认
func (m *CorpusUpdateManager) handleConfirmation(session *UpdateSession, input string) string {
	if strings.ToLower(input) != "确认" {
		delete(m.sessions, session.UserID)
		return "已取消更新操作。"
	}

	operation := session.Context["operation"]
	switch operation {
	case "update":
		return m.executeUpdate(session)
	case "delete":
		return m.executeDeletion(session)
	default:
		return "未知的操作类型。"
	}
}

// 执行更新操作
func (m *CorpusUpdateManager) executeUpdate(session *UpdateSession) string {
	updateType := session.Context["update_type"]

	switch updateType {
	case "pattern":
		// 更新模式（用户说的话）
		categories := m.kernel.GetAllCategories()
		index, _ := strconv.Atoi(session.Context["selected_index"])
		oldCategory := categories[index]

		// 创建新的Category并替换
		newCategory := model.Category{
			Pattern:   session.NewContent,
			Templates: oldCategory.Templates,
		}

		m.kernel.ReplaceCategory(index, newCategory)
		delete(m.sessions, session.UserID)
		return fmt.Sprintf("成功更新！\n用户说的话已从 \"%s\" 修改为 \"%s\"",
			session.TargetNode, session.NewContent)

	case "update_response":
		// 更新特定回复
		responseIndex, _ := strconv.Atoi(session.Context["response_index"])
		responseList := strings.Split(session.Context["response_list"], "|||")

		if responseIndex < 1 || responseIndex > len(responseList) {
			return "回复序号无效，请重新开始。"
		}

		responseList[responseIndex-1] = session.NewContent
		m.kernel.UpdateCategory(session.TargetNode, responseList)

		delete(m.sessions, session.UserID)
		return fmt.Sprintf("成功更新回复！\n当用户说 \"%s\" 时，机器人现在会回复：%v",
			session.TargetNode, responseList)

	case "delete_response":
		// 删除特定回复
		responseIndex, _ := strconv.Atoi(session.Context["response_index"])
		responseList := strings.Split(session.Context["response_list"], "|||")

		if len(responseList) <= 1 {
			return "不能删除最后一个回复，请选择其他操作。"
		}

		newResponses := append(responseList[:responseIndex-1], responseList[responseIndex:]...)
		m.kernel.UpdateCategory(session.TargetNode, newResponses)

		delete(m.sessions, session.UserID)
		return fmt.Sprintf("成功删除回复！\n当用户说 \"%s\" 时，机器人现在会回复：%v",
			session.TargetNode, newResponses)

	default:
		return "未知的更新类型。"
	}
}

// 执行删除操作
func (m *CorpusUpdateManager) executeDeletion(session *UpdateSession) string {
	categories := m.kernel.GetAllCategories()
	index, _ := strconv.Atoi(session.Context["selected_index"])

	if index < 0 || index >= len(categories) {
		return "节点序号无效，请重新开始。"
	}

	deletedPattern := categories[index].Pattern
	m.kernel.RemoveCategory(index)

	delete(m.sessions, session.UserID)
	return fmt.Sprintf("成功删除对话：\"%s\"", deletedPattern)
}
