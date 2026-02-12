import { createRouter, createWebHistory } from 'vue-router'
import Prices from '../views/Prices.vue'
import Providers from '../views/Providers.vue'
import Login from '../views/Login.vue'
import Home from '../views/Home.vue'

const SITE_NAME = 'AI模型价格汇总'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home,
      meta: {
        title: `AI模型价格 - ${SITE_NAME}`,
        description: '专业的AI模型价格管理系统，汇总 OpenAI、Claude、Gemini、Grok、Qwen、DeepSeek 等几十家模型的价格，支持多厂商、多币种对比，提供标准 API 接口。',
        keywords: 'AI模型价格,GPT价格,Claude价格,Gemini价格,DeepSeek价格,API定价,大模型价格对比'
      }
    },
    {
      path: '/prices',
      name: 'prices',
      component: Prices,
      meta: {
        title: `价格列表 - ${SITE_NAME}`,
        description: '查看和对比各大AI模型的详细价格信息，包括输入输出 Token 价格、多币种换算，支持按厂商筛选和模型搜索。',
        keywords: 'AI模型价格列表,Token价格,模型定价对比,API价格查询'
      }
    },
    {
      path: '/providers',
      name: 'providers',
      component: Providers,
      meta: {
        title: `模型厂商 - ${SITE_NAME}`,
        description: '浏览 AI 模型厂商列表，包括 OpenAI、Anthropic、Google、阿里云、百度等主流大模型服务商信息。',
        keywords: 'AI模型厂商,OpenAI,Anthropic,Google AI,大模型服务商'
      }
    },
    {
      path: '/login',
      name: 'login',
      component: Login,
      meta: {
        title: `登录 - ${SITE_NAME}`,
        description: '登录 AI 模型价格管理系统，提交和管理模型价格数据。',
        keywords: ''
      }
    }
  ]
})

// 动态更新页面 title 和 meta 标签
router.afterEach((to) => {
  const { title, description, keywords } = to.meta

  // 更新 title
  document.title = title || SITE_NAME

  // 更新 meta description
  const descEl = document.querySelector('meta[name="description"]')
  if (descEl && description) {
    descEl.setAttribute('content', description)
  }

  // 更新 meta keywords
  let keywordsEl = document.querySelector('meta[name="keywords"]')
  if (keywords) {
    if (!keywordsEl) {
      keywordsEl = document.createElement('meta')
      keywordsEl.setAttribute('name', 'keywords')
      document.head.appendChild(keywordsEl)
    }
    keywordsEl.setAttribute('content', keywords)
  }

  // 更新 Open Graph 标签
  const ogTags = {
    'og:title': title || SITE_NAME,
    'og:description': description || '',
    'og:url': window.location.href
  }
  for (const [property, content] of Object.entries(ogTags)) {
    let el = document.querySelector(`meta[property="${property}"]`)
    if (el) {
      el.setAttribute('content', content)
    }
  }

  // 更新 canonical
  let canonical = document.querySelector('link[rel="canonical"]')
  if (canonical) {
    canonical.setAttribute('href', window.location.origin + to.path)
  }
})

export default router