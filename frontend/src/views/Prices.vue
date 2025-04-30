<template>
  <div class="prices">
    <el-card v-loading="loading" element-loading-text="加载中...">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span>价格列表</span>
          </div>
          <div class="header-buttons">
            <template v-if="isAdmin && selectedPrices.length > 0">
              <el-button type="success" @click="batchUpdateStatus('approved')">批量通过</el-button>
              <el-button type="danger" @click="batchUpdateStatus('rejected')">批量拒绝</el-button>
              <el-button type="danger" @click="batchDelete">批量删除</el-button>
              <el-divider direction="vertical" />
            </template>
            <template v-if="isAdmin">
              <el-button type="success" @click="approveAllPending">全部通过</el-button>
              <el-divider direction="vertical" />
            </template>
            <el-button type="primary" @click="handleBatchAdd">批量添加</el-button>
            <el-button type="primary" @click="handleAdd">提交价格</el-button>
          </div>
        </div>
      </template>

      <!-- 添加搜索框 -->
      <div class="filter-section">
        <div class="filter-label" style="min-width:80px;">搜索模型:</div>
        <div>
          <el-input v-model="searchQuery" placeholder="搜索模型名称" clearable prefix-icon="Search" @input="handleSearch">
            <template #prefix>
              <el-icon>
                <Search />
              </el-icon>
            </template>
          </el-input>
        </div>
      </div>

      <div class="filter-section">
        <div class="filter-label" style="min-width:80px;">厂商筛选:</div>
        <div class="provider-filters">
          <el-button :type="!selectedProvider ? 'primary' : ''" @click="selectedProvider = ''">全部</el-button>
          <el-button v-for="provider in providers" :key="provider.id"
            :type="selectedProvider === provider.id.toString() ? 'primary' : ''"
            @click="selectedProvider = provider.id.toString()">
            <div style="display: flex; align-items: center; gap: 8px">
              <el-image v-if="provider.icon" :src="provider.icon" style="width: 16px; height: 16px" />
              <span>{{ provider.name }}</span>
            </div>
          </el-button>
        </div>
      </div>

      <div class="filter-section">
        <div class="filter-label" style="min-width:80px;">模型类别:</div>
        <div class="model-type-filters">
          <el-button :type="!selectedModelType ? 'primary' : ''" @click="selectedModelType = ''">全部</el-button>
          <el-button v-for="(label, key) in modelTypeMap" :key="key" :type="selectedModelType === key ? 'primary' : ''"
            @click="selectedModelType = key">
            {{ label }}
          </el-button>
        </div>
      </div>

      <!-- 添加状态筛选 -->
      <div class="filter-section">
        <div class="filter-label" style="min-width:80px;">状态筛选:</div>
        <div class="status-filters">
          <el-button :type="!selectedStatus ? 'primary' : ''" @click="selectedStatus = ''">全部</el-button>
          <el-button v-for="(status, key) in statusMap" :key="key" 
            :type="selectedStatus === key ? 'primary' : ''"
            @click="selectedStatus = key">
            {{ status }}
          </el-button>
        </div>
      </div>

      <!-- 替换表格为卡片布局 -->
      <div class="price-cards-container">
        <template v-if="loading">
          <div v-for="i in 6" :key="i" class="price-card skeleton">
            <el-skeleton :rows="4" animated />
          </div>
        </template>
        <template v-else>
          <div v-for="price in prices" :key="price.id" class="price-card">
            <div class="price-card-header">
              <div class="provider-info">
                <el-image 
                  v-if="getProvider(price.channel_type)?.icon" 
                  :src="getProvider(price.channel_type)?.icon"
                  class="provider-icon" 
                />
                <span class="provider-name">{{ getProvider(price.channel_type)?.name }}</span>
              </div>
              <div class="model-status" :class="price.status">
                {{ getStatus(price.status) }}
              </div>
            </div>

            <div class="model-info">
              <h3 class="model-name">
                {{ price.model }}
                <el-tag v-if="price.temp_model && price.temp_model !== 'NULL'" 
                  type="warning" size="small" effect="light">
                  待审核: {{ price.temp_model }}
                </el-tag>
              </h3>
              <div class="model-meta">
                <el-tag size="small" effect="plain">{{ getModelType(price.model_type) }}</el-tag>
                <el-tag size="small" effect="plain">{{ getBillingType(price.billing_type) }}</el-tag>
                <el-tag size="small" effect="plain">{{ price.currency }}</el-tag>
              </div>
            </div>

            <div class="price-info">
              <div class="price-row">
                <span class="price-label">输入价格(M):</span>
                <span class="price-value">{{ price.input_price === 0 ? '免费' : price.input_price }}</span>
                <el-tag v-if="price.temp_input_price !== null && price.temp_input_price !== undefined" 
                  type="warning" size="small" effect="light">
                  待审核: {{ price.temp_input_price === 0 ? '免费' : price.temp_input_price }}
                </el-tag>
              </div>
              <div class="price-row">
                <span class="price-label">输出价格(M):</span>
                <span class="price-value">{{ price.output_price === 0 ? '免费' : price.output_price }}</span>
                <el-tag v-if="price.temp_output_price !== null && price.temp_output_price !== undefined" 
                  type="warning" size="small" effect="light">
                  待审核: {{ price.temp_output_price === 0 ? '免费' : price.temp_output_price }}
                </el-tag>
              </div>
            </div>

            <div v-if="hasExtendedPrices(price)" class="extended-prices">
              <div class="section-title">扩展价格</div>
              <div class="extended-price-grid">
                <template v-if="price.input_audio_tokens">
                  <div class="extended-price-item">
                    <span class="ext-price-label">音频输入倍率</span>
                    <span class="ext-price-value">{{ price.input_audio_tokens }}</span>
                    <el-tag v-if="price.temp_input_audio_tokens" type="warning" size="small" effect="light">
                      待审核: {{ price.temp_input_audio_tokens }}
                    </el-tag>
                  </div>
                </template>
                <template v-if="price.cached_read_tokens">
                  <div class="extended-price-item">
                    <span class="ext-price-label">缓存读取倍率</span>
                    <span class="ext-price-value">{{ price.cached_read_tokens }}</span>
                    <el-tag v-if="price.temp_cached_read_tokens" type="warning" size="small" effect="light">
                      待审核: {{ price.temp_cached_read_tokens }}
                    </el-tag>
                  </div>
                </template>
                <template v-if="price.reasoning_tokens">
                  <div class="extended-price-item">
                    <span class="ext-price-label">推理倍率</span>
                    <span class="ext-price-value">{{ price.reasoning_tokens }}</span>
                    <el-tag v-if="price.temp_reasoning_tokens" type="warning" size="small" effect="light">
                      待审核: {{ price.temp_reasoning_tokens }}
                    </el-tag>
                  </div>
                </template>
                <template v-if="price.input_text_tokens">
                  <div class="extended-price-item">
                    <span class="ext-price-label">输入文本倍率</span>
                    <span class="ext-price-value">{{ price.input_text_tokens }}</span>
                    <el-tag v-if="price.temp_input_text_tokens" type="warning" size="small" effect="light">
                      待审核: {{ price.temp_input_text_tokens }}
                    </el-tag>
                  </div>
                </template>
                <template v-if="price.output_text_tokens">
                  <div class="extended-price-item">
                    <span class="ext-price-label">输出文本倍率</span>
                    <span class="ext-price-value">{{ price.output_text_tokens }}</span>
                    <el-tag v-if="price.temp_output_text_tokens" type="warning" size="small" effect="light">
                      待审核: {{ price.temp_output_text_tokens }}
                    </el-tag>
                  </div>
                </template>
                <template v-if="price.input_image_tokens">
                  <div class="extended-price-item">
                    <span class="ext-price-label">输入图片倍率</span>
                    <span class="ext-price-value">{{ price.input_image_tokens }}</span>
                    <el-tag v-if="price.temp_input_image_tokens" type="warning" size="small" effect="light">
                      待审核: {{ price.temp_input_image_tokens }}
                    </el-tag>
                  </div>
                </template>
                <template v-if="price.output_image_tokens">
                  <div class="extended-price-item">
                    <span class="ext-price-label">输出图片倍率</span>
                    <span class="ext-price-value">{{ price.output_image_tokens }}</span>
                    <el-tag v-if="price.temp_output_image_tokens" type="warning" size="small" effect="light">
                      待审核: {{ price.temp_output_image_tokens }}
                    </el-tag>
                  </div>
                </template>
              </div>
            </div>

            <div class="price-card-footer">
              <div class="meta-info">
                <span class="created-by">创建者: {{ price.created_by }}</span>
                <span class="created-at">{{ new Date(price.created_at).toLocaleString() }}</span>
              </div>
              <div class="action-buttons">
                <template v-if="isAdmin">
                  <el-tooltip content="编辑" placement="top">
                    <el-button type="primary" link @click="handleEdit(price)">
                      <el-icon><Edit /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-tooltip content="删除" placement="top">
                    <el-button type="danger" link @click="handleDelete(price)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-tooltip :content="price.status === 'pending' ? '通过审核' : '已审核'" placement="top">
                    <el-button type="success" link @click="updateStatus(price.id, 'approved')"
                      :disabled="price.status !== 'pending'">
                      <el-icon><Check /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <el-tooltip :content="price.status === 'pending' ? '拒绝审核' : '已审核'" placement="top">
                    <el-button type="danger" link @click="updateStatus(price.id, 'rejected')"
                      :disabled="price.status !== 'pending'">
                      <el-icon><Close /></el-icon>
                    </el-button>
                  </el-tooltip>
                </template>
                <template v-else>
                  <el-tooltip :content="price.status === 'pending' ? '等待审核中' : '提交修改'" placement="top">
                    <el-button type="primary" link @click="handleQuickEdit(price)" :disabled="price.status === 'pending'">
                      <el-icon><Edit /></el-icon>
                    </el-button>
                  </el-tooltip>
                </template>
              </div>
            </div>
          </div>
        </template>
      </div>

      <!-- 修改分页组件 -->
      <div class="pagination-container">
        <el-pagination v-model:current-page="currentPage" v-model:page-size="pageSize" :page-sizes="[10, 20, 50, 100]"
          :total="total" layout="total, sizes, prev, pager, next" :small="false" @size-change="handleSizeChange"
          @current-change="handleCurrentChange">
          <template #sizes>
            <el-select v-model="pageSize"
              :options="[10, 20, 50, 100].map(item => ({ value: item, label: `${item} 条/页` }))">
              <template #prefix>每页</template>
            </el-select>
          </template>
        </el-pagination>
      </div>
    </el-card>

    <!-- 批量添加对话框 -->
    <el-dialog v-model="batchDialogVisible" title="批量添加模型价格" width="1330px">
      <div class="batch-add-container">
        <div class="batch-toolbar">
          <el-button type="primary" @click="addRow">添加行</el-button>
          <el-divider direction="vertical" />
          <el-popover placement="bottom" :width="400" trigger="click">
            <template #reference>
              <el-button type="success">从表格导入</el-button>
            </template>
            <div class="import-popover">
              <p class="import-tip">请粘贴表格数据（支持从Excel复制），每行格式为：</p>
              <p class="import-format">模型名称 计费类型 厂商 货币 输入价格 输出价格</p>
              <el-input v-model="importText" type="textarea" :rows="8" placeholder="例如：
