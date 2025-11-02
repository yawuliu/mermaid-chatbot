// kernel/matcher_test.go
package kernel

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	matcher := &AdvancedPatternMatcher{}

	testCases := []struct {
		input    string
		pattern  string
		expected bool
		name     string
	}{
		{"hello world", "hello world", true, "精确匹配"},
		{"hello world", "hello *", true, "星号通配符结尾"},
		{"hello world", "* world", true, "星号通配符开头"},
		{"hello beautiful world", "hello * world", true, "星号通配符中间"},
		{"hello world", "hello", true, "部分匹配"},
		{"hello world", "goodbye", false, "不匹配"},
		{"test 123", "test _", true, "下划线通配符匹配单个单词"},
		{"test 123 456", "test _", true, "下划线通配符匹配多个单词"}, // 修复了这个测试
		{"I am going to Beijing", "I am going to *", true, "AIML风格模式"},
		{"TEST 123", "test _", true, "大小写不敏感匹配"},
		{"hello", "hello _", false, "下划线需要匹配内容"}, // 下划线必须匹配至少一个单词
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.MatchPattern(tc.input, tc.pattern)
			if result != tc.expected {
				t.Errorf("MatchPattern(%q, %q) = %v, 期望 %v",
					tc.input, tc.pattern, result, tc.expected)
			}
		})
	}
}

func TestExtractWildcards(t *testing.T) {
	matcher := &AdvancedPatternMatcher{}

	testCases := []struct {
		input    string
		pattern  string
		expected map[string]string
		name     string
	}{
		{
			"I am going to Beijing",
			"I am going to *",
			map[string]string{"star": "Beijing", "star1": "Beijing"},
			"提取单个星号并保留大小写",
		},
		{
			"HE IS GOING TO PARIS",
			"HE IS GOING TO *",
			map[string]string{"star": "PARIS", "star1": "PARIS"},
			"提取大写模式并保留大小写",
		},
		{
			"hello world",
			"* world",
			map[string]string{"star": "hello", "star1": "hello"},
			"提取开头的星号",
		},
		{
			"hello beautiful world",
			"hello * world",
			map[string]string{"star": "beautiful", "star1": "beautiful"},
			"提取中间的星号",
		},
		{
			"test 123 456",
			"test *",
			map[string]string{"star": "123 456", "star1": "123 456"},
			"提取多个单词的星号",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.ExtractWildcards(tc.input, tc.pattern)

			for key, expectedValue := range tc.expected {
				if result[key] != expectedValue {
					t.Errorf("ExtractWildcards(%q, %q)[%q] = %q, 期望 %q",
						tc.input, tc.pattern, key, result[key], expectedValue)
				}
			}

			// 检查是否有额外的键
			for key := range result {
				if _, exists := tc.expected[key]; !exists {
					t.Errorf("ExtractWildcards(%q, %q) 返回了意外的键: %q",
						tc.input, tc.pattern, key)
				}
			}
		})
	}
}

func TestUnderscoreMatching(t *testing.T) {
	matcher := &AdvancedPatternMatcher{}

	testCases := []struct {
		input    string
		pattern  string
		expected bool
		name     string
	}{
		{"test 123", "test _", true, "下划线匹配单个单词"},
		{"test abc", "test _", true, "下划线匹配字母单词"},
		{"test 123 456", "test _", true, "下划线匹配多个单词"}, // 修复了这个测试
		{"test", "test _", false, "下划线需要匹配内容"},
		{"test 123 456 789", "test _", true, "下划线匹配多个单词"},
		{"hello world", "hello _", true, "简单下划线匹配"},
		{"hello beautiful world", "hello _ world", true, "下划线在中间匹配多个单词"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.MatchPattern(tc.input, tc.pattern)
			if result != tc.expected {
				t.Errorf("MatchPattern(%q, %q) = %v, 期望 %v",
					tc.input, tc.pattern, result, tc.expected)
			}
		})
	}
}

func TestComplexPatterns(t *testing.T) {
	matcher := &AdvancedPatternMatcher{}

	testCases := []struct {
		input    string
		pattern  string
		expected bool
		name     string
	}{
		{"my name is John", "my name is *", true, "星号在结尾"},
		{"John is my name", "* is my name", true, "星号在开头"},
		{"I love programming very much", "I * very much", true, "星号在中间"},
		{"hello there world", "hello _ world", true, "下划线在中间"},
		{"test this pattern here", "test * here", true, "星号匹配多个单词"},
		{"test this pattern here", "test _ here", true, "下划线匹配多个单词"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := matcher.MatchPattern(tc.input, tc.pattern)
			if result != tc.expected {
				t.Errorf("MatchPattern(%q, %q) = %v, 期望 %v",
					tc.input, tc.pattern, result, tc.expected)
			}
		})
	}
}
