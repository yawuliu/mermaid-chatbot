// kernel/condition_parser_test.go
package parser

import (
	"testing"
)

func TestEdgeParsing(t *testing.T) {
	parser := &ConditionParser{debug: true}

	testCases := []struct {
		line         string
		expectedFrom string
		expectedTo   string
		name         string
	}{
		{
			"A[I AM *] --> B{Condition: predicate isanumber}",
			"A", "B",
			"复杂边解析",
		},
		{
			"B -->|true| C[<srai>MY AGE IS <star/></srai>]",
			"B", "C",
			"带标签边解析",
		},
		{
			"B -->|*| D{Condition: predicate isaname}",
			"B", "D",
			"带通配符标签边解析",
		},
		{
			"A --> B",
			"A", "B",
			"简单边解析",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			graph := NewGraph()
			parser.parseConditionEdge(tc.line, graph)

			// 检查边是否被正确添加
			edges, exists := graph.Edges[tc.expectedFrom]
			if !exists {
				t.Errorf("边 %s -> %s 没有被添加", tc.expectedFrom, tc.expectedTo)
				return
			}

			found := false
			for _, to := range edges {
				if to == tc.expectedTo {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("期望边 %s -> %s，但找到的边是: %v",
					tc.expectedFrom, tc.expectedTo, edges)
			}
		})
	}
}

func TestFullConditionalParsing(t *testing.T) {
	parser := &ConditionParser{debug: true}

	content := `flowchart TD
    A[I AM *] --> B{Condition: predicate isanumber}
    B -->|true| C[<srai>MY AGE IS <star/></srai>]
    B -->|*| D{Condition: predicate isaname}
    D -->|true| E[<srai>MY NAME IS <star/></srai>]
    D -->|*| F[<srai>IAMRESPONSE</srai>]`

	categories, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// 检查是否找到了 I AM * 模式
	foundIAMPattern := false
	for _, category := range categories {
		t.Logf("找到分类: 模式=%s, 条件数=%d", category.Pattern, len(category.Conditions))
		if category.Pattern == "I AM *" {
			foundIAMPattern = true
			if len(category.Conditions) == 0 {
				t.Errorf("I AM * 分类没有找到任何条件")
			} else {
				t.Logf("I AM * 分类的条件:")
				for i, condition := range category.Conditions {
					t.Logf("  条件 %d: 类型=%s, 名称=%s, 响应=%s",
						i, condition.Condition.Type, condition.Condition.Name, condition.Response)
				}
			}
			break
		}
	}

	if !foundIAMPattern {
		t.Errorf("没有找到 I AM * 模式分类")
		t.Logf("找到的所有分类:")
		for i, category := range categories {
			t.Logf("  分类 %d: %s", i, category.Pattern)
		}
	}
}
