// kernel/management.go
package kernel

import (
	"fmt"
	"mchat/model"
	"mchat/parser"
	"path/filepath"
	"strings"
	"time"
)

// ManagementCommand 管理命令处理
func (k *Kernel) HandleManagementCommand(userID, command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "无效的管理命令"
	}

	switch parts[0] {
	case "/backup":
		return k.handleBackupCommand(parts[1:])
	case "/restore":
		return k.handleRestoreCommand(parts[1:])
	case "/list-backups":
		return k.handleListBackupsCommand()
	case "/reload":
		return k.handleReloadCommand()
	default:
		return "未知的管理命令"
	}
}

func (k *Kernel) handleBackupCommand(args []string) string {
	// 手动创建备份
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("manual_backup_%s.mmd", timestamp)
	backupPath := filepath.Join(k.config.CorpusPath, "backups", backupName)

	categories := k.getCategoriesCopy()
	if err := k.persistenceManager.writeMermaidToFile(
		k.categoriesToMermaid(categories), backupPath); err != nil {
		return fmt.Sprintf("备份失败: %v", err)
	}

	return fmt.Sprintf("手动备份已创建: %s", backupName)
}

func (k *Kernel) handleRestoreCommand(args []string) string {
	if len(args) == 0 {
		return "请指定要恢复的备份文件名"
	}

	if err := k.persistenceManager.RestoreFromBackup(args[0]); err != nil {
		return fmt.Sprintf("恢复失败: %v", err)
	}

	// 重新加载语料
	if err := k.ReloadCorpus(); err != nil {
		return fmt.Sprintf("恢复成功但重新加载失败: %v", err)
	}

	return "恢复成功并重新加载语料"
}

func (k *Kernel) handleListBackupsCommand() string {
	backups, err := k.persistenceManager.ListBackups()
	if err != nil {
		return fmt.Sprintf("获取备份列表失败: %v", err)
	}

	if len(backups) == 0 {
		return "没有找到备份文件"
	}

	var result strings.Builder
	result.WriteString("可用备份文件:\n")
	for _, backup := range backups {
		result.WriteString(fmt.Sprintf("  - %s\n", backup))
	}
	return result.String()
}

func (k *Kernel) handleReloadCommand() string {
	if err := k.ReloadCorpus(); err != nil {
		return fmt.Sprintf("重新加载失败: %v", err)
	}
	return fmt.Sprintf("语料库重新加载成功，共加载 %d 个对话类别", len(k.categories))
}

// categoriesToMermaid 将分类转换为Mermaid格式
func (k *Kernel) categoriesToMermaid(categories []model.Category) string {
	writer := parser.NewAdvancedMermaidWriter()
	content, err := writer.ConvertToOptimizedMermaid(categories)
	if err != nil {
		return "转换失败"
	}
	return content
}
