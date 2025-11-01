package adapter

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"mchat/kernel"
)

type ConsoleAdapter struct {
	kernel  *kernel.Kernel
	running bool
	scanner *bufio.Scanner
	users   map[string]string // userID -> username
}

func NewConsoleAdapter(kernel *kernel.Kernel) *ConsoleAdapter {
	return &ConsoleAdapter{
		kernel:  kernel,
		running: false,
		scanner: bufio.NewScanner(os.Stdin),
		users:   make(map[string]string),
	}
}

func (c *ConsoleAdapter) Run() error {
	c.running = true

	// 初始化默认用户
	defaultUserID := c.generateUserID()
	c.users[defaultUserID] = "默认用户"

	currentUserID := defaultUserID

	fmt.Println("=== Mermaid聊天机器人控制台 ===")
	fmt.Printf("当前用户: %s (ID: %s)\n", c.users[currentUserID], currentUserID)
	fmt.Println("命令:")
	fmt.Println("  /switch <用户名> - 切换用户")
	fmt.Println("  /users - 显示所有用户")
	fmt.Println("  /exit - 退出")
	fmt.Println("  /debug - 查看调试信息")
	fmt.Println("----------------------------")

	for c.running {
		fmt.Printf("%s: ", c.users[currentUserID])

		if !c.scanner.Scan() {
			break
		}

		input := c.scanner.Text()

		// 处理命令
		if strings.HasPrefix(input, "/") {
			currentUserID = c.handleCommand(input, currentUserID)
			continue
		}

		if strings.ToLower(input) == "exit" {
			break
		}

		// 使用当前用户ID处理输入
		response := c.kernel.ProcessInput(currentUserID, input)
		fmt.Printf("Bot: %s\n", response)
	}

	return c.scanner.Err()
}
func (c *ConsoleAdapter) generateUserID() string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("user_%d", timestamp)
}

func (c *ConsoleAdapter) handleCommand(cmd string, currentUserID string) string {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return currentUserID
	}

	switch parts[0] {
	case "/switch":
		if len(parts) < 2 {
			fmt.Println("用法: /switch <用户名>")
			return currentUserID
		}

		username := parts[1]
		// 查找或创建用户
		newUserID := c.findOrCreateUser(username)
		fmt.Printf("已切换到用户: %s\n", username)
		return newUserID

	case "/users":
		fmt.Println("当前用户列表:")
		for userID, username := range c.users {
			marker := ""
			if userID == currentUserID {
				marker = " [当前]"
			}
			fmt.Printf("  %s (ID: %s)%s\n", username, userID, marker)
		}

	case "/exit":
		c.running = false
		// 处理调试命令
	case "/debug":
		debugInfo := c.kernel.GetDebugInfo(currentUserID)
		fmt.Printf("Debug: %s\n", debugInfo)
	default:
		fmt.Printf("未知命令: %s\n", parts[0])
	}

	return currentUserID
}

func (c *ConsoleAdapter) findOrCreateUser(username string) string {
	// 首先检查是否已存在该用户名
	for userID, existingUsername := range c.users {
		if existingUsername == username {
			return userID
		}
	}

	// 创建新用户
	newUserID := c.generateUserID()
	c.users[newUserID] = username
	return newUserID
}

// 其他方法保持不变...
func (c *ConsoleAdapter) Stop() {
	c.running = false
}

func (c *ConsoleAdapter) SendResponse(response string) {
	fmt.Printf("Bot: %s\n", response)
}
