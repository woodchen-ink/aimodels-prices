package seo

import (
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

const baseURL = "https://ai-prices.sunai.net"

type PageMeta struct {
	Title       string
	Description string
	Keywords    string
}

// 各路由的 SEO 元信息
var routeMeta = map[string]PageMeta{
	"/": {
		Title:       "AI模型价格 - AI模型价格汇总",
		Description: "专业的AI模型价格管理系统，汇总 OpenAI、Claude、Gemini、Grok、Qwen、DeepSeek 等几十家模型的价格，支持多厂商、多币种对比，提供标准 API 接口。",
		Keywords:    "AI模型价格,GPT价格,Claude价格,Gemini价格,DeepSeek价格,API定价,大模型价格对比",
	},
	"/prices": {
		Title:       "价格列表 - AI模型价格汇总",
		Description: "查看和对比各大AI模型的详细价格信息，包括输入输出 Token 价格、多币种换算，支持按厂商筛选和模型搜索。",
		Keywords:    "AI模型价格列表,Token价格,模型定价对比,API价格查询",
	},
	"/providers": {
		Title:       "模型厂商 - AI模型价格汇总",
		Description: "浏览 AI 模型厂商列表，包括 OpenAI、Anthropic、Google、阿里云、百度等主流大模型服务商信息。",
		Keywords:    "AI模型厂商,OpenAI,Anthropic,Google AI,大模型服务商",
	},
	"/login": {
		Title:       "登录 - AI模型价格汇总",
		Description: "登录 AI 模型价格管理系统，提交和管理模型价格数据。",
		Keywords:    "AI模型价格,用户登录",
	},
}

// 默认 SEO（未匹配到的路由使用首页信息）
var defaultMeta = routeMeta["/"]

var (
	indexTemplate string
	templateOnce  sync.Once
)

// loadTemplate 读取并缓存 index.html 模板
func loadTemplate(staticDir string) string {
	templateOnce.Do(func() {
		data, err := os.ReadFile(staticDir + "/index.html")
		if err != nil {
			indexTemplate = ""
			return
		}
		indexTemplate = string(data)
	})
	return indexTemplate
}

// RenderIndex 根据请求路径渲染带有正确 SEO 信息的 index.html
func RenderIndex(c *gin.Context, staticDir string) {
	tmpl := loadTemplate(staticDir)
	if tmpl == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	path := c.Request.URL.Path
	meta, ok := routeMeta[path]
	if !ok {
		meta = defaultMeta
	}

	canonical := baseURL + path

	html := tmpl
	html = strings.ReplaceAll(html, "{{SEO_TITLE}}", meta.Title)
	html = strings.ReplaceAll(html, "{{SEO_DESCRIPTION}}", meta.Description)
	html = strings.ReplaceAll(html, "{{SEO_KEYWORDS}}", meta.Keywords)
	html = strings.ReplaceAll(html, "{{SEO_CANONICAL}}", canonical)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
