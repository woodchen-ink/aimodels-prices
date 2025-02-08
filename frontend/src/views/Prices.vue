<template>
  <div class="prices">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span>价格列表</span>
          </div>
          <div class="header-buttons">
            <template v-if="isAdmin && selectedPrices.length > 0">
              <el-button type="success" @click="batchUpdateStatus('approved')">批量通过</el-button>
              <el-button type="danger" @click="batchUpdateStatus('rejected')">批量拒绝</el-button>
              <el-divider direction="vertical" />
            </template>
            <el-button type="primary" @click="handleBatchAdd">批量添加</el-button>
            <el-button type="primary" @click="handleAdd">提交价格</el-button>
          </div>
        </div>
      </template>

      <div class="filter-section">
        <div class="filter-label">厂商筛选:</div>
        <div class="provider-filters">
          <el-button 
            :type="!selectedProvider ? 'primary' : ''" 
            @click="selectedProvider = ''"
          >全部</el-button>
          <el-button
            v-for="provider in providers"
            :key="provider.id"
            :type="selectedProvider === provider.id.toString() ? 'primary' : ''"
            @click="selectedProvider = provider.id.toString()"
          >
            <div style="display: flex; align-items: center; gap: 8px">
              <el-image
                v-if="provider.icon"
                :src="provider.icon"
                style="width: 16px; height: 16px"
              />
              <span>{{ provider.name }}</span>
            </div>
          </el-button>
        </div>
      </div>

      <el-table 
        :data="filteredPrices" 
        style="width: 100%"
        @selection-change="handlePriceSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column label="模型">
          <template #default="{ row }">
            <div class="value-container">
              <span>{{ row.model }}</span>
              <el-tag v-if="row.temp_model" type="warning" size="small" effect="light">
                待审核: {{ row.temp_model }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="计费类型" width="120">
          <template #default="{ row }">
            <div class="value-container">
              <span>{{ getBillingType(row.billing_type) }}</span>
              <el-tag v-if="row.temp_billing_type" type="warning" size="small" effect="light">
                待审核: {{ getBillingType(row.temp_billing_type) }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="模型厂商" width="180">
          <template #default="{ row }">
            <div class="value-container">
              <div style="display: flex; align-items: center; gap: 8px">
                <el-image 
                  v-if="getProvider(row.channel_type)?.icon"
                  :src="getProvider(row.channel_type)?.icon"
                  style="width: 24px; height: 24px"
                />
                <span>{{ getProvider(row.channel_type)?.name || row.channel_type }}</span>
              </div>
              <el-tag v-if="row.temp_channel_type" type="warning" size="small" effect="light">
                待审核: {{ getProvider(row.temp_channel_type)?.name || row.temp_channel_type }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="货币" width="80">
          <template #default="{ row }">
            <div class="value-container">
              <span>{{ row.currency }}</span>
              <el-tag v-if="row.temp_currency" type="warning" size="small" effect="light">
                待审核: {{ row.temp_currency }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="输入价格(M)" width="120">
          <template #default="{ row }">
            <div class="value-container">
              <span>{{ row.input_price === 0 ? '免费' : row.input_price }}</span>
              <el-tag v-if="row.temp_input_price !== null && row.temp_input_price !== undefined" type="warning" size="small" effect="light">
                待审核: {{ row.temp_input_price === 0 ? '免费' : row.temp_input_price }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="输出价格(M)" width="120">
          <template #default="{ row }">
            <div class="value-container">
              <span>{{ row.output_price === 0 ? '免费' : row.output_price }}</span>
              <el-tag v-if="row.temp_output_price !== null && row.temp_output_price !== undefined" type="warning" size="small" effect="light">
                待审核: {{ row.temp_output_price === 0 ? '免费' : row.temp_output_price }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="输入倍率" width="120">
          <template #default="{ row }">
            {{ row.input_price === 0 ? '免费' : calculateRate(row.input_price, row.currency) }}
          </template>
        </el-table-column>
        <el-table-column label="输出倍率" width="120">
          <template #default="{ row }">
            {{ row.output_price === 0 ? '免费' : calculateRate(row.output_price, row.currency) }}
          </template>
        </el-table-column>
        <el-table-column width="80">
          <template #default="{ row }">
            <el-popover
              placement="left"
              :width="200"
              trigger="hover"
            >
              <template #reference>
                <el-button link type="primary">详情</el-button>
              </template>
              <div class="price-detail">
                <div class="detail-item">
                  <span class="detail-label">创建者:</span>
                  <span class="detail-value">{{ row.created_by }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">价格来源:</span>
                  <div class="detail-value">
                    <span>{{ row.price_source }}</span>
                    <el-tag v-if="row.temp_price_source !== null && row.temp_price_source !== undefined" type="warning" size="small" effect="light">
                      待审核: {{ row.temp_price_source }}
                    </el-tag>
                  </div>
                </div>
                <div class="detail-item">
                  <span class="detail-label">状态:</span>
                  <span class="detail-value">{{ getStatus(row.status) }}</span>
                </div>
              </div>
            </el-popover>
          </template>
        </el-table-column>
        <el-table-column v-if="isAdmin" label="操作">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">删除</el-button>
              <el-button type="success" size="small" @click="updateStatus(row.id, 'approved')" :disabled="row.status !== 'pending'">通过</el-button>
              <el-button type="danger" size="small" @click="updateStatus(row.id, 'rejected')" :disabled="row.status !== 'pending'">拒绝</el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 批量添加对话框 -->
    <el-dialog v-model="batchDialogVisible" title="批量添加模型价格" width="90%">
      <div class="batch-add-container">
        <div class="batch-toolbar">
          <el-button type="primary" @click="addRow">添加行</el-button>
          <el-button type="danger" @click="removeSelectedRows" :disabled="!selectedRows.length">删除选中行</el-button>
          <el-divider direction="vertical" />
          <el-popover
            placement="bottom"
            :width="400"
            trigger="click"
          >
            <template #reference>
              <el-button type="success">从表格导入</el-button>
            </template>
            <div class="import-popover">
              <p class="import-tip">请粘贴表格数据（支持从Excel复制），每行格式为：</p>
              <p class="import-format">模型名称 计费类型 厂商 货币 输入价格 输出价格</p>
              <el-input
                v-model="importText"
                type="textarea"
                :rows="8"
                placeholder="例如：
dall-e-2 按Token收费 OpenAI 美元 16.000000 16.000000
dall-e-3 按Token收费 OpenAI 美元 40.000000 40.000000"
              />
              <div class="import-actions">
                <el-button type="primary" @click="handleImport">导入</el-button>
              </div>
            </div>
          </el-popover>
        </div>
        <el-table
          :data="batchForms"
          style="width: 100%"
          @selection-change="handleSelectionChange"
          height="400"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column label="模型" width="200">
            <template #default="{ row }">
              <el-input v-model="row.model" placeholder="请输入模型名称" />
            </template>
          </el-table-column>
          <el-table-column label="计费类型" width="150">
            <template #default="{ row }">
              <el-select v-model="row.billing_type" placeholder="请选择">
                <el-option label="按量计费" value="tokens" />
                <el-option label="按次计费" value="times" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="模型厂商" width="200">
            <template #default="{ row }">
              <el-select v-model="row.channel_type" placeholder="请选择">
                <el-option
                  v-for="provider in providers"
                  :key="provider.id"
                  :label="provider.name"
                  :value="provider.id.toString()"
                >
                  <div style="display: flex; align-items: center; gap: 8px">
                    <el-image
                      v-if="provider.icon"
                      :src="provider.icon"
                      style="width: 24px; height: 24px"
                    />
                    <span>{{ provider.name }}</span>
                  </div>
                </el-option>
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="货币" width="120">
            <template #default="{ row }">
              <el-select v-model="row.currency" placeholder="请选择">
                <el-option label="美元" value="USD" />
                <el-option label="人民币" value="CNY" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="输入价格(M)" width="150">
            <template #default="{ row }">
              <el-input-number v-model="row.input_price" :precision="4" :step="0.0001" />
            </template>
          </el-table-column>
          <el-table-column label="输出价格(M)" width="150">
            <template #default="{ row }">
              <el-input-number v-model="row.output_price" :precision="4" :step="0.0001" />
            </template>
          </el-table-column>
          <el-table-column label="价格来源">
            <template #default="{ row }">
              <el-input v-model="row.price_source" placeholder="请输入价格来源" />
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="batchDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitBatchForms" :loading="batchSubmitting">
            {{ batchSubmitting ? '提交中...' : '确定' }}
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 现有的单个添加对话框 -->
    <el-dialog v-model="dialogVisible" title="提交价格">
      <el-form :model="form" label-width="120px">
        <el-form-item label="模型">
          <el-input v-model="form.model" />
        </el-form-item>
        <el-form-item label="计费类型">
          <el-select v-model="form.billing_type" placeholder="请选择">
            <el-option label="按量计费" value="tokens" />
            <el-option label="按次计费" value="times" />
          </el-select>
        </el-form-item>
        <el-form-item label="模型厂商">
          <el-select v-model="form.channel_type" placeholder="请选择">
            <el-option 
              v-for="provider in providers" 
              :key="provider.id" 
              :label="provider.name"
              :value="provider.id.toString()"
            >
              <div style="display: flex; align-items: center; gap: 8px">
                <el-image 
                  v-if="provider.icon"
                  :src="provider.icon"
                  style="width: 24px; height: 24px"
                />
                <span>{{ provider.name }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="货币">
          <el-select v-model="form.currency" placeholder="请选择">
            <el-option label="美元" value="USD" />
            <el-option label="人民币" value="CNY" />
          </el-select>
        </el-form-item>
        <el-form-item label="输入价格(M)">
          <el-input-number v-model="form.input_price" :precision="4" :step="0.0001" />
        </el-form-item>
        <el-form-item label="输出价格(M)">
          <el-input-number v-model="form.output_price" :precision="4" :step="0.0001" />
        </el-form-item>
        <el-form-item label="价格来源">
          <el-input v-model="form.price_source" />
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
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'

const props = defineProps({
  user: Object
})

const prices = ref([])
const dialogVisible = ref(false)
const form = ref({
  model: '',
  billing_type: 'tokens',
  channel_type: '',
  currency: 'USD',
  input_price: 0,
  output_price: 0,
  price_source: '',
  created_by: ''
})
const router = useRouter()
const selectedProvider = ref('')

const isAdmin = computed(() => props.user?.role === 'admin')

const providers = ref([])
const getProvider = (id) => providers.value.find(p => p.id.toString() === id)

const statusMap = {
  'pending': '待审核',
  'approved': '已通过',
  'rejected': '已拒绝'
}

const billingTypeMap = {
  'tokens': '按量计费',
  'times': '按次计费'
}

const getStatus = (status) => statusMap[status] || status
const getBillingType = (type) => billingTypeMap[type] || type

const calculateRate = (price, currency) => {
  if (!price) return 0
  return currency === 'USD' ? (price / 2).toFixed(4) : (price / 14).toFixed(4)
}

const filteredPrices = computed(() => {
  if (!selectedProvider.value) return prices.value
  return prices.value.filter(p => p.channel_type === selectedProvider.value)
})

const editingPrice = ref(null)

const loadPrices = async () => {
  try {
    const { data } = await axios.get('/api/prices')
    prices.value = data
  } catch (error) {
    console.error('Failed to load prices:', error)
    ElMessage.error('加载数据失败')
  }
}

const handleEdit = (price) => {
  editingPrice.value = price
  form.value = { ...price }
  dialogVisible.value = true
}

const handleDelete = (price) => {
  ElMessageBox.confirm(
    '确定要删除这个价格吗？',
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await axios.delete(`/api/prices/${price.id}`)
      await loadPrices()
      ElMessage.success('删除成功')
    } catch (error) {
      console.error('Failed to delete price:', error)
      if (error.response?.data?.error) {
        ElMessage.error(error.response.data.error)
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
  editingPrice.value = null
  form.value = {
    model: '',
    billing_type: 'tokens',
    channel_type: '',
    currency: 'USD',
    input_price: 0,
    output_price: 0,
    price_source: '',
    created_by: ''
  }
  dialogVisible.value = true
}

const submitForm = async () => {
  try {
    form.value.created_by = props.user.username
    let response
    if (editingPrice.value) {
      // 更新已存在的价格
      response = await axios.put(`/api/prices/${editingPrice.value.id}`, form.value)
    } else {
      // 检查模型是否已存在
      const existingPrice = prices.value?.find(p => 
        p.model === form.value.model && 
        p.channel_type === form.value.channel_type
      )
      if (existingPrice) {
        ElMessageBox.confirm(
          '该模型价格已存在，是否要更新？',
          '提示',
          {
            confirmButtonText: '更新',
            cancelButtonText: '取消',
            type: 'warning',
          }
        ).then(async () => {
          response = await axios.put(`/api/prices/${existingPrice.id}`, form.value)
          handleSubmitResponse(response)
        }).catch(() => {
          // 用户取消更新
        })
        return
      }
      // 创建新价格
      response = await axios.post('/api/prices', form.value)
    }
    handleSubmitResponse(response)
  } catch (error) {
    console.error('Failed to submit price:', error)
    if (error.response?.data?.error) {
      ElMessage.error(error.response.data.error)
    } else {
      ElMessage.error('操作失败')
    }
  }
}

const handleSubmitResponse = async (response) => {
  const { data } = response
  if (data.error) {
    ElMessage.error(data.error)
    return
  }
  await loadPrices()
  dialogVisible.value = false
  ElMessage.success(editingPrice.value ? '更新成功' : '添加成功')
  editingPrice.value = null
  form.value = {
    model: '',
    billing_type: 'tokens',
    channel_type: '',
    currency: 'USD',
    input_price: 0,
    output_price: 0,
    price_source: '',
    created_by: ''
  }
}

const updateStatus = async (id, status) => {
  try {
    const { data } = await axios.put(`/api/prices/${id}/status`, { status })
    await loadPrices()
    ElMessage.success(data.message || '更新成功')
  } catch (error) {
    console.error('Failed to update status:', error)
    if (error.response?.data?.error) {
      ElMessage.error(error.response.data.error)
    } else if (error.response?.status === 401) {
      ElMessage.error('请先登录')
      router.push('/login')
    } else if (error.response?.status === 403) {
      ElMessage.error('需要管理员权限')
    } else {
      ElMessage.error('更新失败')
    }
  }
}

// 批量添加相关的状态
const batchDialogVisible = ref(false)
const batchForms = ref([])
const selectedRows = ref([])
const batchSubmitting = ref(false)

// 创建新行的默认数据
const createNewRow = () => ({
  model: '',
  billing_type: 'tokens',
  channel_type: '',
  currency: 'USD',
  input_price: 0,
  output_price: 0,
  price_source: '',
  created_by: props.user?.username || ''
})

// 添加新行
const addRow = () => {
  batchForms.value.push(createNewRow())
}

// 处理选择变化
const handleSelectionChange = (rows) => {
  selectedRows.value = rows
}

// 删除选中的行
const removeSelectedRows = () => {
  const selectedIds = new Set(selectedRows.value.map(row => batchForms.value.indexOf(row)))
  batchForms.value = batchForms.value.filter((_, index) => !selectedIds.has(index))
  selectedRows.value = []
}

// 打开批量添加对话框
const handleBatchAdd = () => {
  if (!props.user) {
    router.push('/login')
    ElMessage.warning('请先登录')
    return
  }
  batchForms.value = [createNewRow()]
  batchDialogVisible.value = true
}

// 提交批量表单
const submitBatchForms = async () => {
  if (!batchForms.value.length) {
    ElMessage.warning('请至少添加一条数据')
    return
  }

  // 验证数据
  const invalidForms = batchForms.value.filter(form => 
    !form.model || !form.channel_type || !form.price_source
  )
  
  if (invalidForms.length) {
    ElMessage.error('请填写完整所有必填字段')
    return
  }

  batchSubmitting.value = true
  try {
    // 逐个提交数据
    for (const form of batchForms.value) {
      await axios.post('/api/prices', form)
    }
    
    await loadPrices()
    batchDialogVisible.value = false
    ElMessage.success('批量添加成功')
  } catch (error) {
    console.error('Failed to submit batch prices:', error)
    if (error.response?.data?.error) {
      ElMessage.error(error.response.data.error)
    } else {
      ElMessage.error('批量添加失败')
    }
  } finally {
    batchSubmitting.value = false
  }
}

// 添加导入相关的状态
const importText = ref('')

// 处理导入
const handleImport = () => {
  if (!importText.value.trim()) {
    ElMessage.warning('请先粘贴数据')
    return
  }

  const lines = importText.value.trim().split('\n')
  const newRows = lines.map(line => {
    // 使用正则表达式匹配制表符或多个空格作为分隔符
    const parts = line.trim().split(/\t+|\s{2,}/)
    if (!parts || parts.length < 6) {
      ElMessage.warning(`行格式不正确：${line}`)
      return null
    }

    const [model, billingType, providerName, currency, inputPrice, outputPrice] = parts
    
    // 查找模型厂商ID
    const provider = providers.value.find(p => p.name === providerName)
    if (!provider) {
      ElMessage.warning(`未找到模型厂商：${providerName}`)
      return null
    }

    // 处理计费类型
    let billing_type = 'tokens'
    if (billingType.includes('Token')) {
      billing_type = 'tokens'
    } else if (billingType.includes('次')) {
      billing_type = 'times'
    }

    // 处理货币
    let currencyCode = 'USD'
    if (currency.includes('美元')) {
      currencyCode = 'USD'
    } else if (currency.includes('人民币') || currency.includes('CNY')) {
      currencyCode = 'CNY'
    }

    return {
      model,
      billing_type,
      channel_type: provider.id.toString(),
      currency: currencyCode,
      input_price: parseFloat(inputPrice),
      output_price: parseFloat(outputPrice),
      price_source: '官方',
      created_by: props.user?.username || ''
    }
  }).filter(row => row !== null)

  if (newRows.length > 0) {
    batchForms.value = [...batchForms.value, ...newRows]
    importText.value = ''
    ElMessage.success(`成功导入 ${newRows.length} 条数据`)
  }
}

const selectedPrices = ref([])

const handlePriceSelectionChange = (selection) => {
  selectedPrices.value = selection
}

const batchUpdateStatus = async (status) => {
  if (!selectedPrices.value.length) {
    ElMessage.warning('请先选择要审核的价格')
    return
  }

  // 过滤出待审核的价格
  const pendingPrices = selectedPrices.value.filter(price => price.status === 'pending')
  if (!pendingPrices.length) {
    ElMessage.warning('选中的价格中没有待审核的项目')
    return
  }

  try {
    // 确认操作
    await ElMessageBox.confirm(
      `确定要${status === 'approved' ? '通过' : '拒绝'}选中的 ${pendingPrices.length} 条待审核价格吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: status === 'approved' ? 'success' : 'warning'
      }
    )

    // 批量更新状态
    for (const price of pendingPrices) {
      await axios.put(`/api/prices/${price.id}/status`, { status })
    }

    await loadPrices()
    ElMessage.success('批量审核成功')
  } catch (error) {
    if (error === 'cancel') return
    console.error('Failed to batch update status:', error)
    ElMessage.error('批量审核失败')
  }
}

onMounted(async () => {
  await loadPrices()
  try {
    const { data } = await axios.get('/api/providers')
    providers.value = data
  } catch (error) {
    console.error('Failed to load providers:', error)
    ElMessage.error('加载模型厂商数据失败')
  }
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
}

.filter-section {
  margin: 16px 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-label {
  font-size: 14px;
  color: #606266;
}

.provider-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

:deep(.el-button) {
  margin: 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
}

:deep(.el-dialog__body) {
  padding-right: 20px;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

:deep(.el-dialog) {
  margin: 0 !important;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.prices {
  padding-right: 0;
}

.value-container {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.value-container :deep(.el-tag) {
  margin: 0;
  width: fit-content;
}

.value-container span {
  word-break: break-all;
}

.header-buttons {
  display: flex;
  gap: 12px;
}

.batch-add-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.batch-toolbar {
  display: flex;
  gap: 12px;
  padding: 8px 0;
}

:deep(.el-input-number) {
  width: 100%;
}

:deep(.el-select) {
  width: 100%;
}

.import-popover {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.import-tip {
  margin: 0;
  color: #606266;
  font-size: 14px;
}

.import-format {
  margin: 0;
  color: #409EFF;
  font-size: 13px;
  background-color: #ecf5ff;
  padding: 8px;
  border-radius: 4px;
}

.import-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

.price-detail {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-label {
  color: #909399;
  font-size: 13px;
}

.detail-value {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.creator-name {
  display: inline-block;
  width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style> 