dall-e-2 按Token收费 OpenAI 美元 16.000000 16.000000
dall-e-3 按Token收费 OpenAI 美元 40.000000 40.000000" />
              <div class="import-actions">
                <el-button type="primary" @click="handleImport">导入</el-button>
              </div>
            </div>
          </el-popover>
        </div>

        <el-table :data="batchForms" style="width: 100%" height="400">
          <el-table-column label="操作" width="100">
            <template #default="{ row, $index }">
              <div class="row-actions">
                <el-tooltip content="复制" placement="top">
                  <el-button type="primary" link @click="duplicateRow($index)">
                    <el-icon>
                      <Document />
                    </el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top">
                  <el-button type="danger" link @click="removeRow($index)">
                    <el-icon>
                      <Delete />
                    </el-icon>
                  </el-button>
                </el-tooltip>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="模型" width="180">
            <template #default="{ row }">
              <el-input v-model="row.model" placeholder="请输入模型名称" />
            </template>
          </el-table-column>
          <el-table-column label="模型类型" width="120">
            <template #default="{ row }">
              <el-select v-model="row.model_type" placeholder="请选择或输入" allow-create filterable
                @create="handleModelTypeCreate">
                <el-option v-for="(label, value) in modelTypeMap" :key="value" :label="label" :value="value" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="计费类型" width="120">
            <template #default="{ row }">
              <el-select v-model="row.billing_type" placeholder="请选择">
                <el-option label="按量计费" value="tokens" />
                <el-option label="按次计费" value="times" />
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="模型厂商" width="180">
            <template #default="{ row }">
              <el-select v-model="row.channel_type" placeholder="请选择">
                <el-option v-for="provider in providers" :key="provider.id" :label="provider.name"
                  :value="provider.id.toString()">
                  <div style="display: flex; align-items: center; gap: 8px">
                    <el-image v-if="provider.icon" :src="provider.icon" style="width: 24px; height: 24px" />
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
              <el-input-number v-model="row.input_price" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入价格" />
            </template>
          </el-table-column>
          <el-table-column label="输出价格(M)" width="150">
            <template #default="{ row }">
              <el-input-number v-model="row.output_price" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入价格" />
            </template>
          </el-table-column>
          <el-table-column label="价格来源" min-width="200" width="200">
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
    <el-dialog v-model="dialogVisible" :title="editingPrice ? (isAdmin ? '编辑价格' : '提交价格修改') : '提交价格'" width="700px">
      <el-form :model="form" label-width="100px">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="模型">
              <el-input v-model="form.model" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="模型类型">
              <el-select v-model="form.model_type" placeholder="请选择或输入" allow-create filterable
                @create="handleModelTypeCreate">
                <el-option v-for="(label, value) in modelTypeMap" :key="value" :label="label" :value="value" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="计费类型">
              <el-select v-model="form.billing_type" placeholder="请选择">
                <el-option label="按量计费" value="tokens" />
                <el-option label="按次计费" value="times" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="模型厂商">
              <el-select v-model="form.channel_type" placeholder="请选择">
                <el-option v-for="provider in providers" :key="provider.id" :label="provider.name"
                  :value="provider.id.toString()">
                  <div style="display: flex; align-items: center; gap: 8px">
                    <el-image v-if="provider.icon" :src="provider.icon" style="width: 24px; height: 24px" />
                    <span>{{ provider.name }}</span>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="货币">
              <el-select v-model="form.currency" placeholder="请选择">
                <el-option label="美元" value="USD" />
                <el-option label="人民币" value="CNY" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="输入价格(M)">
              <el-input-number v-model="form.input_price" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入价格" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="输出价格(M)">
              <el-input-number v-model="form.output_price" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入价格" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-divider>扩展价格（可选）</el-divider>
          </el-col>
          <el-col :span="12">
            <el-form-item label="音频输入倍率">
              <el-input-number v-model="form.input_audio_tokens" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入倍率" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="缓存读取倍率">
              <el-input-number v-model="form.cached_read_tokens" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入倍率" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="推理倍率">
              <el-input-number v-model="form.reasoning_tokens" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入倍率" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="输入文本倍率">
              <el-input-number v-model="form.input_text_tokens" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入倍率" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="输出文本倍率">
              <el-input-number v-model="form.output_text_tokens" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入倍率" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="输入图片倍率">
              <el-input-number v-model="form.input_image_tokens" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入倍率" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="输出图片倍率">
              <el-input-number v-model="form.output_image_tokens" :precision="4" :step="0.0001" style="width: 100%"
                :controls="false" placeholder="请输入倍率" />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="价格来源">
              <el-input v-model="form.price_source" />
            </el-form-item>
          </el-col>
        </el-row>
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
import { ref, computed, onMounted, watch } from 'vue'
import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import { Edit, Delete, Check, Close, Document, Search } from '@element-plus/icons-vue'

