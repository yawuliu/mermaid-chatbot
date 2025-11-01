// parser/mermaid_writer_test.go
package parser

import (
	"strings"
	"testing"

	"mchat/model"
)

func TestMermaidWriterQuoteHandling(t *testing.T) {
	writer := NewAdvancedMermaidWriter()

	testCases := []struct {
		input    string
		expected string
		name     string
	}{
		{"简单文本", `"简单文本"`, "中文文本"},
		{"hello", "hello", "简单英文"},
		{"hello world", `"hello world"`, "带空格的英文"},
		{"你好世界", `"你好世界"`, "中文文本"},
		{`"已有引号"`, `"已有引号"`, "已有引号"},
		{`"""多层引号"""`, `"多层引号"`, "多层引号"},
		{"test123", "test123", "英文数字"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := writer.cleanNodeText(tc.input)
			if result != tc.expected {
				t.Errorf("cleanNodeText(%q) = %q, 期望 %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestMermaidGeneration(t *testing.T) {
	writer := NewAdvancedMermaidWriter()

	categories := []model.Category{
		model.NewCategory("你好", []string{"你好！", "欢迎！"}),
		model.NewCategory("help", []string{"I can help", "How can I assist?"}),
		model.NewCategory("天气", []string{"今天天气不错", `"明天会下雨"`}),
	}

	content, err := writer.ConvertToOptimizedMermaid(categories)
	if err != nil {
		t.Fatalf("生成Mermaid失败: %v", err)
	}

	t.Logf("生成的Mermaid内容:\n%s", content)

	// 检查是否有多余的引号
	if strings.Contains(content, `""`) {
		t.Errorf("发现多余的双引号: %s", content)
	}
}
