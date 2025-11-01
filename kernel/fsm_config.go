// kernel/fsm_config.go
package kernel

// createFSM 创建配置好的FSM
func (m *FSMManager) createFSM(session *UpdateSession) *FSM {
	fsm := NewFSM(StateIdle, session)

	// 添加所有状态转换
	m.addTransitions(fsm)

	return fsm
}

// addTransitions 添加所有状态转换
func (m *FSMManager) addTransitions(fsm *FSM) {
	// 从空闲状态开始
	fsm.AddTransition(StateIdle, EventStartUpdate, StateWaitingForNodeSelection, m.actionShowNodes)
	fsm.AddTransition(StateIdle, EventStartAdd, StateWaitingForContentInput, m.actionStartAddNode)
	fsm.AddTransition(StateIdle, EventStartDelete, StateWaitingForNodeSelection, m.actionShowNodesForDelete)

	// 节点选择
	fsm.AddTransition(StateWaitingForNodeSelection, EventNodeSelected, StateWaitingForUpdateType, m.actionNodeSelected)
	fsm.AddTransition(StateWaitingForNodeSelection, EventCancel, StateIdle, m.actionCancel)

	// 更新类型选择
	fsm.AddTransition(StateWaitingForUpdateType, EventUpdateTypeSelected, StateWaitingForContentInput, m.actionUpdateTypeSelected)

	// 修复：为选项2（修改回复）添加特殊转换
	fsm.AddTransition(StateWaitingForUpdateType, EventUpdateResponseSelected, StateWaitingForResponseSelect, m.actionShowResponseOptions)
	fsm.AddTransition(StateWaitingForUpdateType, EventCancel, StateIdle, m.actionCancel)

	// 内容输入
	fsm.AddTransition(StateWaitingForContentInput, EventContentProvided, StateIdle, m.actionContentProvided)
	fsm.AddTransition(StateWaitingForContentInput, EventCancel, StateIdle, m.actionCancel)

	// 回复选择
	fsm.AddTransition(StateWaitingForResponseSelect, EventResponseSelected, StateWaitingForContentInput, m.actionResponseSelected)
	fsm.AddTransition(StateWaitingForResponseSelect, EventCancel, StateIdle, m.actionCancel)

	// 回复删除
	fsm.AddTransition(StateWaitingForResponseDelete, EventResponseSelected, StateWaitingForConfirmation, m.actionResponseDeleteSelected)
	fsm.AddTransition(StateWaitingForResponseDelete, EventCancel, StateIdle, m.actionCancel)

	// 确认
	fsm.AddTransition(StateWaitingForConfirmation, EventConfirmationYes, StateIdle, m.actionConfirmationYes)
	fsm.AddTransition(StateWaitingForConfirmation, EventConfirmationNo, StateIdle, m.actionCancel)
}
