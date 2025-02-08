<template>
  <div class="home">
    <el-card class="intro-card">
      <template #header>
        <div class="card-header">
          <h1>AI模型价格</h1>
        </div>
      </template>
      <div class="content">
        <h2>项目简介</h2>
        <p>这是一个专门用于管理AI模型价格的系统，支持多模型厂商、多币种的价格管理，并提供标准的API接口供其他系统调用。</p>
        
        <h2>主要功能</h2>
        <ul>
          <li>模型厂商管理：添加、编辑和删除AI模型模型厂商</li>
          <li>价格管理：设置和更新各个模型的价格</li>
          <li>多币种支持：支持USD和CNY两种货币</li>
          <li>审核流程：价格变更需要管理员审核</li>
          <li>API接口：提供标准的REST API</li>
        </ul>

        <h2>交流讨论</h2>
        <p>请在帖子下留言: <a href="https://q58.pro/t/topic/277?u=wood" target="_blank">https://q58.pro/t/topic/277</a></p>

        <h2>API文档</h2>
        <el-collapse>
          <el-collapse-item title="获取价格倍率">
            <div class="api-doc">
              <div class="api-url">
                <span class="method">GET</span>
                <el-tooltip content="点击复制" placement="top">
                  <span class="url" @click="copyToClipboard(origin + '/api/prices/rates')">
                    {{ origin }}/api/prices/rates
                  </span>
                </el-tooltip>
              </div>
              <p>获取所有已审核通过的价格的倍率信息</p>
              <h4>响应示例：</h4>
              <pre>
[
  {
    "model": "babbage-002",
    "type": "tokens",
    "channel_type": 1,
    "input": 0.2,
    "output": 0.2
  }
]</pre>
              <h4>字段说明：</h4>
              <ul>
                <li>model: 模型名称</li>
                <li>type: 计费类型（tokens/times）</li>
                <li>channel_type: 模型厂商ID</li>
                <li>input: 输入价格倍率</li>
                <li>output: 输出价格倍率</li>
              </ul>
            </div>
          </el-collapse-item>

          <el-collapse-item title="获取价格列表">
            <div class="api-doc">
              <div class="api-url">
                <span class="method">GET</span>
                <el-tooltip content="点击复制" placement="top">
                  <span class="url" @click="copyToClipboard(origin + '/api/prices')">
                    {{ origin }}/api/prices
                  </span>
                </el-tooltip>
              </div>
              <p>获取所有价格信息，包括待审核的价格</p>
              <h4>响应示例：</h4>
              <pre>
[
  {
    "id": 1,
    "model": "gpt-4",
    "billing_type": "tokens",
    "channel_type": "1",
    "currency": "USD",
    "input_price": 0.01,
    "output_price": 0.03,
    "price_source": "官方",
    "status": "approved"
  }
]</pre>
            </div>
          </el-collapse-item>

          <el-collapse-item title="获取模型厂商">
            <div class="api-doc">
              <div class="api-url">
                <span class="method">GET</span>
                <el-tooltip content="点击复制" placement="top">
                  <span class="url" @click="copyToClipboard(origin + '/api/providers')">
                    {{ origin }}/api/providers
                  </span>
                </el-tooltip>
              </div>
              <p>获取所有模型厂商信息</p>
              <h4>响应示例：</h4>
              <pre>
[
  {
    "id": 1,
    "name": "OpenAI",
    "icon": "https://example.com/openai.png",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "created_by": "admin"
  }
]</pre>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.home {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.intro-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

h1 {
  margin: 0;
  font-size: 24px;
  color: #409EFF;
}

h2 {
  margin-top: 30px;
  margin-bottom: 15px;
  font-size: 20px;
  color: #303133;
}

.content {
  line-height: 1.6;
  color: #606266;
}

ul {
  padding-left: 20px;
}

li {
  margin-bottom: 8px;
}

.api-doc {
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 4px;
}

pre {
  background-color: #f1f1f1;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
}

h4 {
  margin: 15px 0 10px;
  color: #303133;
}

.api-url {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  background-color: #f1f1f1;
  padding: 8px 12px;
  border-radius: 4px;
}

.method {
  color: #67C23A;
  font-weight: bold;
  background-color: #f0f9eb;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 14px;
}

.url {
  color: #409EFF;
  cursor: pointer;
  font-family: monospace;
  font-size: 14px;
}

.url:hover {
  text-decoration: underline;
}
</style>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const origin = ref('')

onMounted(() => {
  origin.value = window.location.origin
})

// 添加页面元信息
const meta = {
  title: 'AI模型价格 - 首页',
  description: '专业的AI模型价格管理系统，支持多模型厂商、多币种的价格管理，提供标准的API接口。',
  keywords: 'AI模型,价格管理,API接口,OpenAI,Azure,Anthropic'
}

const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch (err) {
    ElMessage.error('复制失败')
  }
}

defineExpose({ meta })
</script> 