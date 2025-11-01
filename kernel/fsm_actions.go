// kernel/fsm_actions.go
package kernel

import (
	"fmt"
	"mchat/model"
	"strconv"
	"strings"
)

func (m *FSMManager) showResponseOptionsForUpdate(session *UpdateSession) string {
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

	return builder.String()
}

func (m *FSMManager) showResponseOptionsForDeletion(session *UpdateSession) string {
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

	session.Context["response_list"] = strings.Join(category.Templates, "|||")

	return builder.String()
}

// 动作实现
func (m *FSMManager) actionShowNodes(session *UpdateSession, input string) (string, error) {
	session.Context["operation"] = "update"
	return m.showAvailableNodes(session), nil
}

func (m *FSMManager) actionShowNodesForDelete(session *UpdateSession, input string) (string, error) {
	session.Context["operation"] = "delete"
	return m.showAvailableNodes(session), nil
}

func (m *FSMManager) actionStartAddNode(session *UpdateSession, input string) (string, error) {
	session.Context["update_type"] = "new_node_pattern"
	return "您想要添加新的对话节点。请先输入用户会说的话（模式）：", nil
}

func (m *FSMManager) actionNodeSelected(session *UpdateSession, input string) (string, error) {
	categories := m.kernel.GetAllCategories()

	index, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || index < 1 || index > len(categories) {
		return "", fmt.Errorf("无效的序号")
	}

	selectedCategory := categories[index-1]
	session.TargetNode = selectedCategory.Pattern
	session.Context["selected_index"] = strconv.Itoa(index - 1)

	operation := session.Context["operation"]
	if operation == "update" {
		return m.showUpdateOptions(session, selectedCategory), nil
	} else if operation == "delete" {
		session.Context["update_type"] = "delete_node"
		return fmt.Sprintf("您确定要删除以下对话吗？\n\n用户说: \"%s\"\n机器人回复: %v\n\n请输入 \"确认\" 或 \"取消\"",
			selectedCategory.Pattern, selectedCategory.Templates), nil
	}

	return "", fmt.Errorf("未知的操作类型")
}

func (m *FSMManager) actionUpdateTypeSelected(session *UpdateSession, input string) (string, error) {
	switch input {
	case "1":
		session.Context["update_type"] = "pattern"
		return fmt.Sprintf("当前用户说的话是：\"%s\"\n请输入新的内容：", session.TargetNode), nil
	case "3":
		session.Context["update_type"] = "add_response"
		return "请输入要添加的新回复：", nil
	default:
		return "", fmt.Errorf("无效的选项")
	}
}

// actionShowResponseOptions 显示回复选项供选择
func (m *FSMManager) actionShowResponseOptions(session *UpdateSession, input string) (string, error) {
	// 根据输入确定是修改回复还是删除回复
	if input == "2" {
		session.Context["update_type"] = "update_response"
	} else if input == "4" {
		session.Context["update_type"] = "delete_response"
	}

	return m.showResponseOptions(session), nil
}

// showResponseOptions 显示回复选项
func (m *FSMManager) showResponseOptions(session *UpdateSession) string {
	categories := m.kernel.GetAllCategories()
	index, _ := strconv.Atoi(session.Context["selected_index"])
	category := categories[index]

	var builder strings.Builder

	updateType := session.Context["update_type"]
	if updateType == "update_response" {
		builder.WriteString("请选择要修改的回复（输入序号）：\n")
	} else if updateType == "delete_response" {
		builder.WriteString("请选择要删除的回复（输入序号）：\n")
	}

	for i, response := range category.Templates {
		// 移除可能的引号
		cleanResponse := strings.Trim(response, "\"")
		builder.WriteString(fmt.Sprintf("%d. \"%s\"\n", i+1, cleanResponse))
	}

	// 保存回复列表到上下文
	session.Context["response_list"] = strings.Join(category.Templates, "|||")

	return builder.String()
}

