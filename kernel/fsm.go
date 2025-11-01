// kernel/fsm.go
package kernel

import (
	"fmt"
	"log"
	"time"
)

// State 状态定义
type State string

const (
	StateIdle                     State = "idle"
	StateWaitingForNodeSelection  State = "waiting_for_node_selection"
	StateWaitingForUpdateType     State = "waiting_for_update_type"
	StateWaitingForContentInput   State = "waiting_for_content_input"
	StateWaitingForConfirmation   State = "waiting_for_confirmation"
	StateWaitingForResponseSelect State = "waiting_for_response_select"
	StateWaitingForResponseDelete State = "waiting_for_response_delete"
)

// Event 事件定义
type Event string

const (
	EventStartUpdate            Event = "start_update"
	EventStartAdd               Event = "start_add"
	EventStartDelete            Event = "start_delete"
	EventNodeSelected           Event = "node_selected"
	EventUpdateTypeSelected     Event = "update_type_selected"
	EventUpdateResponseSelected Event = "update_response_selected" // 新增：选择修改回复
	EventContentProvided        Event = "content_provided"
	EventResponseSelected       Event = "response_selected"
	EventConfirmationYes        Event = "confirmation_yes"
	EventConfirmationNo         Event = "confirmation_no"
	EventCancel                 Event = "cancel"
)

// UpdateSession 更新会话
type UpdateSession struct {
	UserID     string
	State      State
	TargetNode string            // 目标节点模式
	NewContent string            // 新内容（模式或回复）
	Context    map[string]string // 会话上下文
	LastActive time.Time         // 最后活动时间
}

// Transition 状态转换
type Transition struct {
	FromState State
	Event     Event
	ToState   State
	Action    func(session *UpdateSession, input string) (string, error)
}

// FSM 有限状态机
type FSM struct {
	currentState State
	transitions  []Transition
	session      *UpdateSession
}

// NewFSM 创建新的状态机
func NewFSM(initialState State, session *UpdateSession) *FSM {
	return &FSM{
		currentState: initialState,
		session:      session,
		transitions:  make([]Transition, 0),
	}
}

// AddTransition 添加状态转换
func (f *FSM) AddTransition(from State, event Event, to State, action func(*UpdateSession, string) (string, error)) {
	f.transitions = append(f.transitions, Transition{
		FromState: from,
		Event:     event,
		ToState:   to,
		Action:    action,
	})
}

// HandleEvent 处理事件
func (f *FSM) HandleEvent(event Event, input string) (string, error) {
	for _, transition := range f.transitions {
		if transition.FromState == f.currentState && transition.Event == event {
			log.Printf("状态转换: %s -> %s (事件: %s)", f.currentState, transition.ToState, event)

			// 执行动作
			response, err := transition.Action(f.session, input)
			if err != nil {
				return "", err
			}

			// 更新状态
			f.currentState = transition.ToState
			f.session.State = transition.ToState

			return response, nil
		}
	}

	return "", fmt.Errorf("没有找到从状态 %s 处理事件 %s 的转换", f.currentState, event)
}

// GetCurrentState 获取当前状态
func (f *FSM) GetCurrentState() State {
	return f.currentState
}
