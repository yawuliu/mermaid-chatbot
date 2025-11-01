// kernel_extensions.go - 内核扩展方法
package kernel

import (
	"fmt"
	"mchat/model"
	"mchat/parser"
	"os"
	"path/filepath"
)

// 获取所有分类（需要添加到Kernel结构体）
func (k *Kernel) GetAllCategories() []model.Category {
	k.categoriesLock.RLock()
	defer k.categoriesLock.RUnlock()

	categories := make([]model.Category, len(k.categories))
	copy(categories, k.categories)
	return categories
}

// 更新分类
func (k *Kernel) UpdateCategory(pattern string, templates []string) {
	k.categoriesLock.Lock()
	updated := false
	for i, category := range k.categories {
		if category.Pattern == pattern {
			k.categories[i].Templates = templates
			updated = true
			break
		}
	}
	k.categoriesLock.Unlock()
	if updated {
		// 这里应该添加持久化逻辑，将更改保存回Mermaid文件
		k.persistChanges()
	}
}

// 替换分类
func (k *Kernel) ReplaceCategory(index int, newCategory model.Category) {
	k.categoriesLock.Lock()
	if index >= 0 && index < len(k.categories) {
		k.categories[index] = newCategory
	}
	k.categoriesLock.Unlock()

	k.persistChanges()
}

// 移除分类
func (k *Kernel) RemoveCategory(index int) {
	k.categoriesLock.Lock()
	if index >= 0 && index < len(k.categories) {
		k.categories = append(k.categories[:index], k.categories[index+1:]...)
	}
	k.categoriesLock.Unlock()

	k.persistChanges()
}

// writeMermaidToFile 将Mermaid内容写入文件
func (k *Kernel) writeMermaidToFile(content, filePath string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建或截断文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入内容
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// ExportCorpus 导出当前语料库到指定文件（可用于手动调用）
func (k *Kernel) ExportCorpus(filePath string) error {
	k.categoriesLock.RLock()
	categories := make([]model.Category, len(k.categories))
	copy(categories, k.categories)
	k.categoriesLock.RUnlock()

	writer := parser.NewAdvancedMermaidWriter()
	content, err := writer.ConvertToOptimizedMermaid(categories)
	if err != nil {
		return err
	}

	return k.writeMermaidToFile(content, filePath)
}
