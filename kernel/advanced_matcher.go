// kernel/advanced_matcher.go
package kernel

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

// AdvancedPatternMatcher 高级模式匹配器
type AdvancedPatternMatcher struct {
	DefaultMatcher
}

// MatchPattern 增强的模式匹配
func (am *AdvancedPatternMatcher) MatchPattern(input, pattern string) bool {
	// 预处理输入和模式
	processedInput := am.preprocessInput(input)
	processedPattern := am.preprocessPattern(pattern)

	// 检查精确匹配
	if processedInput == processedPattern {
		return true
	}

	// 检查通配符匹配（包括星号和下划线）
	if strings.Contains(processedPattern, "*") || strings.Contains(processedPattern, "_") {
		return am.matchWithAdvancedWildcards(processedInput, processedPattern)
	}

	// 检查单词序列匹配
	return am.matchWordSequence(processedInput, processedPattern)
}

// preprocessInput 预处理输入
func (am *AdvancedPatternMatcher) preprocessInput(input string) string {
	// 转换为小写
	input = strings.ToLower(input)
	// 移除多余空格
	input = strings.TrimSpace(input)
	// 标准化空格
	input = am.normalizeSpaces(input)
	return input
}

// preprocessPattern 预处理模式
func (am *AdvancedPatternMatcher) preprocessPattern(pattern string) string {
	// 转换为小写
	pattern = strings.ToLower(pattern)
	// 移除多余空格
	pattern = strings.TrimSpace(pattern)
	// 标准化空格
	pattern = am.normalizeSpaces(pattern)
	return pattern
}

// normalizeSpaces 标准化空格
func (am *AdvancedPatternMatcher) normalizeSpaces(text string) string {
	// 将多个连续空格替换为单个空格
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(text, " ")
}

// matchWithAdvancedWildcards 使用高级通配符匹配
func (am *AdvancedPatternMatcher) matchWithAdvancedWildcards(input, pattern string) bool {
	// 构建正则表达式
	regexPattern := am.buildRegexFromPattern(pattern)

	matched, err := regexp.MatchString(regexPattern, input)
	if err != nil {
		return false
	}

	return matched
}

// buildRegexFromPattern 从模式构建正则表达式
func (am *AdvancedPatternMatcher) buildRegexFromPattern(pattern string) string {
	// 分割为单词
	words := strings.Fields(pattern)
	var regexParts []string

	for _, word := range words {
		if word == "*" {
			regexParts = append(regexParts, `\S+(?:\s+\S+)*`) // 匹配一个或多个单词
		} else if word == "_" {
			regexParts = append(regexParts, `\S+(?:\s+\S+)*`) // 下划线也匹配一个或多个单词
		} else {
			// 转义特殊字符并作为字面量
			escaped := regexp.QuoteMeta(word)
			regexParts = append(regexParts, escaped)
		}
	}

	// 构建完整的正则表达式
	regexStr := "^" + strings.Join(regexParts, "\\s+") + "$"
	return regexStr
}

