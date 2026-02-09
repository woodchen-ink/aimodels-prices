<template>
  <el-alert
    v-if="!dismissed"
    title="社区投票进行中"
    type="info"
    :closable="true"
    @close="handleClose"
    class="vote-banner"
  >
    当前正在对是否保留"模型类别"进行投票，如有意向可
    <a href="https://www.sunai.net/t/topic/1084" target="_blank" class="vote-link">
      前往参与投票 →
    </a>
  </el-alert>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const STORAGE_KEY = 'voteBannerDismissed'
const dismissed = ref(false)

onMounted(() => {
  dismissed.value = localStorage.getItem(STORAGE_KEY) === 'true'
})

function handleClose() {
  dismissed.value = true
  localStorage.setItem(STORAGE_KEY, 'true')
}
</script>

<style scoped>
.vote-banner {
  margin-bottom: 16px;
}

.vote-link {
  color: var(--color-primary);
  text-decoration: none;
  font-weight: 500;
}

.vote-link:hover {
  text-decoration: underline;
}
</style>
