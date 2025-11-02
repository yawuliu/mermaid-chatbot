package kernel

import (
	"fmt"
	"mchat/model"
	"regexp"
	"strings"
	"sync"

	"mchat/config"
)

type Kernel struct {
	categories            []model.Category
	conditionalCategories []model.ConditionalCategory
	categoriesLock        sync.RWMutex
	conditionProcessor    *ConditionProcessor
	config                *config.Config
	matcher               PatternMatcher // 使用模式匹配器接口
	updateManager         *FSMManager    // 使用FSM管理器
	persistenceManager    *PersistenceManager
}

func NewKernel() *Kernel {
	k := &Kernel{
		categories:            make([]model.Category, 0),
		conditionalCategories: make([]model.ConditionalCategory, 0),
		matcher:               &AdvancedPatternMatcher{}, // 使用高级模式匹配器
		conditionProcessor:    NewConditionProcessor(),
	}
	k.updateManager = NewFSMManager(k) // 使用FSM管理器
	return k
}

// LoadCorpus 加载语料库
func (k *Kernel) LoadCorpus(corpusPath string) error {
	return k.loadConditionalCorpus(corpusPath)
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

	// 正常对话处理 - 先检查条件分类
	response := k.processConditionalInput(userID, input)
	if response != "" {
		return response
	}

	// 正常对话处理
	return k.processNormalInput(input)
}

// processConditionalInput 处理条件输入
func (k *Kernel) processConditionalInput(userID, input string) string {
	k.categoriesLock.RLock()
	defer k.categoriesLock.RUnlock()

	processedInput := strings.TrimSpace(strings.ToLower(input))

	// 在条件分类中查找匹配
	for _, category := range k.conditionalCategories {
		if k.matcher.MatchPattern(processedInput, category.Pattern) {
			// 提取通配符并设置到用户状态中
			wildcards := k.matcher.ExtractWildcards(processedInput, category.Pattern)
			k.setWildcardsToUserState(userID, wildcards)

			response := k.conditionProcessor.GetConditionalResponse(userID, category)
			if response != "" {
				return k.processTemplate(response, userID)
			}
		}
	}

	return ""
}

// setWildcardsToUserState 设置通配符到用户状态
func (k *Kernel) setWildcardsToUserState(userID string, wildcards map[string]string) {
	for key, value := range wildcards {
		k.conditionProcessor.SetVariable(userID, key, value)
	}
}

// processTemplate 处理模板中的特殊标签
func (k *Kernel) processTemplate(template, userID string) string {
	// 处理 <get> 标签
	template = k.processGetTags(template, userID)

	// 处理 <star> 标签
	template = k.processStarTags(template, userID)

	// 这里可以添加其他模板标签的处理

	return template
}

// processGetTags 处理 <get> 标签
func (k *Kernel) processGetTags(template, userID string) string {
	re := regexp.MustCompile(`<get\s+name="([^"]+)"\s*/>`)

	return re.ReplaceAllStringFunc(template, func(match string) string {
		matches := re.FindStringSubmatch(match)
		if len(matches) > 1 {
			predicateName := matches[1]
			userState := k.conditionProcessor.GetUserState(userID)
			if value, exists := userState.Predicates[predicateName]; exists {
				return value
			}
		}
		return "" // 如果谓词不存在，返回空字符串
	})
}

// processStarTags 处理 <star> 标签
func (k *Kernel) processStarTags(template, userID string) string {
	re := regexp.MustCompile(`<star\s*/>`)

	return re.ReplaceAllStringFunc(template, func(match string) string {
		userState := k.conditionProcessor.GetUserState(userID)
		if value, exists := userState.Variables["star"]; exists {
			return value
		}
		return "" // 如果star不存在，返回空字符串
	})
}

// AddConditionalCategory 添加条件分类
func (k *Kernel) AddConditionalCategory(category model.ConditionalCategory) {
	k.categoriesLock.Lock()
	defer k.categoriesLock.Unlock()
	k.conditionalCategories = append(k.conditionalCategories, category)
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
