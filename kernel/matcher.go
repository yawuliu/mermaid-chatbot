package kernel

import (
	"mchat/model"
	"strings"
)

type Matcher interface {
	FindBestMatch(input string, categories []model.Category) *model.Category
	SetThreshold(threshold int)
}

type DefaultMatcher struct {
	threshold int
}

func NewDefaultMatcher() *DefaultMatcher {
	return &DefaultMatcher{
		threshold: 70,
	}
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
