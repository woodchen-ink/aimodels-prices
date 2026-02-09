<template>
  <div class="providers-container">
    <el-card v-loading="loading" element-loading-text="加载中...">
      <template #header>
        <div class="card-header">
          <span class="title">模型厂商</span>
          <el-button type="primary" @click="handleAdd">添加模型厂商</el-button>
        </div>
      </template>

      <el-table :data="sortedProviders" style="width: 100%" v-loading="tableLoading" element-loading-text="加载中..." class="providers-table" :header-cell-style="{ background: 'var(--color-bg-light)', color: 'var(--color-text-secondary)' }">
        <el-table-column prop="id" label="ID" min-width="200"/>
        <el-table-column label="名称" min-width="200">
          <template #default="{ row }">
            {{ row.name }}
          </template>
        </el-table-column>
        <el-table-column label="图标" min-width="200">
          <template #default="{ row }">
            <el-image 
              v-if="row.icon"
              :src="row.icon" 
              style="width: 24px; height: 24px"
              :preview-src-list="[row.icon]"
            />
            <span v-else>-</span>
          </template>
        </el-table-column>
        <!-- <el-table-column prop="created_by" label="创建者" min-width="100"/> -->
        <el-table-column v-if="isAdmin" label="操作" min-width="150">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingProvider ? '编辑模型厂商' : '添加模型厂商'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="ID">
          <el-input v-model.number="form.id" placeholder="请输入厂商ID" type="number" />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="请输入厂商名称" />
        </el-form-item>
        <el-form-item label="图标">
          <el-input v-model="form.icon" placeholder="请输入图标URL" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitForm">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { isAdmin as checkIsAdmin } from '@/utils/permission'

const props = defineProps({
  user: Object
})

const providers = ref([])
const dialogVisible = ref(false)
const editingProvider = ref(null)
const form = ref({
  id: '',
  name: '',
  icon: ''
})
const router = useRouter()

const isAdmin = computed(() => checkIsAdmin(props.user))

// 添加加载状态变量
const loading = ref(true)
const tableLoading = ref(true)

// 按ID排序的模型厂商
const sortedProviders = computed(() => {
  return [...providers.value].sort((a, b) => a.id - b.id)
})

const loadProviders = async () => {
  tableLoading.value = true
  try {
    const { data } = await axios.get('/api/providers')
    providers.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error('Failed to load providers:', error)
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
    tableLoading.value = false
  }
}

onMounted(() => {
  loadProviders()
})

const handleEdit = (provider) => {
  editingProvider.value = provider
  // 编辑时复制所有字段
  form.value = { ...provider }
  dialogVisible.value = true
}

const handleDelete = (provider) => {
  if (!isAdmin.value) {
    ElMessage.warning('只有管理员可以删除模型厂商')
    return
  }
  
  ElMessageBox.confirm(
    '确定要删除这个模型厂商吗？',
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await axios.delete(`/api/providers/${provider.id}`)
      providers.value = providers.value.filter(p => p.id !== provider.id)
      ElMessage.success('删除成功')
    } catch (error) {
      console.error('Failed to delete provider:', error)
      if (error.response?.status === 403) {
        ElMessage.error('没有权限执行此操作')
      } else {
        ElMessage.error('删除失败')
      }
    }
  })
}

const handleAdd = () => {
  if (!props.user) {
    router.push('/login')
    ElMessage.warning('请先登录')
    return
  }
  editingProvider.value = null
  // 重置表单时确保所有字段都被重置
  form.value = {
    id: '',
    name: '',
    icon: ''
  }
  dialogVisible.value = true
}

const submitForm = async () => {
  try {
    // 确保ID是数字类型
    const formData = {
      ...form.value,
      id: parseInt(form.value.id)
    }

    if (editingProvider.value) {
      if (!isAdmin.value) {
        ElMessage.error('只有管理员可以编辑模型厂商信息')
        return
      }
      // 管理员编辑模型厂商
      const { data } = await axios.put(`/api/providers/${editingProvider.value.id}`, formData)
      if (data.error) {
        ElMessage.error(data.error)
        return
      }
      const index = providers.value.findIndex(p => p.id === editingProvider.value.id)
      if (index !== -1) {
        providers.value[index] = data
      }
      ElMessage.success('更新成功')
    } else {
      // 创建新模型厂商
      const { data } = await axios.post('/api/providers', formData)
      if (data.error) {
        ElMessage.error(data.error)
        return
      }
      providers.value = providers.value ? [...providers.value, data] : [data]
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    editingProvider.value = null
    form.value = { id: '', name: '', icon: '' }
  } catch (error) {
    console.error('Failed to submit provider:', error)
    if (error.response?.data?.error) {
      ElMessage.error(error.response.data.error)
    } else if (error.response?.status === 403) {
      ElMessage.error('没有权限执行此操作')
    } else {
      ElMessage.error('操作失败')
    }
  }
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

:deep(.el-loading-spinner) {
  .el-loading-text {
    color: var(--color-primary);
  }
  .path {
    stroke: var(--color-primary);
  }
}
.providers-container {
  padding: 20px;
  display: flex;
  justify-content: center;
}
.providers-card {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  box-shadow: 0 2px 12px 0 rgba(20, 20, 19, 0.06);
}
.title {
  font-size: 18px;
  font-weight: bold;
  color: var(--color-text-primary);
}
.providers-table {
  width: 100%;
  margin-top: 10px;
}

/* 响应式布局 */
@media screen and (max-width: 768px) {
  .providers-container {
    padding: 10px;
  }
  .providers-card {
    max-width: 100%;
  }
}
</style>