const props = defineProps({
  user: Object
})

const prices = ref([])
const dialogVisible = ref(false)
const form = ref({
  model: '',
  model_type: '',
  billing_type: 'tokens',
  channel_type: '',
  currency: 'USD',
  input_price: null,
  output_price: null,
  input_audio_tokens: null,
  cached_read_tokens: null,
  reasoning_tokens: null,
  input_text_tokens: null,
  output_text_tokens: null,
  input_image_tokens: null,
  output_image_tokens: null,
  price_source: '',
  created_by: ''
})
const router = useRouter()
const selectedProvider = ref('')
const selectedModelType = ref('')
const selectedStatus = ref('')
const searchQuery = ref('')

const isAdmin = computed(() => props.user?.role === 'admin')

const providers = ref([])
const getProvider = (id) => {
  // 确保id是字符串类型进行比较
  const idStr = id?.toString()
  return providers.value.find(p => p.id.toString() === idStr)
}

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

// 添加getModelType函数
const getModelType = (type) => {
  if (!type) return ''
  return modelTypeMap.value[type] || type
}

const calculateRate = (price, currency) => {
  if (!price) return 0
  return currency === 'USD' ? (price / 2).toFixed(4) : (price / 14).toFixed(4)
}

const filteredPrices = computed(() => prices.value)

