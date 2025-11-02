package kernel

import (
	"fmt"
	"mchat/model"
	"regexp"
	"strings"
)

type Matcher interface {
	FindBestMatch(input string, categories []model.Category) *model.Category
	SetThreshold(threshold int)
}

// PatternMatcher 模式匹配器接口
type PatternMatcher interface {
	Matcher
	MatchPattern(input, pattern string) bool
	ExtractWildcards(input, pattern string) map[string]string
}

type DefaultMatcher struct {
	threshold int
}

func NewDefaultMatcher() *DefaultMatcher {
	return &DefaultMatcher{
		threshold: 70,
	}
}

// MatchPattern 检查输入是否匹配给定的模式
func (m *DefaultMatcher) MatchPattern(input, pattern string) bool {
	// 如果模式是精确匹配
	if input == pattern {
		return true
	}

	// 预处理输入和模式（转换为小写用于匹配）
	processedInput := strings.ToLower(strings.TrimSpace(input))
	processedPattern := strings.ToLower(strings.TrimSpace(pattern))

	// 如果模式包含通配符
	if strings.Contains(processedPattern, "*") {
		return m.matchWithWildcards(processedInput, processedPattern)
	}

	// 如果模式包含下划线通配符
	if strings.Contains(processedPattern, "_") {
		return m.matchWithUnderscore(processedInput, processedPattern)
	}

	// 简单包含匹配
	return strings.Contains(processedInput, processedPattern)
}

// matchWithWildcards 使用星号通配符匹配
func (m *DefaultMatcher) matchWithWildcards(input, pattern string) bool {
	// 将模式转换为正则表达式
	regexPattern := m.patternToRegex(pattern)
	matched, _ := regexp.MatchString(regexPattern, input)
	return matched
}

// matchWithUnderscore 使用下划线通配符匹配
func (m *DefaultMatcher) matchWithUnderscore(input, pattern string) bool {
	// 下划线应该匹配一个或多个单词
	// 将下划线替换为匹配一个或多个非空白字符序列（单词）的正则表达式
	regexPattern := strings.ReplaceAll(pattern, "_", `\S+(?:\s+\S+)*`)
	// 确保匹配整个字符串
	regexPattern = "^" + regexPattern + "$"

	matched, err := regexp.MatchString(regexPattern, input)
	if err != nil {
		return false
	}
	return matched
}

// patternToRegex 将AIML模式转换为正则表达式
func (m *DefaultMatcher) patternToRegex(pattern string) string {
	// 转义正则特殊字符
	pattern = regexp.QuoteMeta(pattern)

	// 将AIML通配符转换为正则表达式
	pattern = strings.ReplaceAll(pattern, `\*`, `.*`)
	pattern = strings.ReplaceAll(pattern, `\_`, `\S+`)

	// 确保匹配整个字符串
	return "^" + pattern + "$"
}

// ExtractWildcards 从输入中提取通配符内容
func (m *DefaultMatcher) ExtractWildcards(input, pattern string) map[string]string {
	result := make(map[string]string)

	// 使用原始输入进行匹配检查
	if !m.MatchPattern(input, pattern) {
		return result
	}

	// 构建提取正则表达式（大小写敏感）
	regexPattern := m.buildExtractionRegexCaseSensitive(pattern)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return result
	}

	// 执行匹配
	matches := re.FindStringSubmatch(input)
	if len(matches) < 2 {
		return result
	}

	// 提取捕获组
	starCount := strings.Count(pattern, "*")
	for i := 1; i <= starCount && i < len(matches); i++ {
		key := fmt.Sprintf("star%d", i)
		result[key] = matches[i]
	}

	// 设置第一个通配符为 "star"
	if starCount >= 1 {
		result["star"] = matches[1]
	}

	return result
}

// buildExtractionRegexCaseSensitive 构建大小写敏感的提取正则表达式
func (m *DefaultMatcher) buildExtractionRegexCaseSensitive(pattern string) string {
	// 转义模式中的所有特殊字符
	escapedPattern := regexp.QuoteMeta(pattern)

	// 将转义后的星号 `\*` 替换为捕获组
	extractionRegex := strings.ReplaceAll(escapedPattern, `\*`, `(.+)`)

	// 添加字符串开始和结束锚点
	return "^" + extractionRegex + "$"
}

// patternToRegexCaseSensitive 大小写敏感的模式转换
func (m *DefaultMatcher) patternToRegexCaseSensitive(pattern string) string {
	// 转义正则特殊字符
	pattern = regexp.QuoteMeta(pattern)

	// 将AIML通配符转换为正则表达式
	pattern = strings.ReplaceAll(pattern, `\*`, `(\S+)`)
	pattern = strings.ReplaceAll(pattern, `\_`, `(\S+)`)

	// 确保匹配整个字符串
	return "^" + pattern + "$"
}

func (m *DefaultMatcher) FindBestMatch(input string, categories []model.Category) *model.Category {
	var bestMatch *model.Category
	bestScore := 0

	for i := range categories {
		score := m.calculateSimilarity(input, categories[i].Pattern)
		if score > bestScore && score >= m.threshold {
			bestScore = score
			bestMatch = &categories[i]
		}
	}

	return bestMatch
}

func (m *DefaultMatcher) calculateSimilarity(input, pattern string) int {
	if input == pattern {
		return 100
	}

	if strings.Contains(input, pattern) || strings.Contains(pattern, input) {
		return 80
	}

	// 简单的编辑距离相似度计算
	inputRunes := []rune(input)
	patternRunes := []rune(pattern)

	maxLen := len(inputRunes)
	if len(patternRunes) > maxLen {
		maxLen = len(patternRunes)
	}

	if maxLen == 0 {
		return 0
	}

	distance := m.levenshteinDistance(inputRunes, patternRunes)
	similarity := (1 - float64(distance)/float64(maxLen)) * 100

	return int(similarity)
}

func (m *DefaultMatcher) levenshteinDistance(a, b []rune) int {
	alen, blen := len(a), len(b)
	if alen == 0 {
		return blen
	}
	if blen == 0 {
		return alen
	}

	matrix := make([][]int, alen+1)
	for i := range matrix {
		matrix[i] = make([]int, blen+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= alen; i++ {
		for j := 1; j <= blen; j++ {
			if a[i-1] == b[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = min(
					matrix[i-1][j]+1,   // deletion
					matrix[i][j-1]+1,   // insertion
					matrix[i-1][j-1]+1, // substitution
				)
			}
		}
	}

	return matrix[alen][blen]
}

func min(values ...int) int {
	minVal := values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
	}
	return minVal
}

func (m *DefaultMatcher) SetThreshold(threshold int) {
	m.threshold = threshold
}
