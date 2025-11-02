// kernel/loader.go
package kernel

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"mchat/model"
	"mchat/parser"
)

//	func isMermaidFile(filename string) bool {
//		extensions := []string{".mmd", ".mermaid", ".md"}
//		for _, ext := range extensions {
//			if len(filename) > len(ext) && filename[len(filename)-len(ext):] == ext {
//				return true
//			}
//		}
//		return false
//	}
//
// isMermaidFile 检查是否为Mermaid文件
func isMermaidFile(filename string) bool {
	extensions := []string{".mmd", ".mermaid"}
	lowerFilename := strings.ToLower(filename)

	for _, ext := range extensions {
		if len(filename) > len(ext) && strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}

	return false
}

// loadMermaidCorpus 加载Mermaid语料文件并记录来源
func (k *Kernel) loadMermaidCorpus(corpusPath string) error {
	files, err := os.ReadDir(corpusPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if isMermaidFile(file.Name()) {
			filePath := filepath.Join(corpusPath, file.Name())
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("警告: 无法读取文件 %s: %v", filePath, err)
				continue
			}

			categories, err := parser.ParseMermaidWithSource(string(content), file.Name())
			if err != nil {
				log.Printf("警告: 解析文件 %s 失败: %v", filePath, err)
				continue
			}

			// 为每个分类设置来源文件
			for i := range categories {
				categories[i].SourceFile = file.Name()
				k.AddCategory(categories[i])
			}

			log.Printf("从 %s 加载了 %d 个分类", file.Name(), len(categories))
		}
	}

	return nil
}

func isConditionalMermaidFile(filename string) bool {
	return isMermaidFile(filename)
}

func (k *Kernel) loadConditionalCorpus(corpusPath string) error {
	files, err := os.ReadDir(corpusPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if isConditionalMermaidFile(file.Name()) {
			filePath := corpusPath + "/" + file.Name()
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("警告: 无法读取文件 %s: %v", filePath, err)
				continue
			}

			categories, err := parser.ParseConditionalMermaid(string(content))
			if err != nil {
				log.Printf("警告: 解析条件文件 %s 失败: %v", filePath, err)
				continue
			}

			for _, category := range categories {
				k.AddConditionalCategory(category)
			}

			log.Printf("从 %s 加载了 %d 个条件分类", file.Name(), len(categories))
		}
	}

	return nil
}

//	func loadMermaidCorpus(kernel *kernel.Kernel, corpusPath string) int {
//		log.Printf("开始加载语料库，路径: %s", corpusPath)
//		// 检查目录是否存在
//		if _, err := os.Stat(corpusPath); os.IsNotExist(err) {
//			log.Printf("错误: 语料库目录不存在: %s", corpusPath)
//			return 0
//		}
//
//		files, err := os.ReadDir(corpusPath)
//		if err != nil {
//			log.Printf("错误: 无法读取目录: %v", err)
//			return 0
//		}
//
//		log.Printf("找到 %d 个文件在目录中", len(files))
//		totalCategories := 0
//
//		for _, file := range files {
//			filename := file.Name()
//			log.Printf("检查文件: %s", filename)
//			if isMermaidFile(filename) {
//				log.Printf("识别为Mermaid文件: %s", filename)
//				filePath := corpusPath + "/" + filename
//				content, err := os.ReadFile(filePath)
//				if err != nil {
//					log.Printf("警告: 无法读取文件 %s: %v", filePath, err)
//					continue
//				}
//				log.Printf("文件内容长度: %d 字节", len(content))
//				log.Printf("文件内容前100字符: %s", string(content)[:min(100, len(content))])
//
//				categories, err := parser.ParseMermaid(string(content))
//				if err != nil {
//					log.Printf("警告: 解析文件 %s 失败: %v", filePath, err)
//					continue
//				}
//
//				log.Printf("从 %s 解析出 %d 个类别", filename, len(categories))
//
//				for i, category := range categories {
//					log.Printf("类别 %d: 模式='%s', 响应=%v", i, category.Pattern, category.Templates)
//					kernel.AddCategory(category)
//					totalCategories++
//				}
//
//				log.Printf("Loaded %d categories from %s", len(categories), file.Name())
//			} else {
//				log.Printf("跳过非Mermaid文件: %s", filename)
//			}
//		}
//
//		return totalCategories
//	}
//
// ReloadCorpus 重新加载语料库
func (k *Kernel) ReloadCorpus() error {
	k.categoriesLock.Lock()
	k.categories = make([]model.Category, 0)
	k.categoriesLock.Unlock()

	return k.loadConditionalCorpus(k.config.CorpusPath)
}

// ReloadCorpus 重新加载语料库
//func (k *Kernel) ReloadCorpus() error {
//	// 实现重新加载逻辑
//	// 这需要清空当前分类并重新解析所有Mermaid文件
//	return nil
//}
