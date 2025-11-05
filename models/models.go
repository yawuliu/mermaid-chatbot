package models

import "regexp"

// 节点类型
type NodeType string

const (
	NodeQuestion  NodeType = "question"
	NodeAnswer    NodeType = "answer"
	NodeCondition NodeType = "condition"
	NodeVariable  NodeType = "variable"
	NodeTopic     NodeType = "topic"
)

// 匹配类型
type MatchType string

const (
	MatchExact    MatchType = "exact"
	MatchContains MatchType = "contains"
	MatchPattern  MatchType = "pattern"
	MatchRegex    MatchType = "regex"
)

// DSL 节点
type Node struct {
	ID         string   `json:"id"`
	Type       NodeType `json:"type"`
	Content    string   `json:"content"`
	Metadata   Metadata `json:"metadata"`
	PatternID  string   `json:"pattern_id,omitempty"`
	ResponseID string   `json:"response_id,omitempty"`
}

// 边
type Edge struct {
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Label    string `json:"label,omitempty"`
}

// 图
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// 元数据
type Metadata struct {
	Pattern    string                 `json:"pattern,omitempty"`
	Response   string                 `json:"response,omitempty"`
	Condition  string                 `json:"condition,omitempty"`
	Variable   string                 `json:"variable,omitempty"`
	Value      string                 `json:"value,omitempty"`
	Topic      string                 `json:"topic,omitempty"`
	Priority   int                    `json:"priority,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// 模式定义
type Pattern struct {
	ID       string         `json:"id"`
	Type     MatchType      `json:"type"`
	Content  []string       `json:"content"`
	Priority int            `json:"priority"`
	Regex    *regexp.Regexp `json:"-"`
}

// 响应模板
type Template struct {
	ID         string      `json:"id"`
	Content    string      `json:"content"`
	Actions    []Action    `json:"actions,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// 动作
type Action struct {
	Type  string `json:"type"` // set, get, topic, random
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// 条件
type Condition struct {
	Variable string `json:"variable"`
	Operator string `json:"operator"` // eq, ne, gt, lt, exists
	Value    string `json:"value,omitempty"`
}

// 对话上下文
type Context struct {
	UserID       string
	SessionID    string
	Variables    map[string]string
	CurrentTopic string
	History      []Message
}

// 消息
type Message struct {
	Role    string `json:"role"` // user, assistant
	Content string `json:"content"`
	Time    int64  `json:"time"`
}
