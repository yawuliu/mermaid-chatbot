// kernel/condition_processor.go
package kernel

import (
	"mchat/model"
)

// ConditionProcessor 条件处理器
type ConditionProcessor struct {
	userStates map[string]*UserState // userID -> UserState
}

// UserState 用户状态
type UserState struct {
	Predicates map[string]string // 谓词值
	Variables  map[string]string // 局部变量
	Context    map[string]string // 对话上下文
}

// NewConditionProcessor 创建条件处理器
func NewConditionProcessor() *ConditionProcessor {
	return &ConditionProcessor{
		userStates: make(map[string]*UserState),
	}
}

// GetUserState 获取用户状态
func (cp *ConditionProcessor) GetUserState(userID string) *UserState {
	if _, exists := cp.userStates[userID]; !exists {
		cp.userStates[userID] = &UserState{
			Predicates: make(map[string]string),
			Variables:  make(map[string]string),
			Context:    make(map[string]string),
		}
	}
	return cp.userStates[userID]
}

// EvaluateCondition 评估条件
func (cp *ConditionProcessor) EvaluateCondition(userID string, condition model.Condition) bool {
	userState := cp.GetUserState(userID)

	switch condition.Type {
	case model.ConditionPredicate:
		// 检查谓词值
		value, exists := userState.Predicates[condition.Name]
		if !exists {
			return condition.IsDefault
		}
		return value == condition.Value

	case model.ConditionVariable:
		// 检查变量值
		value, exists := userState.Variables[condition.Name]
		if !exists {
			return condition.IsDefault
		}
		return value == condition.Value

	case model.ConditionStar:
		// 通配符条件 - 检查是否已设置
		_, exists := userState.Predicates[condition.Name]
		return exists

	default:
		return condition.IsDefault
	}
}

// GetConditionalResponse 获取条件回复
func (cp *ConditionProcessor) GetConditionalResponse(userID string, category model.ConditionalCategory) string {
	// 首先检查非默认条件
	for _, cr := range category.Conditions {
		if !cr.Condition.IsDefault && cp.EvaluateCondition(userID, cr.Condition) {
			return cr.Response
		}
	}

	// 如果没有匹配，检查默认条件
	for _, cr := range category.Conditions {
		if cr.Condition.IsDefault {
			return cr.Response
		}
	}

	return "" // 没有匹配的回复
}

// SetPredicate 设置谓词
func (cp *ConditionProcessor) SetPredicate(userID, name, value string) {
	userState := cp.GetUserState(userID)
	userState.Predicates[name] = value
}

// SetVariable 设置变量
func (cp *ConditionProcessor) SetVariable(userID, name, value string) {
	userState := cp.GetUserState(userID)
	userState.Variables[name] = value
}

// ClearVariables 清除变量（用于对话结束）
func (cp *ConditionProcessor) ClearVariables(userID string) {
	userState := cp.GetUserState(userID)
	userState.Variables = make(map[string]string)
}
