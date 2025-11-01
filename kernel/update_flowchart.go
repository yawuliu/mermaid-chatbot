// update_flowchart.go - 语料更新管理器
package kernel

import "time"

// 更新会话状态
type UpdateSession struct {
	UserID     string
	State      UpdateState
	LastActive time.Time // 记录最后活动时间
	TargetNode string
	NewContent string
	Context    map[string]string
}

type UpdateState int

const (
	StateIdle UpdateState = iota
	StateWaitingForNodeSelection
	StateWaitingForUpdateType
	StateWaitingForNewContent
	StateWaitingForConfirmation
	StateAddingNewNode
	StateWaitingForNewNodePattern
	StateWaitingForNewNodeResponse
	StateWaitingForResponseSelection // 新增：等待选择要修改的回复
	StateWaitingForResponseDeletion  // 新增：等待选择要删除的回复
)

// 更新管理器
type CorpusUpdateManager struct {
	sessions   map[string]*UpdateSession // userID -> session
	kernel     *Kernel
	sessionTTL time.Duration // 会话存活时间
}

func NewCorpusUpdateManager(kernel *Kernel) *CorpusUpdateManager {
	return &CorpusUpdateManager{
		sessions:   make(map[string]*UpdateSession),
		kernel:     kernel,
		sessionTTL: 30 * time.Minute, // 设置会话过期时间，例如30分钟
	}
}
