package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"mchat/kernel"
	"mchat/models"
	"mchat/parser"
)

func main() {
	// 1. 读取Mermaid文件
	content, err := ioutil.ReadFile("corpus/basic_chat.mmd")
	if err != nil {
		log.Fatal("读取文件失败:", err)
	}

	// 2. 解析Mermaid DSL
	mermaidParser := parser.NewMermaidParser()
	graph, err := mermaidParser.Parse(string(content))
	if err != nil {
		log.Fatal("解析Mermaid失败:", err)
	}

	// 3. 创建匹配器
	matcher := kernel.NewMatcher(graph)
	templateEngine := kernel.NewTemplateEngine()

	// 4. 初始化上下文
	ctx := &models.Context{
		Variables: map[string]string{
			"user_name":     "未知用户",
			"current_topic": "general",
		},
		CurrentTopic: "general",
	}

	// 5. 交互循环
	fmt.Println("DeepSeek DSL 聊天机器人已启动 (输入 '退出' 结束)")

	for {
		fmt.Print("你: ")
		var input string
		fmt.Scanln(&input)

		if input == "退出" {
			break
		}

		// 匹配并生成响应
		template, captures := matcher.Match(input, ctx)
		if template != nil {
			response := templateEngine.Render(template, captures, ctx)
			fmt.Printf("机器人: %s\n", response)
		} else {
			fmt.Println("机器人: 抱歉，我不明白您的意思")
		}
	}
}
