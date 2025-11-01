package main

import (
	"fmt"
	"log"
	"mchat/model"
	"os"
	"os/signal"
	"syscall"

	"mchat/adapter"
	"mchat/config"
	"mchat/kernel"
	"mchat/parser"
)

// 在main.go中添加测试函数
func testMermaidParsing() {
	testContents := []string{
		`flowchart TD
    A[Hello] --> B[Hi there!]
    A --> C[Greetings!]`,
		`flowchart TD
    开始[你好] --> 响应1[你好！欢迎！]`,
		`graph TD
    A[Test] --> B[Response]`,
	}

	for i, content := range testContents {
		log.Printf("=== 测试用例 %d ===", i+1)
		categories, err := parser.ParseMermaid(content)
		if err != nil {
			log.Printf("解析失败: %v", err)
		} else {
			log.Printf("解析成功，找到 %d 个类别", len(categories))
			for j, category := range categories {
				log.Printf("  类别 %d: Pattern='%s', Templates=%v", j, category.Pattern, category.Templates)
			}
		}
	}
}

func main() {
	// 先运行解析测试
	testMermaidParsing()

	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}
	log.Printf("配置加载成功: corpus_path=%s", cfg.CorpusPath)
	// 初始化内核
	kernel := kernel.NewKernel()
	kernel.SetConfig(cfg)

	// 加载语料库
	if err := kernel.LoadCorpus(cfg.CorpusPath); err != nil {
		log.Fatalf("加载语料库失败: %v", err)
	}

	totalCategories := kernel.GetCategoryCount()
	if totalCategories == 0 {
		log.Printf("警告: 没有加载任何类别，请检查语料库文件和解析逻辑")
		// 可以在这里添加一些默认类别作为回退
		addFallbackCategories(kernel)
	}

	log.Printf("成功加载 %d 个对话类别", totalCategories)

	// 初始化消息适配器（这里使用控制台适配器）
	adapter := adapter.NewConsoleAdapter(kernel)

	fmt.Printf("Mermaid ChatBot 已启动，共加载 %d 个对话类别\n", totalCategories)
	fmt.Println("输入 'exit' 退出程序")

	// 处理中断信号
	setupGracefulShutdown(adapter)

	// 启动机器人
	if err := adapter.Run(); err != nil {
		log.Fatalf("适配器错误: %v", err)
	}
}

func setupGracefulShutdown(adapter adapter.Adapter) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down...")
		adapter.Stop()
		os.Exit(0)
	}()
}

func addFallbackCategories(k *kernel.Kernel) {
	log.Println("添加回退对话类别...")
	fallbackCategories := []model.Category{
		model.NewCategoryWithSource("你好", []string{"你好！", "嗨！", "欢迎！"}, "corpus/greeting.mmd"),
		model.NewCategoryWithSource("帮助", []string{"我可以帮你解答问题。", "请问你需要什么帮助？"}, "corpus/greeting.mmd"),
	}

	for _, category := range fallbackCategories {
		k.AddCategory(category)
	}
}
