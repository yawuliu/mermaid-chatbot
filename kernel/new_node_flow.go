// new_node_flow.go - 新增节点流程
package kernel

import (
	"fmt"
	"mchat/model"
)

// 处理新增节点流程
func (m *CorpusUpdateManager) handleNewNodeFlow(session *UpdateSession, input string) string {
	if session.State == StateAddingNewNode {
		session.State = StateWaitingForNewNodePattern
		session.TargetNode = input
		return fmt.Sprintf("用户说的话设置为：\"%s\"\n现在请输入给机器人的问题：", input)
	}
	return "状态错误，请重新开始。"
}

// 处理新节点模式
func (m *CorpusUpdateManager) handleNewNodePattern(session *UpdateSession, input string) string {
	session.TargetNode = input
	session.State = StateWaitingForNewNodeResponse
	return "请输入机器人的第一个回复："
}

// 处理新节点回复
func (m *CorpusUpdateManager) handleNewNodeResponse(session *UpdateSession, input string) string {
	// 创建新Category
	newCategory := model.Category{
		Pattern:   session.TargetNode,
		Templates: []string{input},
	}

	m.kernel.AddCategory(newCategory)
	delete(m.sessions, session.UserID)

	return fmt.Sprintf("成功添加新对话！\n当用户说 \"%s\" 时，机器人会回复：\"%s\"\n\n您还可以继续添加更多回复变体。",
		session.TargetNode, input)
}
