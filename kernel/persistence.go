// kernel/persistence.go
package kernel

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mchat/model"
	"mchat/parser"
)

// 文件修改记录，用于跟踪哪些文件被修改过
type FileModification struct {
	OriginalFile string
	ModifiedFile string
	Timestamp    string
}

// PersistenceManager 持久化管理器
type PersistenceManager struct {
	kernel        *Kernel
	modifiedFiles map[string]*FileModification
	corpusPath    string
}

func NewPersistenceManager(kernel *Kernel, corpusPath string) *PersistenceManager {
	return &PersistenceManager{
		kernel:        kernel,
		modifiedFiles: make(map[string]*FileModification),
		corpusPath:    corpusPath,
	}
}

// persistChanges 持久化修改到原始文件
func (k *Kernel) persistChanges() {
	if k.persistenceManager == nil {
		log.Printf("警告: 持久化管理器未初始化")
		return
	}

	categories := k.getCategoriesCopy()
	k.persistenceManager.SaveToOriginalFiles(categories)
}

// 持久化更改到Mermaid文件
//func (k *Kernel) persistChanges() {
//	// 这里实现将内存中的categories转换回Mermaid格式并保存到文件
//	// 需要调用parser的逆向转换功能
//	// k.parser.SaveToMermaid(k.categories, "corpus/updated.mmd")
//	k.categoriesLock.RLock()
//	categories := make([]model.Category, len(k.categories))
//	copy(categories, k.categories)
//	k.categoriesLock.RUnlock()
//
//	// 确定输出文件路径
//	outputPath := k.getOutputFilePath()
//
//	// 使用高级写入器生成优化的Mermaid
//	writer := parser.NewAdvancedMermaidWriter()
//	content, err := writer.ConvertToOptimizedMermaid(categories)
//	if err != nil {
//		log.Printf("生成Mermaid内容失败: %v", err)
//		return
//	}
//
//	// 写入文件
//	if err := k.writeMermaidToFile(content, outputPath); err != nil {
//		log.Printf("写入Mermaid文件失败: %v", err)
//	} else {
//		log.Printf("语料库已成功保存到: %s", outputPath)
//	}
//}

// getCategoriesCopy 获取分类的副本（内部处理锁）
func (k *Kernel) getCategoriesCopy() []model.Category {
	k.categoriesLock.RLock()
	defer k.categoriesLock.RUnlock()

	categories := make([]model.Category, len(k.categories))
	copy(categories, k.categories)
	return categories
}

// SaveToOriginalFiles 将分类保存回原始的Mermaid文件
func (pm *PersistenceManager) SaveToOriginalFiles(categories []model.Category) {
	// 按原始文件分组分类
	fileCategories := pm.groupCategoriesByOriginalFile(categories)

	// 更新每个原始文件
	for filename, fileCats := range fileCategories {
		if err := pm.updateMermaidFile(filename, fileCats); err != nil {
			log.Printf("更新文件 %s 失败: %v", filename, err)
		} else {
			log.Printf("成功更新文件: %s", filename)
		}
	}

	// 清理旧备份
	pm.cleanupOldBackups()
}

// groupCategoriesByOriginalFile 按原始文件对分类进行分组
func (pm *PersistenceManager) groupCategoriesByOriginalFile(categories []model.Category) map[string][]model.Category {
	// 这里我们需要知道每个分类最初来自哪个文件
	// 由于目前没有这个信息，我们可以采用智能推断或配置映射

	// 方案1: 按主题分组到不同文件
	// return pm.groupByTopic(categories)

	// 方案2: 如果记录了来源，使用记录的信息
	return pm.groupByRecordedSource(categories)
}

func (pm *PersistenceManager) groupByRecordedSource(categories []model.Category) map[string][]model.Category {
	fileGroups := make(map[string][]model.Category)

	for _, category := range categories {
		filename := category.SourceFile

		// 如果分类没有来源文件（新增的分类），使用智能推断
		if filename == "" {
			filename = pm.inferFilename(category.Pattern)
			log.Printf("分类 '%s' 没有来源文件，推断保存到: %s", category.Pattern, filename)
		}

		fileGroups[filename] = append(fileGroups[filename], category)
	}

	return fileGroups
}

// groupByTopic 按主题对分类进行分组
func (pm *PersistenceManager) groupByTopic(categories []model.Category) map[string][]model.Category {
	fileGroups := make(map[string][]model.Category)

	// 简单的主题推断规则
	for _, category := range categories {
		filename := pm.inferFilename(category.Pattern)
		fileGroups[filename] = append(fileGroups[filename], category)
	}

	return fileGroups
}

// inferFilename 根据模式推断文件名
func (pm *PersistenceManager) inferFilename(pattern string) string {
	// 基于关键词的文件名推断
	lowerPattern := strings.ToLower(pattern)

	switch {
	case strings.Contains(lowerPattern, "你好") ||
		strings.Contains(lowerPattern, "hello") ||
		strings.Contains(lowerPattern, "嗨"):
		return "greetings.mmd"
	case strings.Contains(lowerPattern, "谢谢") ||
		strings.Contains(lowerPattern, "感谢"):
		return "thanks.mmd"
	case strings.Contains(lowerPattern, "再见") ||
		strings.Contains(lowerPattern, "bye"):
		return "farewell.mmd"
	case strings.Contains(lowerPattern, "帮助") ||
		strings.Contains(lowerPattern, "help"):
		return "help.mmd"
	case strings.Contains(lowerPattern, "天气") ||
		strings.Contains(lowerPattern, "weather"):
		return "weather.mmd"
	default:
		return "general.mmd"
	}
}

// updateMermaidFile 更新单个Mermaid文件
func (pm *PersistenceManager) updateMermaidFile(filename string, categories []model.Category) error {
	filePath := filepath.Join(pm.corpusPath, filename)

	writer := parser.NewAdvancedMermaidWriter()
	content, err := writer.ConvertToOptimizedMermaid(categories)
	if err != nil {
		return fmt.Errorf("生成Mermaid内容失败: %v", err)
	}

	// 创建备份
	if err := pm.createBackup(filePath); err != nil {
		log.Printf("创建备份失败: %v", err)
	}

	// 写入文件
	return pm.writeMermaidToFile(content, filePath)
}

// createBackup 创建文件备份
func (pm *PersistenceManager) createBackup(originalPath string) error {
	if _, err := os.Stat(originalPath); os.IsNotExist(err) {
		return nil // 文件不存在，无需备份
	}

	backupDir := filepath.Join(pm.corpusPath, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return err
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	backupName := fmt.Sprintf("%s.backup.%s",
		filepath.Base(originalPath), timestamp)
	backupPath := filepath.Join(backupDir, backupName)

	data, err := os.ReadFile(originalPath)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, data, 0644)
}

// writeMermaidToFile 写入Mermaid文件
func (pm *PersistenceManager) writeMermaidToFile(content, filePath string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}
