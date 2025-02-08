<template>
  <div class="providers">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>模型厂商</span>
          <el-button type="primary" @click="handleAdd">添加模型厂商</el-button>
        </div>
      </template>

      <el-table :data="sortedProviders" style="width: 100%">
        <el-table-column prop="id" label="ID" />
        <el-table-column label="名称">
          <template #default="{ row }">
            {{ row.name }}
          </template>
        </el-table-column>
        <el-table-column label="图标">
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
        <el-table-column prop="created_by" label="创建者" />
        <el-table-column v-if="isAdmin" label="操作" width="200">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="dialogTitle">
      <el-form :model="form" label-width="80px">
        <el-form-item label="ID" v-if="!editingProvider">
          <el-input-number v-model="form.id" :min="1" />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="图标链接">
          <el-input v-model="form.icon" />
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

const props = defineProps({
  user: Object
})

const providers = ref([])
const dialogVisible = ref(false)
const editingProvider = ref(null)
const form = ref({
  id: 1,
  name: '',
  icon: ''
})
const router = useRouter()

const isAdmin = computed(() => props.user?.role === 'admin')

// 按ID排序的模型厂商
const sortedProviders = computed(() => {
  return [...providers.value].sort((a, b) => a.id - b.id)
})

const loadProviders = async () => {
  try {
    const { data } = await axios.get('/api/providers')
    providers.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error('Failed to load providers:', error)
    ElMessage.error('加载数据失败')
  }
}

onMounted(() => {
  loadProviders()
})

const dialogTitle = computed(() => {
  if (editingProvider.value) {
    return '编辑模型厂商'
  }
  return '添加模型厂商'
})

const handleEdit = (provider) => {
  if (!isAdmin.value) {
    ElMessage.warning('只有管理员可以编辑模型厂商信息')
    return
  }
  editingProvider.value = provider
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
  dialogVisible.value = true
  form.value = { id: 1, name: '', icon: '' }
}

const submitForm = async () => {
  try {
    if (editingProvider.value) {
      if (!isAdmin.value) {
        ElMessage.error('只有管理员可以编辑模型厂商信息')
        return
      }
      // 管理员编辑模型厂商
      const { data } = await axios.put(`/api/providers/${editingProvider.value.id}`, form.value)
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
      const { data } = await axios.post('/api/providers', form.value)
      if (data.error) {
        ElMessage.error(data.error)
        return
      }
      providers.value = providers.value ? [...providers.value, data] : [data]
      ElMessage.success('添加成功')
    }
    dialogVisible.value = false
    editingProvider.value = null
    form.value = { id: 1, name: '', icon: '' }
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
</style> 