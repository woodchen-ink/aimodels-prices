<template>
  <el-container>
    <el-header height="60px">
      <div class="nav-container">
        <div class="nav-left">
          <router-link to="/" class="logo">
            AI模型价格
          </router-link>
          <div class="nav-buttons">
            <el-button @click="$router.push('/prices')" :type="$route.path === '/prices' ? 'primary' : ''">价格列表</el-button>
            <el-button @click="$router.push('/providers')" :type="$route.path === '/providers' ? 'primary' : ''">模型厂商</el-button>
          </div>
        </div>
        <div class="auth-buttons">
          <template v-if="globalUser">
            <span class="user-info">
              <el-icon><User /></el-icon>
              {{ globalUser.username }}
              <el-tag v-if="globalUser.role === 'admin'" size="small" type="success">管理员</el-tag>
            </span>
            <el-button @click="handleLogout">退出</el-button>
          </template>
          <el-button v-else type="primary" @click="handleLogin" :loading="loading">登录</el-button>
        </div>
      </div>
    </el-header>

    <el-main>
      <div class="content-container">
        <router-view v-slot="{ Component }">
          <component :is="Component" :user="globalUser" />
        </router-view>
      </div>
    </el-main>

    <el-footer height="60px">
      <div class="footer-content">
        <p>© 2025 Q58 AI模型价格 | <a href="https://q58.club/t/topic/277?u=wood" target="_blank">介绍帖子</a></p>
      </div>
    </el-footer>
  </el-container>
</template>

<script setup>
import { ref, onMounted, provide } from 'vue'
import { User } from '@element-plus/icons-vue'
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const globalUser = ref(null)
const loading = ref(false)

const updateGlobalUser = (user) => {
  globalUser.value = user
}

provide('updateGlobalUser', updateGlobalUser)

const handleLogin = async () => {
  loading.value = true
  try {
    const { data } = await axios.post('/api/auth/login')
    if (data.auth_url) {
      // 直接重定向到授权页面
      window.location.href = data.auth_url
    } else {
      // 处理开发环境下的直接登录
      const { data: userData } = await axios.get('/api/auth/status')
      updateGlobalUser(userData.user)
      ElMessage.success('登录成功')
    }
  } catch (error) {
    console.error('Failed to login:', error)
    ElMessage.error('登录失败')
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const { data } = await axios.get('/api/auth/status')
    globalUser.value = data.user
  } catch (error) {
    console.error('Failed to get auth status:', error)
  }
})

const handleLogout = async () => {
  try {
    await axios.post('/api/auth/logout')
    globalUser.value = null
    router.push('/login')
  } catch (error) {
    console.error('Failed to logout:', error)
  }
}
</script>

<style>
.el-container {
  min-height: 100vh;
}

.el-header {
  background-color: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  position: fixed;
  width: 100%;
  z-index: 100;
  padding: 0;
}

.el-main {
  padding-top: 80px;
  background-color: #f5f7fa;
}

.nav-container {
  max-width: 1400px;
  margin: 0 auto;
  height: 60px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
}

.nav-left {
  display: flex;
  align-items: center;
  gap: 40px;
}

.logo {
  font-size: 20px;
  font-weight: bold;
  color: #409EFF;
  text-decoration: none;
  white-space: nowrap;
}

.nav-buttons {
  display: flex;
  gap: 10px;
}

.content-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 20px;
}

.auth-buttons {
  display: flex;
  align-items: center;
  gap: 16px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #606266;
}

.el-footer {
  background-color: #fff;
  border-top: 1px solid #e4e7ed;
}

.footer-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 20px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.footer-content a {
  color: #409EFF;
  text-decoration: none;
}

.footer-content a:hover {
  text-decoration: underline;
}
</style> 