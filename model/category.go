package model

import (
	"math/rand"
	"time"
)

type Category struct {
	Pattern    string
	Templates  []string
	Context    *Context
	SourceFile string // 新增：记录来源文件
}

type Context struct {
	Topic   string
	Memory  map[string]string
	Expires time.Time
}

// NewCategory 创建新的分类
func NewCategory(pattern string, templates []string) Category {
	return Category{
		Pattern:   pattern,
		Templates: templates,
	}
}

// NewCategoryWithSource 创建带来源信息的分类
func NewCategoryWithSource(pattern string, templates []string, sourceFile string) Category {
	return Category{
		Pattern:    pattern,
		Templates:  templates,
		SourceFile: sourceFile,
	}
}

func (c *Category) GetRandomResponse() string {
	if len(c.Templates) == 0 {
		return ""
	}
	if len(c.Templates) == 1 {
		return c.Templates[0]
	}

	rand.Seed(time.Now().UnixNano())
	return c.Templates[rand.Intn(len(c.Templates))]
}

func (c *Category) AddTemplate(template string) {
	c.Templates = append(c.Templates, template)
}
