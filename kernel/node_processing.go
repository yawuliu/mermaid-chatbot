package kernel

import (
	"fmt"
	"mchat/model"
	"strings"
)

// showUpdateOptions 显示更新选项
func (m *FSMManager) showUpdateOptions(session *UpdateSession, category model.Category) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("您选择了对话：\n用户说: \"%s\"\n", category.Pattern))
	builder.WriteString(fmt.Sprintf("当前回复: %v\n\n", category.Templates))
	builder.WriteString("请选择要修改的内容：\n")
	builder.WriteString("1. 修改用户说的话（模式）\n")
	builder.WriteString("2. 修改机器人的回复\n")
	builder.WriteString("3. 添加新的回复选项\n")
	builder.WriteString("4. 删除某个回复选项\n")

	return builder.String()
}

// showAvailableNodes 显示可用节点
func (m *FSMManager) showAvailableNodes(session *UpdateSession) string {
	categories := m.kernel.GetAllCategories()

	var builder strings.Builder

	operation := session.Context["operation"]
	if operation == "update" {
		builder.WriteString("请选择要修改的对话节点（输入序号）：\n\n")
	} else if operation == "delete" {
		builder.WriteString("请选择要删除的对话节点（输入序号）：\n\n")
	}

	for i, category := range categories {
		builder.WriteString(fmt.Sprintf("%d. 用户说: \"%s\"\n", i+1, category.Pattern))
		builder.WriteString(fmt.Sprintf("   机器人回复: %v\n\n", category.Templates))
	}

	builder.WriteString("输入 \"取消\" 可以随时取消操作。")
	return builder.String()
}