const editingPrice = ref(null)

const loading = ref(true)
const tableLoading = ref(true)

// 添加分页相关的状态
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const cachedPrices = ref(new Map()) // 用于缓存数据

const loadPrices = async () => {
  tableLoading.value = true

  // 构建查询参数
  const params = {
    page: currentPage.value,
    pageSize: pageSize.value
  }

  // 添加筛选参数
  if (selectedProvider.value) {
    params.channel_type = selectedProvider.value
  }
  if (selectedModelType.value) {
    params.model_type = selectedModelType.value
  }
  // 添加状态筛选参数
  if (selectedStatus.value) {
    params.status = selectedStatus.value
  }
  // 添加搜索参数
  if (searchQuery.value) {
    params.search = searchQuery.value
  }

  try {
    const [pricesRes, providersRes] = await Promise.all([
      axios.get('/api/prices', { params }),
      axios.get('/api/providers')
    ])

    prices.value = pricesRes.data.data
    total.value = pricesRes.data.total
    providers.value = providersRes.data

    // 缓存数据
    const cacheKey = `${currentPage.value}-${pageSize.value}-${selectedProvider.value}-${selectedModelType.value}-${selectedStatus.value}-${searchQuery.value}`
    cachedPrices.value.set(cacheKey, {
      prices: pricesRes.data.data,
      total: pricesRes.data.total
    })

    // 限制缓存大小
    if (cachedPrices.value.size > 10) {
      const firstKey = cachedPrices.value.keys().next().value
      cachedPrices.value.delete(firstKey)
    }
  } catch (error) {
    console.error('Failed to load data:', error)
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
    tableLoading.value = false
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
    model_type: '',
    billing_type: 'tokens',
    channel_type: '',
    currency: 'USD',
    input_price: null,
    output_price: null,
    input_audio_tokens: null,
    cached_read_tokens: null,
    reasoning_tokens: null,
    input_text_tokens: null,
    output_text_tokens: null,
    input_image_tokens: null,
    output_image_tokens: null,
    price_source: '',
    created_by: ''
  }
  dialogVisible.value = true
}