// matchWordSequence 单词序列匹配
func (am *AdvancedPatternMatcher) matchWordSequence(input, pattern string) bool {
	inputWords := strings.Fields(input)
	patternWords := strings.Fields(pattern)

	// 如果模式单词数大于输入单词数，不可能匹配
	if len(patternWords) > len(inputWords) {
		return false
	}

	// 检查连续的单词匹配
	for i := 0; i <= len(inputWords)-len(patternWords); i++ {
		match := true
		for j := 0; j < len(patternWords); j++ {
			if inputWords[i+j] != patternWords[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}

// ExtractWildcards 修复大小写保留问题
func (am *AdvancedPatternMatcher) ExtractWildcards(input, pattern string) map[string]string {
	result := make(map[string]string)

	log.Printf("提取通配符 - 输入: %q, 模式: %q", input, pattern)

	// 使用大小写敏感的匹配检查
	if !am.MatchPatternCaseSensitive(input, pattern) {
		log.Printf("  输入不匹配模式")
		return result
	}

	// 构建提取正则表达式
	regexPattern := am.buildExtractionRegex(pattern)
	log.Printf("  构建的正则表达式: %q", regexPattern)

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		log.Printf("  编译正则表达式失败: %v", err)
		return result
	}

	// 执行匹配
	matches := re.FindStringSubmatch(input)
	log.Printf("  匹配结果: %v", matches)

	if len(matches) < 2 { // 第一个是完整匹配，后面是捕获组
		log.Printf("  没有找到捕获组")
		return result
	}

	// 提取捕获组
	starCount := strings.Count(pattern, "*")
	log.Printf("  模式中的星号数量: %d", starCount)

	for i := 1; i <= starCount && i < len(matches); i++ {
		key := fmt.Sprintf("star%d", i)
		result[key] = matches[i]
		log.Printf("  提取 %s: %q", key, matches[i])
	}

	// 设置第一个通配符为 "star"
	if starCount >= 1 {
		result["star"] = matches[1]
		log.Printf("  设置 star: %q", matches[1])
	}

	return result
}

// MatchPatternCaseSensitive 大小写敏感的模式匹配（用于提取）
func (am *AdvancedPatternMatcher) MatchPatternCaseSensitive(input, pattern string) bool {
	// 预处理但不转换为小写
	processedInput := am.preprocessInputCaseSensitive(input)
	processedPattern := am.preprocessPatternCaseSensitive(pattern)

	// 检查精确匹配
	if processedInput == processedPattern {
		return true
	}

	// 检查通配符匹配
	if strings.Contains(processedPattern, "*") {
		return am.matchWithAdvancedWildcardsCaseInSensitive(processedInput, processedPattern)
	}

	// 检查单词序列匹配
	return am.matchWordSequenceCaseSensitive(processedInput, processedPattern)
}

// preprocessInputCaseSensitive 预处理输入（保留大小写）
func (am *AdvancedPatternMatcher) preprocessInputCaseSensitive(input string) string {
	// 不移除大小写，只处理空格
	input = strings.TrimSpace(input)
	input = am.normalizeSpaces(input)
	return input
}

// preprocessPatternCaseSensitive 预处理模式（保留大小写）
func (am *AdvancedPatternMatcher) preprocessPatternCaseSensitive(pattern string) string {
	// 不移除大小写，只处理空格
	pattern = strings.TrimSpace(pattern)
	pattern = am.normalizeSpaces(pattern)
	return pattern
}

// matchWithAdvancedWildcardsCaseInSensitive 大小写敏感的通配符匹配
func (am *AdvancedPatternMatcher) matchWithAdvancedWildcardsCaseInSensitive(input, pattern string) bool {
	// 构建正则表达式（大小写敏感）
	regexPattern := am.buildRegexFromPatternCaseSensitive(pattern)
	regexPattern = "(?i)" + regexPattern
	matched, err := regexp.MatchString(regexPattern, input)
	if err != nil {
		return false
	}

	return matched
}

// buildRegexFromPatternCaseSensitive 构建大小写敏感的正则表达式
func (am *AdvancedPatternMatcher) buildRegexFromPatternCaseSensitive(pattern string) string {
	// 分割为单词
	words := strings.Fields(pattern)
	var regexParts []string

	for _, word := range words {
		if word == "*" {
			regexParts = append(regexParts, `\S+(?:\s+\S+)*`) //`\S+`
		} else if word == "_" {
			regexParts = append(regexParts, `\S+(?:\s+\S+)*`) // `\w+`
		} else {
			// 转义特殊字符并作为字面量（大小写敏感）
			escaped := regexp.QuoteMeta(word)
			regexParts = append(regexParts, escaped)
		}
	}

	// 构建完整的正则表达式（大小写敏感）
	regexStr := "^" + strings.Join(regexParts, "\\s+") + "$"
	return regexStr
}

// matchWordSequenceCaseSensitive 大小写敏感的单词序列匹配
func (am *AdvancedPatternMatcher) matchWordSequenceCaseSensitive(input, pattern string) bool {
	inputWords := strings.Fields(input)
	patternWords := strings.Fields(pattern)

	if len(patternWords) > len(inputWords) {
		return false
	}

	for i := 0; i <= len(inputWords)-len(patternWords); i++ {
		match := true
		for j := 0; j < len(patternWords); j++ {
			if inputWords[i+j] != patternWords[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}

// extractStarWildcardsCaseSensitive 大小写敏感的星号通配符提取
func (am *AdvancedPatternMatcher) extractStarWildcardsCaseSensitive(input, pattern string) map[string]string {
	result := make(map[string]string)

	// 构建正则表达式来提取通配符（大小写敏感）
	regexPattern := am.buildExtractionRegexCaseSensitive(pattern)
	re := regexp.MustCompile(regexPattern)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		// 统计通配符数量
		starCount := strings.Count(pattern, "*")
		matchIndex := 1

		for i := 0; i < starCount && matchIndex < len(matches); i++ {
			key := fmt.Sprintf("star%d", i+1)
			result[key] = matches[matchIndex]
			matchIndex++
		}

		// 设置第一个通配符为 "star"
		if starCount >= 1 {
			result["star"] = matches[1]
		}
	}

	return result
}

// buildExtractionRegexCaseSensitive 构建用于提取的大小写敏感正则表达式
func (am *AdvancedPatternMatcher) buildExtractionRegexCaseSensitive(pattern string) string {
	// 将模式中的每个星号替换为捕获组
	parts := strings.Split(pattern, "*")
	var regexParts []string

	for i, part := range parts {
		if part != "" {
			// 转义特殊字符（大小写敏感）
			escaped := regexp.QuoteMeta(part)
			regexParts = append(regexParts, escaped)
		}
		// 在每个星号位置添加捕获组（除了最后一个）
		if i < len(parts)-1 {
			regexParts = append(regexParts, `(\S+)`)
		}
	}

	return "^" + strings.Join(regexParts, "") + "$"
}

// extractStarWildcards 提取星号通配符内容
func (am *AdvancedPatternMatcher) extractStarWildcards(input, pattern string) map[string]string {
	result := make(map[string]string)

	// 构建正则表达式来提取通配符
	regexPattern := am.buildExtractionRegex(pattern)
	re := regexp.MustCompile(regexPattern)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		// 统计通配符数量
		starCount := strings.Count(pattern, "*")
		matchIndex := 1

		for i := 0; i < starCount && matchIndex < len(matches); i++ {
			key := fmt.Sprintf("star%d", i+1)
			result[key] = matches[matchIndex]
			matchIndex++
		}

		// 设置第一个通配符为 "star"
		if starCount >= 1 {
			result["star"] = matches[1]
		}
	}

	return result
}

// extractUnderscoreWildcards 提取下划线通配符内容
func (am *AdvancedPatternMatcher) extractUnderscoreWildcards(input, pattern string) map[string]string {
	result := make(map[string]string)

	// 构建正则表达式
	regexPattern := strings.ReplaceAll(pattern, "_", `(\S+)`)
	regexPattern = "^" + regexPattern + "$"

	re := regexp.MustCompile(regexPattern)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 1 {
		underscoreCount := strings.Count(pattern, "_")
		for i := 1; i <= underscoreCount && i < len(matches); i++ {
			key := fmt.Sprintf("underscore%d", i)
			result[key] = matches[i]
		}
	}

	return result
}

// buildExtractionRegex 构建用于提取的通配符正则表达式
func (am *AdvancedPatternMatcher) buildExtractionRegex(pattern string) string {
	// 转义模式中的所有特殊字符
	escapedPattern := regexp.QuoteMeta(pattern)

	// 将转义后的星号 `\*` 替换为捕获组
	// 注意：因为使用了 QuoteMeta，星号被转义为 `\*`，所以我们要替换 `\*`
	extractionRegex := strings.ReplaceAll(escapedPattern, `\*`, `(.+)`)
	regexPattern := "^" + extractionRegex + "$"
	egexPattern := "(?i)" + regexPattern
	// 添加字符串开始和结束锚点
	return egexPattern
}
