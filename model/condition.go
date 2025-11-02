// model/condition.go
package model

// ConditionType 条件类型
type ConditionType string

const (
	ConditionPredicate ConditionType = "predicate" // 谓词条件
	ConditionVariable  ConditionType = "variable"  // 变量条件
	ConditionStar      ConditionType = "star"      // 通配符条件
)

// Condition 条件定义
type Condition struct {
	Type      ConditionType
	Name      string // 谓词名或变量名
	Value     string // 条件值
	IsDefault bool   // 是否为默认分支
}

// ConditionalResponse 条件回复
type ConditionalResponse struct {
	Condition Condition
	Response  string
}

// ConditionalCategory 条件分类
type ConditionalCategory struct {
	Pattern    string
	Conditions []ConditionalResponse
}