const handleQuickEdit = (row) => {
  if (!props.user) {
    router.push('/login')
    ElMessage.warning('请先登录')
    return
  }
  editingPrice.value = row
  // 复制现有数据作为修改建议的基础
  form.value = {
    model: row.model,
    model_type: row.model_type,
    billing_type: row.billing_type,
    channel_type: row.channel_type,
    currency: row.currency,
    input_price: row.input_price,
    output_price: row.output_price,
    input_audio_tokens: row.input_audio_tokens,
    cached_read_tokens: row.cached_read_tokens,
    reasoning_tokens: row.reasoning_tokens,
    input_text_tokens: row.input_text_tokens,
    output_text_tokens: row.output_text_tokens,
    input_image_tokens: row.input_image_tokens,
    output_image_tokens: row.output_image_tokens,
    price_source: row.price_source,
    created_by: props.user.username
  }
  dialogVisible.value = true
}

const submitForm = async () => {
  try {
    form.value.created_by = props.user.username

    // 创建一个新对象用于提交，将 channel_type 转换为数字类型
    const formToSubmit = { ...form.value }
    if (formToSubmit.channel_type) {
      formToSubmit.channel_type = parseInt(formToSubmit.channel_type, 10)
    }

    let response
    if (editingPrice.value) {
      // 更新已存在的价格
      response = await axios.put(`/api/prices/${editingPrice.value.id}`, formToSubmit)
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
          response = await axios.put(`/api/prices/${existingPrice.id}`, formToSubmit)
          handleSubmitResponse(response)
        }).catch(() => {
          // 用户取消更新
        })
        return
      }
      // 创建新价格
      response = await axios.post('/api/prices', formToSubmit)
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
    model_type: '',
    billing_type: 'tokens',
    channel_type: '',
    currency: 'USD',
    input_price: null,
    output_price: null,
    input_audio_tokens: null,
    cached_read_tokens: null,
    reasoning_tokens: null,
    input_text_tokens: null,
    output_text_tokens: null,
    input_image_tokens: null,
    output_image_tokens: null,
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

