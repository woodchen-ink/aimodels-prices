<template>
  <div class="login">
    <el-card class="login-card">
      <template #header>
        <h2>登录</h2>
      </template>
      <el-button type="primary" @click="handleLogin" :loading="loading">
        {{ loading ? '登录中...' : '登录' }}
      </el-button>
    </el-card>
  </div>
</template>

<script setup>
import { ref, inject } from 'vue'
import axios from 'axios'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const updateGlobalUser = inject('updateGlobalUser')

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
      router.push('/prices')
    }
  } catch (error) {
    console.error('Failed to login:', error)
    ElMessage.error('登录失败')
    loading.value = false
  }
}
</script>

<style scoped>
.login {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: calc(100vh - 60px);
}

.login-card {
  width: 100%;
  max-width: 400px;
}

h2 {
  text-align: center;
  margin: 0;
}

.el-button {
  width: 100%;
}
</style> 