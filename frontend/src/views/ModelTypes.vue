<template>
  <div class="model-types-container">
    <el-card v-loading="loading" element-loading-text="加载中..." class="model-types-card">
      <template #header>
        <div class="card-header">
          <span class="title">模型类别</span>
          <el-button v-if="isAdmin" type="primary" @click="handleAdd">添加模型类别</el-button>
        </div>
      </template>

      <el-table 
        :data="modelTypes" 
        v-loading="tableLoading" 
        element-loading-text="加载中..."
        row-key="key"
        class="model-types-table"
        :header-cell-style="{ background: '#f5f7fa', color: '#606266' }"
      >
        <el-table-column prop="key" label="类别键值" min-width="200" />
        <el-table-column prop="label" label="类别名称" min-width="200" />
        <el-table-column prop="sort_order" label="排序" min-width="120" align="center" />
        <el-table-column v-if="isAdmin" label="操作" min-width="150" align="center">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="editingType ? '编辑模型类别' : '添加模型类别'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="类别键值">
          <el-input v-model="form.key" placeholder="请输入类别键值" :disabled="!!editingType" />
        </el-form-item>
        <el-form-item label="类别名称">
          <el-input v-model="form.label" placeholder="请输入类别名称" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" :step="1" />
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

const props = defineProps({
  user: Object
})

const modelTypes = ref([])
const dialogVisible = ref(false)
const editingType = ref(null)
const form = ref({
  key: '',
  label: '',
  sort_order: 0
})
const loading = ref(true)
const tableLoading = ref(false)

const isAdmin = computed(() => props.user?.role === 'admin')

// 获取所有模型类别
const fetchModelTypes = async () => {
  try {
    tableLoading.value = true
    const response = await axios.get('/api/model-types')
    modelTypes.value = response.data
  } catch (error) {
    ElMessage.error('获取模型类别失败')
    console.error(error)
  } finally {
    loading.value = false
    tableLoading.value = false
  }
}

// 添加模型类别
const handleAdd = () => {
  editingType.value = null
  form.value = {
    key: '',
    label: '',
    sort_order: 0
  }
  dialogVisible.value = true
}

// 编辑模型类别
const handleEdit = (row) => {
  editingType.value = row
  form.value = {
    key: row.key,
    label: row.label,
    sort_order: row.sort_order
  }
  dialogVisible.value = true
}

// 删除模型类别
const handleDelete = (row) => {
  ElMessageBox.confirm(
    '确定要删除此模型类别吗？如果有价格记录使用此类别，将无法删除。',
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(async () => {
      try {
        await axios.delete(`/api/model-types/${row.key}`)
        ElMessage.success('删除成功')
        fetchModelTypes()
      } catch (error) {
        ElMessage.error(error.response?.data?.error || '删除失败')
      }
    })
    .catch(() => {
      // 用户取消删除
    })
}

// 提交表单
const submitForm = async () => {
  try {
    if (editingType.value) {
      // 更新
      await axios.put(`/api/model-types/${editingType.value.key}`, {
        key: form.value.key,
        label: form.value.label,
        sort_order: form.value.sort_order
      })
      ElMessage.success('更新成功')
    } else {
      // 创建
      await axios.post('/api/model-types', {
        key: form.value.key,
        label: form.value.label,
        sort_order: form.value.sort_order
      })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    fetchModelTypes()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '操作失败')
  }
}

onMounted(() => {
  fetchModelTypes()
})
</script>

<style scoped>
.model-types-container {
  padding: 20px;
  display: flex;
  justify-content: center;
}

.model-types-card {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.title {
  font-size: 18px;
  font-weight: bold;
  color: #303133;
}

.model-types-table {
  width: 100%;
  margin-top: 10px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
}

/* 响应式布局 */
@media screen and (max-width: 768px) {
  .model-types-container {
    padding: 10px;
  }
  
  .model-types-card {
    max-width: 100%;
  }
}
</style> 