// 添加模型类型映射
const modelTypeMap = ref({})

// 加载模型类型
const loadModelTypes = async () => {
  try {
    const response = await axios.get('/api/model-types')
    const types = response.data
    const map = {}
    types.forEach(type => {
      map[type.key] = type.label
    })
    modelTypeMap.value = map
  } catch (error) {
    console.error('Failed to load model types:', error)
    ElMessage.error('加载模型类型失败')
  }
}

// 处理新增的模型类型
const handleModelTypeCreate = async (value) => {
  // 如果输入的是中文描述，尝试查找对应的key
  const existingKey = Object.entries(modelTypeMap.value).find(([_, label]) => label === value)?.[0]
  if (existingKey) {
    return existingKey
  }

  // 如果输入的是英文key，直接使用
  let type_key = value
  let type_label = value
  if (!/^[a-zA-Z0-9_]+$/.test(value)) {
    // 如果是中文描述，生成一个新的key
    type_key = `type_${Date.now()}`
    type_label = value
  }

  try {
    await axios.post('/api/model-types', { type_key, type_label })
    modelTypeMap.value[type_key] = type_label
    return type_key
  } catch (error) {
    console.error('Failed to create model type:', error)
    ElMessage.error('创建模型类型失败')
    return 'other'
  }
}