// actionContentProvided 处理内容提供
func (m *FSMManager) actionContentProvided(session *UpdateSession, input string) (string, error) {
	updateType := session.Context["update_type"]

	switch updateType {
	case "pattern":
		// 更新模式
		session.NewContent = input
		categories := m.kernel.GetAllCategories()
		index, _ := strconv.Atoi(session.Context["selected_index"])
		oldCategory := categories[index]

		newCategory := model.NewCategoryWithSource(
			session.NewContent,
			oldCategory.Templates,
			oldCategory.SourceFile,
		)

		m.kernel.ReplaceCategory(index, newCategory)
		return fmt.Sprintf("成功更新！\n用户说的话已从 \"%s\" 修改为 \"%s\"",
			session.TargetNode, session.NewContent), nil

	case "add_response":
		// 添加回复
		categories := m.kernel.GetAllCategories()
		index, _ := strconv.Atoi(session.Context["selected_index"])
		category := categories[index]

		newTemplates := append(category.Templates, input)
		m.kernel.UpdateCategory(session.TargetNode, newTemplates)

		return fmt.Sprintf("成功添加新的回复！\n现在当用户说 \"%s\" 时，机器人会回复：%v",
			session.TargetNode, newTemplates), nil

	case "update_response":
		// 更新回复
		responseIndex, _ := strconv.Atoi(session.Context["response_index"])
		responseList := strings.Split(session.Context["response_list"], "|||")

		if responseIndex < 0 || responseIndex >= len(responseList) {
			return "", fmt.Errorf("回复序号无效")
		}

		oldResponse := strings.Trim(responseList[responseIndex], "\"")
		responseList[responseIndex] = input
		m.kernel.UpdateCategory(session.TargetNode, responseList)

		return fmt.Sprintf("成功更新回复！\n回复已从 \"%s\" 修改为 \"%s\"",
			oldResponse, input), nil

	case "new_node_pattern":
		// 新节点模式
		session.TargetNode = input
		session.Context["update_type"] = "new_node"
		return fmt.Sprintf("用户说的话设置为：\"%s\"\n现在请输入机器人的回复：", input), nil

	case "new_node":
		// 新节点回复
		newCategory := model.NewCategoryWithSource(
			session.TargetNode,
			[]string{input},
			"", // 来源文件将在持久化时推断
		)
		m.kernel.AddCategory(newCategory)
		return fmt.Sprintf("成功添加新对话！\n当用户说 \"%s\" 时，机器人会回复：\"%s\"",
			session.TargetNode, input), nil

	default:
		return "", fmt.Errorf("未知的更新类型")
	}
}

// actionResponseSelected 处理回复选择
func (m *FSMManager) actionResponseSelected(session *UpdateSession, input string) (string, error) {
	responseList := strings.Split(session.Context["response_list"], "|||")

	index, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || index < 1 || index > len(responseList) {
		return "", fmt.Errorf("无效的回复序号")
	}

	session.Context["response_index"] = strconv.Itoa(index - 1)
	oldResponse := strings.Trim(responseList[index-1], "\"")
	session.Context["old_response"] = oldResponse

	updateType := session.Context["update_type"]

	if updateType == "update_response" {
		return fmt.Sprintf("当前回复内容为：\"%s\"\n请输入新的回复内容：", oldResponse), nil
	} else if updateType == "delete_response" {
		if len(responseList) <= 1 {
			return "", fmt.Errorf("不能删除最后一个回复")
		}
		return fmt.Sprintf("您确定要删除以下回复吗？\n\n回复：\"%s\"\n\n请输入 \"确认\" 或 \"取消\"", oldResponse), nil
	}

	return "", fmt.Errorf("未知的更新类型")
}

func (m *FSMManager) actionResponseDeleteSelected(session *UpdateSession, input string) (string, error) {
	responseList := strings.Split(session.Context["response_list"], "|||")

	index, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || index < 1 || index > len(responseList) {
		return "", fmt.Errorf("无效的回复序号")
	}

	if len(responseList) <= 1 {
		return "", fmt.Errorf("不能删除最后一个回复")
	}

	session.Context["response_index"] = strconv.Itoa(index - 1)

	return fmt.Sprintf("您确定要删除以下回复吗？\n\n回复：\"%s\"\n\n请输入 \"确认\" 或 \"取消\"",
		responseList[index-1]), nil
}

func (m *FSMManager) actionConfirmationYes(session *UpdateSession, input string) (string, error) {
	updateType := session.Context["update_type"]

	switch updateType {
	case "pattern":
		// 更新模式
		categories := m.kernel.GetAllCategories()
		index, _ := strconv.Atoi(session.Context["selected_index"])
		oldCategory := categories[index]

		newCategory := model.NewCategory(session.NewContent, oldCategory.Templates)
		m.kernel.ReplaceCategory(index, newCategory)

		return fmt.Sprintf("成功更新！\n用户说的话已从 \"%s\" 修改为 \"%s\"",
			session.TargetNode, session.NewContent), nil

	case "delete_response":
		// 删除回复
		responseIndex, _ := strconv.Atoi(session.Context["response_index"])
		responseList := strings.Split(session.Context["response_list"], "|||")

		deletedResponse := responseList[responseIndex]
		newResponses := append(responseList[:responseIndex], responseList[responseIndex+1:]...)
		m.kernel.UpdateCategory(session.TargetNode, newResponses)

		return fmt.Sprintf("成功删除回复：\"%s\"\n现在当用户说 \"%s\" 时，机器人会回复：%v",
			deletedResponse, session.TargetNode, newResponses), nil

	case "delete_node":
		// 删除节点
		categories := m.kernel.GetAllCategories()
		index, _ := strconv.Atoi(session.Context["selected_index"])

		if index < 0 || index >= len(categories) {
			return "", fmt.Errorf("节点序号无效")
		}

		deletedPattern := categories[index].Pattern
		m.kernel.RemoveCategory(index)

		return fmt.Sprintf("成功删除对话：\"%s\"", deletedPattern), nil

	default:
		return "", fmt.Errorf("未知的确认操作")
	}
}

func (m *FSMManager) actionCancel(session *UpdateSession, input string) (string, error) {
	return "已取消更新操作。", nil
}
