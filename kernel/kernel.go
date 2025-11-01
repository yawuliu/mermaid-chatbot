package kernel

import (
	"fmt"
	"mchat/model"
	"strings"
	"sync"

	"mchat/config"
)

type Kernel struct {
	categories         []model.Category
	categoriesLock     sync.RWMutex
	config             *config.Config
	matcher            Matcher
	updateManager      *CorpusUpdateManager
	persistenceManager *PersistenceManager
}

func NewKernel() *Kernel {
	k := &Kernel{
		categories: make([]model.Category, 0),
		matcher:    NewDefaultMatcher(),
	}
	k.updateManager = NewCorpusUpdateManager(k)
	return k
}

// LoadCorpus 加载语料库
func (k *Kernel) LoadCorpus(corpusPath string) error {
	return k.loadMermaidCorpus(corpusPath)
}

// 修改主处理逻辑，在kernel.go中添加
func (k *Kernel) ProcessInput(userID, input string) string {
	// 首先检查是否是更新相关的请求
	if k.updateManager != nil {
		// 检查是否在更新会话中
		if k.updateManager.HasActiveSession(userID) {
			return k.updateManager.ProcessUpdateRequest(userID, input)
		}

		// 检查是否是更新命令
		if isUpdateCommand(input) {
			return k.updateManager.ProcessUpdateRequest(userID, input)
		}
	}

	// 正常对话处理
	return k.processNormalInput(input)
}

// 检查是否是更新命令
func isUpdateCommand(input string) bool {
	updateKeywords := []string{
		"更新", "修改", "编辑", "添加", "增加", "新建", "删除", "移除",
		"update", "modify", "edit", "add", "create", "new", "delete", "remove",
	}

	inputLower := strings.ToLower(input)
	for _, keyword := range updateKeywords {
		if strings.Contains(inputLower, keyword) {
			// 检查是否与语料、对话等相关
			if strings.Contains(inputLower, "语料") ||
				strings.Contains(inputLower, "对话") ||
				strings.Contains(inputLower, "回复") ||
				strings.Contains(inputLower, "corpus") ||
				strings.Contains(inputLower, "dialog") ||
				strings.Contains(inputLower, "response") {
				return true
			}
			// 单独的更新关键词也认为是更新命令
			if keyword == "更新" || keyword == "update" ||
				keyword == "添加" || keyword == "add" ||
				keyword == "删除" || keyword == "delete" {
				return true
			}
		}
	}
	return false
}

func (k *Kernel) processNormalInput(input string) string {
	k.categoriesLock.RLock()
	defer k.categoriesLock.RUnlock()

	processedInput := strings.TrimSpace(strings.ToLower(input))

	// 查找最佳匹配
	bestMatch := k.matcher.FindBestMatch(processedInput, k.categories)
	if bestMatch != nil {
		return bestMatch.GetRandomResponse()
	}

	if k.config != nil {
		return k.config.DefaultResponse
	}
	return "抱歉，我没明白您的意思。"
}

func (k *Kernel) AddCategory(category model.Category) {
	k.categoriesLock.Lock()
	defer k.categoriesLock.Unlock()
	k.categories = append(k.categories, category)
}

func (k *Kernel) GetCategoryCount() int {
	k.categoriesLock.RLock()
	defer k.categoriesLock.RUnlock()
	return len(k.categories)
}

func (k *Kernel) SetConfig(config *config.Config) {
	k.config = config
	k.matcher.SetThreshold(config.MatchThreshold)

	// 初始化持久化管理器
	if k.persistenceManager == nil {
		k.persistenceManager = NewPersistenceManager(k, config.CorpusPath)
	}
}

func (k *Kernel) ReloadCategories(categories []model.Category) {
	k.categoriesLock.Lock()
	defer k.categoriesLock.Unlock()
	k.categories = categories
}

func (k *Kernel) GetDebugInfo(userID string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("分类数量: %d\n", len(k.categories)))

	if k.updateManager != nil {
		builder.WriteString(k.updateManager.DebugSessions())

		// 检查当前用户是否有活跃会话
		hasSession := k.updateManager.HasActiveSession(userID)
		builder.WriteString(fmt.Sprintf("用户 %s 有活跃会话: %v\n", userID, hasSession))
	}

	return builder.String()
}