// 创建新行的默认数据
const createNewRow = () => ({
  model: '',
  model_type: '',
  billing_type: 'tokens',
  channel_type: '',
  currency: 'USD',
  input_price: null,
  output_price: null,
  input_audio_tokens: null,
  cached_read_tokens: null,
  reasoning_tokens: null,
  input_text_tokens: null,
  output_text_tokens: null,
  input_image_tokens: null,
  output_image_tokens: null,
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
      // 创建一个新对象用于提交，将 channel_type 转换为数字类型
      const formToSubmit = { ...form }
      if (formToSubmit.channel_type) {
        formToSubmit.channel_type = parseInt(formToSubmit.channel_type, 10)
      }
      await axios.post('/api/prices', formToSubmit)
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
      channel_type: parseInt(provider.id, 10), // 确保是数字类型
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

  let statusFilter = ['pending']
  let statusLabel = '待审核'
  
  // 如果是通过状态，也可以选择已拒绝的价格
  if (status === 'approved') {
    statusFilter = ['pending', 'rejected']
    statusLabel = '待审核或已拒绝'
  }
  
  // 过滤出符合条件的价格
  const filteredPrices = selectedPrices.value.filter(price => statusFilter.includes(price.status))
  if (!filteredPrices.length) {
    ElMessage.warning(`选中的价格中没有${statusLabel}的项目`)
    return
  }

  try {
    // 确认操作
    await ElMessageBox.confirm(
      `确定要${status === 'approved' ? '通过' : '拒绝'}选中的 ${filteredPrices.length} 条${statusLabel}价格吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: status === 'approved' ? 'success' : 'warning'
      }
    )

    // 批量更新状态
    for (const price of filteredPrices) {
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

// 批量删除价格记录
const batchDelete = async () => {
  if (!selectedPrices.value.length) {
    ElMessage.warning('请先选择要删除的价格')
    return
  }

  try {
    // 确认操作
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedPrices.value.length} 条价格吗？此操作不可恢复！`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 批量删除
    for (const price of selectedPrices.value) {
      await axios.delete(`/api/prices/${price.id}`)
    }

    await loadPrices()
    ElMessage.success('批量删除成功')
  } catch (error) {
    if (error === 'cancel') return
    console.error('Failed to batch delete prices:', error)
    ElMessage.error('批量删除失败')
  }
}

// 处理分页变化
const handleSizeChange = (val) => {
  pageSize.value = val
  currentPage.value = 1
  loadPrices()
}

const handleCurrentChange = (val) => {
  currentPage.value = val
  loadPrices()
}

// 监听厂商选择变化
watch(selectedProvider, () => {
  currentPage.value = 1 // 重置到第一页
  loadPrices()
})

// 监听模型类型选择变化
watch(selectedModelType, () => {
  currentPage.value = 1 // 重置到第一页
  loadPrices()
})

// 监听状态选择变化
watch(selectedStatus, () => {
  currentPage.value = 1 // 重置到第一页
  loadPrices()
})

// 监听搜索查询变化
watch(searchQuery, () => {
  // 使用防抖处理，避免频繁请求
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)
  searchDebounceTimer = setTimeout(() => {
    currentPage.value = 1 // 重置到第一页
    loadPrices()
  }, 300)
})

// 添加防抖定时器
let searchDebounceTimer = null

// 复制行
const duplicateRow = (index) => {
  const newRow = { ...batchForms.value[index] }
  batchForms.value.splice(index + 1, 0, newRow)
}

// 删除行
const removeRow = (index) => {
  batchForms.value.splice(index, 1)
  if (batchForms.value.length === 0) {
    addRow() // 如果删除后没有行了，添加一个空行
  }
}

// 添加全部通过功能
const approveAllPending = async () => {
  try {
    // 获取所有待审核的价格数量
    const { data } = await axios.get('/api/prices', {
      params: {
        status: 'pending',
        pageSize: 1
      }
    })
    const pendingCount = data.total

    if (pendingCount === 0) {
      ElMessage.info('当前没有待审核的价格')
      return
    }

    // 确认操作
    await ElMessageBox.confirm(
      `确定要通过所有 ${pendingCount} 条待审核价格吗？`,
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'success'
      }
    )

    // 批量更新所有待审核价格的状态
    const response = await axios.put('/api/prices/approve-all', { action: 'approve' })

    await loadPrices()
    // 使用后端返回的实际审核数量
    ElMessage.success(`已通过 ${response.data.count} 条待审核价格`)
  } catch (error) {
    if (error === 'cancel') return
    console.error('Failed to approve all pending prices:', error)
    ElMessage.error('操作失败')
  }
}

// 处理搜索
const handleSearch = () => {
  currentPage.value = 1 // 重置到第一页
  loadPrices()
}

// 添加检查是否有扩展价格的方法
const hasExtendedPrices = (row) => {
  return row.input_audio_tokens ||
    row.cached_read_tokens ||
    row.reasoning_tokens ||
    row.input_text_tokens ||
    row.output_text_tokens ||
    row.input_image_tokens ||
    row.output_image_tokens
}

onMounted(() => {
  loadModelTypes()
  loadPrices()
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

.status-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.model-type-filters {
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

:deep(.el-loading-spinner) {
  .el-loading-text {
    color: #409EFF;
  }

  .path {
    stroke: #409EFF;
  }
}

.skeleton-row {
  padding: 10px;
  border-bottom: 1px solid #EBEEF5;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
  padding: 0 10px;
}

/* 添加表格行动画 */
:deep(.el-table__body-wrapper) {
  .el-table__row {
    transition: all 0.3s ease;
  }
}

/* 添加分页选择框样式 */
:deep(.el-pagination) {
  .el-select {
    width: auto !important;
    margin: 0 8px;
  }

  .el-select .el-input {
    width: 140px !important;
  }

  .el-select-dropdown__item {
    padding-right: 15px;
  }

  .el-pagination__sizes {
    margin-right: 15px;
  }

  /* 修复选择框宽度问题 */
  .el-select__wrapper {
    min-width: 140px !important;
    width: auto !important;
  }

  /* 确保下拉菜单也足够宽 */
  .el-select__popper {
    min-width: 140px !important;
  }
}

.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.action-buttons :deep(.el-button) {
  padding: 4px;
}

.action-buttons :deep(.el-icon) {
  font-size: 16px;
}

/* 添加全局样式覆盖 */
:global(.el-pagination .el-select__wrapper) {
  min-width: 140px !important;
  width: auto !important;
}

:global(.el-pagination .el-select-dropdown__wrap) {
  min-width: 140px !important;
}

:global(.el-pagination .el-select .el-input__wrapper) {
  width: auto !important;
  min-width: 140px !important;
}

.extended-prices {
  font-size: 12px;
}

.price-item {
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.price-label {
  color: #666;
  min-width: 100px;
}

.price-value {
  font-weight: 500;
  color: #333;
}

.el-tag {
  margin-left: 4px;
}

.price-cards-container {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
  padding: 1rem 0;
}

.price-card {
  background: #fff;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  gap: 1rem;
  transition: all 0.2s ease;
}

.price-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.price-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.provider-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.provider-icon {
  width: 24px;
  height: 24px;
  border-radius: 4px;
}

.provider-name {
  font-weight: 500;
  color: #333;
}

.model-status {
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
}

.model-status.pending {
  background: #fff7e6;
  color: #d46b08;
}

.model-status.approved {
  background: #f6ffed;
  color: #389e0d;
}

.model-status.rejected {
  background: #fff1f0;
  color: #cf1322;
}

.model-info {
  margin-top: 0.5rem;
}

.model-name {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 0.5rem 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.model-meta {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.price-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.price-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.price-label {
  color: #666;
  min-width: 100px;
}

.price-value {
  font-weight: 500;
  color: #333;
}

.extended-prices {
  border-top: 1px solid #f0f0f0;
  padding-top: 1rem;
  margin-top: 0.5rem;
}

.section-title {
  font-weight: 500;
  color: #666;
  margin-bottom: 0.75rem;
}

.extended-price-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0.75rem;
}

.extended-price-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding: 0.5rem;
  background: #f9fafb;
  border-radius: 8px;
}

.ext-price-label {
  font-size: 0.875rem;
  color: #666;
}

.ext-price-value {
  font-weight: 500;
  color: #333;
}

.price-card-footer {
  margin-top: auto;
  padding-top: 1rem;
  border-top: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.meta-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.875rem;
  color: #666;
}

.created-by {
  font-weight: 500;
}

.created-at {
  color: #999;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.skeleton {
  min-height: 300px;
}

:deep(.el-tag) {
  margin: 0;
}

:deep(.el-button) {
  padding: 4px;
}

:deep(.el-icon) {
  font-size: 16px;
}
</style>

<style>
/* 全局样式，确保分页选择框宽度足够 */
.el-pagination .el-select__wrapper {
  min-width: 140px !important;
  width: auto !important;
}

.el-pagination .el-select .el-input__wrapper {
  width: auto !important;
  min-width: 140px !important;
}

.el-select-dropdown {
  min-width: 140px !important;
}
</style>