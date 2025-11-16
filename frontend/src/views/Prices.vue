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
          <div v-for="price in prices" :key="price.id" class="price-card" :class="price.status">
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
                <span class="copyable-model-name" @click="copyModelName(price.model)" title="点击复制模型名称">
                  {{ price.model }}
                  <el-icon class="copy-icon"><Document /></el-icon>
                </span>
                <el-tag v-if="price.temp_model && price.temp_model !== 'NULL'" 
                  type="warning" size="small" effect="light">
                  待审核: {{ price.temp_model }}
                </el-tag>
              </h3>
              <div class="model-meta">
                <el-tag size="small" effect="plain">{{ getBillingType(price.billing_type) }}</el-tag>
                <el-tag size="small" effect="plain">{{ price.currency }}</el-tag>
              </div>
            </div>

            <div class="price-info new-price-layout">
              <div class="price-box input-price-box">
                <div class="price-value-main">{{ price.input_price === 0 ? '免费' : price.input_price }}</div>
                <div class="price-description">
                  <span class="price-label-small">输入价格</span>
                  <span class="price-unit-small">(M)</span>
                </div>
                <el-tag v-if="price.temp_input_price !== null && price.temp_input_price !== undefined" 
                  type="warning" size="small" effect="light" class="pending-tag">
                  待审核: {{ price.temp_input_price === 0 ? '免费' : price.temp_input_price }}
                </el-tag>
              </div>
              <div class="price-box output-price-box">
                <div class="price-value-main">{{ price.output_price === 0 ? '免费' : price.output_price }}</div>
                <div class="price-description">
                  <span class="price-label-small">输出价格</span>
                  <span class="price-unit-small">(M)</span>
                </div>
                <el-tag v-if="price.temp_output_price !== null && price.temp_output_price !== undefined" 
                  type="warning" size="small" effect="light" class="pending-tag">
                  待审核: {{ price.temp_output_price === 0 ? '免费' : price.temp_output_price }}
                </el-tag>
              </div>
            </div>

            <div class="extended-prices" v-if="hasAnyExtendedPrices(price)">
              <div class="section-title">
                <span>扩展价格</span>
              </div>
              <div class="extended-price-grid">
                <div v-if="hasSpecificPrice(price.input_audio_tokens)" class="extended-price-item">
                  <span class="ext-price-label">音频输入</span>
                  <span class="ext-price-value">{{ price.input_audio_tokens }}</span>
                  <el-tag v-if="price.temp_input_audio_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_input_audio_tokens }}
                  </el-tag>
                </div>
                <div v-if="hasSpecificPrice(price.output_audio_tokens)" class="extended-price-item">
                  <span class="ext-price-label">音频输出</span>
                  <span class="ext-price-value">{{ price.output_audio_tokens }}</span>
                  <el-tag v-if="price.temp_output_audio_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_output_audio_tokens }}
                  </el-tag>
                </div>
                <div v-if="hasSpecificPrice(price.cached_read_tokens)" class="extended-price-item">
                  <span class="ext-price-label">缓存读取</span>
                  <span class="ext-price-value">{{ price.cached_read_tokens }}</span>
                  <el-tag v-if="price.temp_cached_read_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_cached_read_tokens }}
                  </el-tag>
                </div>
                <div v-if="hasSpecificPrice(price.cached_write_tokens)" class="extended-price-item">
                  <span class="ext-price-label">缓存写入</span>
                  <span class="ext-price-value">{{ price.cached_write_tokens }}</span>
                  <el-tag v-if="price.temp_cached_write_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_cached_write_tokens }}
                  </el-tag>
                </div>
                <div v-if="hasSpecificPrice(price.reasoning_tokens)" class="extended-price-item">
                  <span class="ext-price-label">推理</span>
                  <span class="ext-price-value">{{ price.reasoning_tokens }}</span>
                  <el-tag v-if="price.temp_reasoning_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_reasoning_tokens }}
                  </el-tag>
                </div>
                <div v-if="hasSpecificPrice(price.input_text_tokens)" class="extended-price-item">
                  <span class="ext-price-label">文本输入</span>
                  <span class="ext-price-value">{{ price.input_text_tokens }}</span>
                  <el-tag v-if="price.temp_input_text_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_input_text_tokens }}
                  </el-tag>
                </div>
                <div v-if="hasSpecificPrice(price.output_text_tokens)" class="extended-price-item">
                  <span class="ext-price-label">文本输出</span>
                  <span class="ext-price-value">{{ price.output_text_tokens }}</span>
                  <el-tag v-if="price.temp_output_text_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_output_text_tokens }}
                  </el-tag>
                </div>
                <div v-if="hasSpecificPrice(price.input_image_tokens)" class="extended-price-item">
                  <span class="ext-price-label">图片输入</span>
                  <span class="ext-price-value">{{ price.input_image_tokens }}</span>
                  <el-tag v-if="price.temp_input_image_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_input_image_tokens }}
                  </el-tag>
                </div>
                <div v-if="hasSpecificPrice(price.output_image_tokens)" class="extended-price-item">
                  <span class="ext-price-label">图片输出</span>
                  <span class="ext-price-value">{{ price.output_image_tokens }}</span>
                  <el-tag v-if="price.temp_output_image_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_output_image_tokens }}
                  </el-tag>
                </div>
                <div v-if="hasSpecificPrice(price.cached_tokens)" class="extended-price-item">
                  <span class="ext-price-label">缓存</span>
                  <span class="ext-price-value">{{ price.cached_tokens }}</span>
                  <el-tag v-if="price.temp_cached_tokens" type="warning" size="small" effect="light" class="temp-tag">
                    {{ price.temp_cached_tokens }}
                  </el-tag>
                </div>
              </div>
            </div>

            <div class="price-card-footer">
              <div class="meta-info">
                <span class="updated-by"><el-icon><User /></el-icon> {{ price.updated_by || price.created_by }}</span>
                <span class="updated-at"><el-icon><Timer /></el-icon> {{ new Date(price.updated_at).toLocaleString() }}</span>
                <div v-if="price.price_source" class="price-source">
                  <span class="source-label"><el-icon><InfoFilled /></el-icon></span>
                  <a v-if="isValidUrl(price.price_source)" :href="price.price_source" target="_blank" class="source-link">
                    <span>{{ formatSourceUrl(price.price_source) }}</span>
                  </a>
                  <span v-else>{{ price.price_source }}</span>
                </div>
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
          <el-table-column label="扩展价格" min-width="280">
            <template #default="{ row }">
              <el-dropdown trigger="click">
                <el-button type="primary" plain size="small">
                  设置扩展价格 <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <div class="extended-price-dropdown">
                      <div class="dropdown-title">扩展价格设置</div>
                      <div class="dropdown-row">
                        <span>音频输入价格:</span>
                        <el-input-number v-model="row.input_audio_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                      <div class="dropdown-row">
                        <span>音频输出价格:</span>
                        <el-input-number v-model="row.output_audio_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                      <div class="dropdown-row">
                        <span>缓存读取价格:</span>
                        <el-input-number v-model="row.cached_read_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                      <div class="dropdown-row">
                        <span>缓存写入价格:</span>
                        <el-input-number v-model="row.cached_write_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                      <div class="dropdown-row">
                        <span>推理价格:</span>
                        <el-input-number v-model="row.reasoning_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                      <div class="dropdown-row">
                        <span>输入文本价格:</span>
                        <el-input-number v-model="row.input_text_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                      <div class="dropdown-row">
                        <span>输出文本价格:</span>
                        <el-input-number v-model="row.output_text_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                      <div class="dropdown-row">
                        <span>输入图片价格:</span>
                        <el-input-number v-model="row.input_image_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                      <div class="dropdown-row">
                        <span>输出图片价格:</span>
                        <el-input-number v-model="row.output_image_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                      <div class="dropdown-row">
                        <span>缓存价格:</span>
                        <el-input-number v-model="row.cached_tokens" :precision="4" :step="0.0001" 
                          :controls="false" :min="0" placeholder="请输入价格" />
                      </div>
                    </div>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <div v-if="hasAnyExtendedPrices(row)" class="batch-extended-prices">
                <div v-if="row.input_audio_tokens" class="batch-price-tag">音频输入: {{ row.input_audio_tokens }}</div>
                <div v-if="row.output_audio_tokens" class="batch-price-tag">音频输出: {{ row.output_audio_tokens }}</div>
                <div v-if="row.cached_read_tokens" class="batch-price-tag">缓存读取: {{ row.cached_read_tokens }}</div>
                <div v-if="row.cached_write_tokens" class="batch-price-tag">缓存写入: {{ row.cached_write_tokens }}</div>
                <div v-if="row.reasoning_tokens" class="batch-price-tag">推理: {{ row.reasoning_tokens }}</div>
                <div v-if="row.input_text_tokens" class="batch-price-tag">输入文本: {{ row.input_text_tokens }}</div>
                <div v-if="row.output_text_tokens" class="batch-price-tag">输出文本: {{ row.output_text_tokens }}</div>
                <div v-if="row.input_image_tokens" class="batch-price-tag">输入图片: {{ row.input_image_tokens }}</div>
                <div v-if="row.output_image_tokens" class="batch-price-tag">输出图片: {{ row.output_image_tokens }}</div>
                <div v-if="row.cached_tokens" class="batch-price-tag">缓存: {{ row.cached_tokens }}</div>
              </div>
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
            <div class="extended-price-header">
              <h3>扩展价格（可选）</h3>
              <p class="extended-price-tip">这些价格字段仅适用于支持特定功能的模型，留空表示不使用</p>
            </div>
          </el-col>
          <el-col :span="24">
            <div class="extended-price-container">
              <div v-if="selectedExtensionPrices.length === 0" class="no-extensions">
                暂无扩展价格，点击下方按钮添加
              </div>
              <template v-else>
                <div v-for="(type, index) in selectedExtensionPrices" :key="type" class="extension-item">
                  <div class="extension-header">
                    <span class="extension-label">{{ getExtensionLabel(type) }}</span>
                    <el-button type="danger" size="small" circle @click="removeExtension(index)">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                  <el-input-number v-model="form[type]" :precision="4" :step="0.0001" style="width: 100%"
                    :controls="false" placeholder="请输入价格" />
                </div>
              </template>
              <el-dropdown @command="addExtension" trigger="click">
                <el-button type="primary" class="add-extension-btn">
                  添加扩展价格 <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <div class="extended-type-dropdown">
                      <div class="dropdown-title">选择扩展价格类型</div>
                      <el-dropdown-item 
                        v-for="(label, type) in availableExtensionTypes" 
                        :key="type" 
                        :disabled="isExtensionSelected(type)"
                        :command="type"
                      >
                        {{ label }}
                      </el-dropdown-item>
                    </div>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
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
import { Edit, Delete, Check, Close, Document, Search, ArrowDown, User, Timer, InfoFilled } from '@element-plus/icons-vue'
import { isModerator } from '@/utils/permission'

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
  input_price: null,
  output_price: null,
  input_audio_tokens: null,
  output_audio_tokens: null,
  cached_tokens: null,
  cached_read_tokens: null,
  cached_write_tokens: null,
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
const selectedStatus = ref('')
const searchQuery = ref('')

// 使用新的权限判定：t4或admin可以审核价格
const isAdmin = computed(() => isModerator(props.user))

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

// 检查URL是否有效
const isValidUrl = (url) => {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

// 格式化URL以便显示
const formatSourceUrl = (url) => {
  try {
    const urlObj = new URL(url)
    return urlObj.hostname + (urlObj.pathname !== '/' ? urlObj.pathname : '')
  } catch {
    return url
  }
}

// 复制模型名称到剪贴板
const copyModelName = (modelName) => {
  navigator.clipboard.writeText(modelName)
    .then(() => {
      ElMessage({
        message: `已复制模型名称: ${modelName}`,
        type: 'success',
        duration: 2000
      })
    })
    .catch(err => {
      console.error('复制失败:', err)
      ElMessage.error('复制失败')
    })
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
    const cacheKey = `${currentPage.value}-${pageSize.value}-${selectedProvider.value}-${selectedStatus.value}-${searchQuery.value}`
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
  setExtensionPricesFromPrice(price)
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
    input_price: null,
    output_price: null,
    input_audio_tokens: null,
    output_audio_tokens: null,
    cached_tokens: null,
    cached_read_tokens: null,
    cached_write_tokens: null,
    reasoning_tokens: null,
    input_text_tokens: null,
    output_text_tokens: null,
    input_image_tokens: null,
    output_image_tokens: null,
    price_source: '',
    created_by: ''
  }
  resetExtensionPrices()
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
    billing_type: row.billing_type,
    channel_type: row.channel_type,
    currency: row.currency,
    input_price: row.input_price,
    output_price: row.output_price,
    input_audio_tokens: row.input_audio_tokens,
    output_audio_tokens: row.output_audio_tokens,
    cached_tokens: row.cached_tokens,
    cached_read_tokens: row.cached_read_tokens,
    cached_write_tokens: row.cached_write_tokens,
    reasoning_tokens: row.reasoning_tokens,
    input_text_tokens: row.input_text_tokens,
    output_text_tokens: row.output_text_tokens,
    input_image_tokens: row.input_image_tokens,
    output_image_tokens: row.output_image_tokens,
    price_source: row.price_source,
    created_by: props.user.username
  }
  setExtensionPricesFromPrice(row)
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
    billing_type: 'tokens',
    channel_type: '',
    currency: 'USD',
    input_price: null,
    output_price: null,
    input_audio_tokens: null,
    output_audio_tokens: null,
    cached_tokens: null,
    cached_read_tokens: null,
    cached_write_tokens: null,
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

// 创建新行的默认数据
const createNewRow = () => ({
  model: '',
  billing_type: 'tokens',
  channel_type: '',
  currency: 'USD',
  input_price: null,
  output_price: null,
  input_audio_tokens: null,
  output_audio_tokens: null,
  cached_tokens: null,
  cached_read_tokens: null,
  cached_write_tokens: null,
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
    row.output_audio_tokens ||
    row.cached_read_tokens ||
    row.cached_write_tokens ||
    row.reasoning_tokens ||
    row.input_text_tokens ||
    row.output_text_tokens ||
    row.input_image_tokens ||
    row.output_image_tokens
}

// 添加更细粒度的检查函数
const hasAnyExtendedPrices = (row) => {
  return hasSpecificPrice(row.input_audio_tokens) ||
    hasSpecificPrice(row.output_audio_tokens) ||
    hasSpecificPrice(row.cached_tokens) ||
    hasSpecificPrice(row.cached_read_tokens) ||
    hasSpecificPrice(row.cached_write_tokens) ||
    hasSpecificPrice(row.reasoning_tokens) ||
    hasSpecificPrice(row.input_text_tokens) ||
    hasSpecificPrice(row.output_text_tokens) ||
    hasSpecificPrice(row.input_image_tokens) ||
    hasSpecificPrice(row.output_image_tokens)
}

// 检查具体价格字段是否存在且有效
const hasSpecificPrice = (price) => {
  return price !== null && price !== undefined && price !== ''
}

// 扩展价格类型定义
const extensionTypes = {
  input_audio_tokens: '音频输入价格',
  output_audio_tokens: '音频输出价格',
  cached_tokens: '缓存价格',
  cached_read_tokens: '缓存读取价格',
  cached_write_tokens: '缓存写入价格',
  reasoning_tokens: '推理价格',
  input_text_tokens: '输入文本价格',
  output_text_tokens: '输出文本价格',
  input_image_tokens: '输入图片价格',
  output_image_tokens: '输出图片价格'
}

// 选中的扩展价格类型
const selectedExtensionPrices = ref([])

// 可用的扩展价格类型
const availableExtensionTypes = computed(() => {
  return extensionTypes
})

// 获取扩展类型的标签
const getExtensionLabel = (type) => {
  return extensionTypes[type] || type
}

// 检查扩展类型是否已选中
const isExtensionSelected = (type) => {
  return selectedExtensionPrices.value.includes(type)
}

// 添加扩展价格
const addExtension = (type) => {
  if (!isExtensionSelected(type)) {
    selectedExtensionPrices.value.push(type)
    form.value[type] = null
  }
}

// 移除扩展价格
const removeExtension = (index) => {
  const type = selectedExtensionPrices.value[index]
  form.value[type] = null
  selectedExtensionPrices.value.splice(index, 1)
}

// 重置扩展价格选择
const resetExtensionPrices = () => {
  selectedExtensionPrices.value = []
  Object.keys(extensionTypes).forEach(type => {
    form.value[type] = null
  })
}

// 从价格对象中设置扩展价格选择
const setExtensionPricesFromPrice = (price) => {
  selectedExtensionPrices.value = []
  Object.keys(extensionTypes).forEach(type => {
    if (price[type] !== undefined && price[type] !== null) {
      selectedExtensionPrices.value.push(type)
    }
  })
}

onMounted(() => {
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

.skeleton {
  min-height: 240px;
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
  border-radius: 8px;
  padding: 0.75rem; /* 减小内边距 */
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 1px 2px rgba(0, 0, 0, 0.1); /* 更柔和的阴影 */
  display: flex;
  flex-direction: column;
  gap: 0.5rem; /* 减小元素间距 */
  transition: all 0.2s ease;
  position: relative;
  overflow: hidden;
  height: auto;
  min-height: 200px; /* 减小最小高度 */
  border: 1px solid #e0e0e0; /* 调整边框颜色，使其更明显一点 */
  /* 移除左侧彩色条相关的 ::before 伪元素样式 */
}

.price-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1); /* 稍微加深悬停阴影 */
  transform: translateY(-2px);
  border-color: #d0d0d0; /* 悬停时边框颜色加深 */
}

/* .price-card::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  width: 4px; 
  height: 100%;
  background: #409EFF; 
}

.price-card.pending::before {
  background: #E6A23C; 
}

.price-card.rejected::before {
  background: #F56C6C; 
} */

.price-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #f0f0f0; /* 将分隔线改为实线，更清晰 */
}

.provider-info {
  display: flex;
  align-items: center;
  gap: 0.5rem; /* 减小间距 */
}

.provider-icon {
  width: 20px; /* 减小图标大小 */
  height: 20px;
  border-radius: 4px;
  object-fit: contain;
  background-color: #f5f7fa;
  padding: 2px;
}

.provider-name {
  font-weight: 500; /* 减轻字重 */
  color: #606266; /* 使用 Element Plus 文本颜色 */
  font-size: 0.9rem; /* 减小字体 */
}

.model-status {
  padding: 0.15rem 0.5rem; /* 减小内边距 */
  border-radius: 4px; /* 减小圆角 */
  font-size: 0.7rem; /* 减小字体 */
  font-weight: 500;
  letter-spacing: 0.5px;
}

.model-status.pending {
  background: #fdf6ec; /* 使用 Element Plus 警告色背景 */
  color: #E6A23C;
}

.model-status.approved {
  background: #f0f9eb; /* 使用 Element Plus 成功色背景 */
  color: #67C23A;
}

.model-status.rejected {
  background: #fef0f0; /* 使用 Element Plus 危险色背景 */
  color: #F56C6C;
}

.model-info {
  margin-top: 0.5rem;
  min-height: 50px; /* 减小最小高度 */
  display: flex;
  flex-direction: column;
}

.model-name {
  font-size: 1.1rem; /* 减小字体 */
  font-weight: 600; /* 减轻字重 */
  color: #303133; /* 使用 Element Plus 标题颜色 */
  line-height: 1.3;
  max-height: calc(1.3em * 2);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  text-overflow: ellipsis;
  margin: 0 0 0.3rem 0; /* 减小下边距 */
}

.copyable-model-name {
  display: inline-flex;
  align-items: center;
  gap: 4px; /* 减小间距 */
  cursor: pointer;
  position: relative;
  padding-right: 20px; /* 减小右内边距 */
  transition: color 0.2s ease;
}

.copyable-model-name:hover {
  color: #409EFF;
}

.copy-icon {
  font-size: 12px; /* 减小图标 */
  opacity: 0.5;
  position: absolute;
  right: 0;
  transition: all 0.2s ease;
}

.copyable-model-name:hover .copy-icon {
  opacity: 1;
  color: #409EFF;
}

.model-meta {
  display: flex;
  gap: 0.3rem; /* 减小间距 */
  flex-wrap: wrap;
  margin-bottom: 0.5rem; /* 减小下边距 */
}

.price-info {
  display: flex;
  flex-direction: row; /* 直接设置为水平排列 */
  justify-content: space-between;
  gap: 0.5rem; /* 减小间距 */
  padding: 0.5rem;
  margin-top: 0.3rem; /* 减小上边距 */
  background-color: #f5f7fa; /* 更淡的背景色 */
  border-radius: 4px; /* 减小圆角 */
}

.price-box {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 0.3rem; /* 减小内边距 */
  text-align: center;
  position: relative;
}

.input-price-box::after {
  content: '';
  position: absolute;
  right: -0.25rem; /* 调整分隔线位置 */
  top: 20%;
  bottom: 20%;
  width: 1px;
  background-color: #dcdfe6; /* 使用 Element Plus 边框颜色 */
}

.price-value-main {
  font-size: 1.2rem; /* 减小价格字体 */
  font-weight: 500; /* 减轻字重 */
  color: #303133; /* 使用 Element Plus 文本颜色 */
  line-height: 1.2;
}

.price-description {
  margin-top: 0.1rem; /* 减小间距 */
}

.price-label-small {
  font-size: 0.7rem; /* 减小描述字体 */
  color: #909399; /* 使用 Element Plus 次要文本颜色 */
}

.price-unit-small {
  font-size: 0.65rem; /* 减小单位字体 */
  color: #c0c4cc; /* 使用 Element Plus 占位符颜色 */
  margin-left: 2px;
}

.pending-tag {
  margin-top: 0.2rem; /* 减小上边距 */
  font-size: 0.6rem; /* 减小字体 */
  padding: 0 3px; /* 减小内边距 */
  height: auto;
}


.price-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.price-label {
  color: #606266;
  min-width: 100px;
  font-weight: 500;
}

.price-value {
  font-weight: 600;
  color: #303133;
}

.extended-prices {
  border-top: 1px solid #f0f0f0; /* 将分隔线改为实线 */
  padding-top: 0.3rem; /* 减小上内边距 */
  margin-top: 0.2rem; /* 减小上边距 */
}

.section-title {
  font-weight: 500; /* 减轻字重 */
  color: #909399; /* 使用 Element Plus 次要文本颜色 */
  font-size: 0.75rem; /* 减小字体 */
  margin-bottom: 0.3rem; /* 减小下边距 */
  display: flex;
  align-items: center;
}

.extended-price-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem; /* 减小间距 */
  margin-bottom: 0.3rem; /* 减小下边距 */
}

.extended-price-item {
  font-size: 0.65rem; /* 减小字体 */
  display: inline-flex;
  align-items: center;
  padding: 0.2rem 0.3rem; /* 减小内边距 */
  background: #f5f7fa; /* 使用与价格信息相同的背景色 */
  border-radius: 3px; /* 减小圆角 */
  border: none; /* 移除边框 */
  transition: all 0.2s ease;
  max-width: fit-content;
}

.extended-price-item:hover {
  background: #ecf5ff; /* 使用 Element Plus 主色调背景 */
}

.ext-price-label {
  font-size: 0.65rem; /* 减小字体 */
  color: #909399; /* 使用 Element Plus 次要文本颜色 */
  margin-right: 0.2rem;
}

.ext-price-value {
  font-weight: 500;
  color: #606266; /* 使用 Element Plus 主要文本颜色 */
  font-size: 0.65rem; /* 减小字体 */
}

.price-card-footer {
  margin-top: auto;
  padding-top: 0.5rem; /* 减小上内边距 */
  border-top: 1px solid #f0f0f0; /* 将分隔线改为实线 */
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.meta-info {
  display: flex;
  flex-direction: column;
  gap: 0.3rem; /* 减小间距 */
  font-size: 0.7rem; /* 减小字体 */
  color: #909399; /* 使用 Element Plus 次要文本颜色 */
}

.meta-info .el-icon {
  margin-right: 3px; /* 减小右边距 */
  vertical-align: middle;
  font-size: 0.8rem; /* 减小图标 */
}

.updated-by {
  font-weight: normal; /* 使用正常字重 */
  display: flex;
  align-items: center;
}

.updated-at {
  color: #909399; /* 使用 Element Plus 次要文本颜色 */
  font-weight: normal; /* 使用正常字重 */
  display: flex;
  align-items: center;
}

.price-source {
  display: flex;
  align-items: center;
  margin-top: 1px; /* 减小上边距 */
}

.source-label {
  display: flex;
  align-items: center;
  margin-right: 3px; /* 减小右边距 */
}

.source-link {
  color: #409EFF;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 2px; /* 减小间距 */
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.skeleton {
  min-height: 240px;
}

:deep(.el-tag) {
  margin: 0;
  font-size: 0.7rem;
}

:deep(.el-button) {
  padding: 4px;
}

:deep(.el-icon) {
  font-size: 16px;
}

.extended-price-header {
  margin: 20px 0 10px 0;
  border-top: 1px solid #EBEEF5;
  padding-top: 20px;
}

.extended-price-header h3 {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 10px 0;
}

.extended-price-tip {
  font-size: 13px;
  color: #909399;
  margin: 0;
}

.extended-price-dropdown {
  padding: 15px;
  min-width: 300px;
}

.dropdown-title {
  font-weight: 600;
  color: #303133;
  margin-bottom: 15px;
  font-size: 15px;
  border-bottom: 1px solid #EBEEF5;
  padding-bottom: 10px;
}

.dropdown-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.dropdown-row span {
  flex: 0 0 110px;
  font-size: 14px;
  color: #606266;
}

.dropdown-row .el-input-number {
  flex: 1;
}

.batch-extended-prices {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.batch-price-tag {
  background-color: #ecf5ff;
  color: #409EFF;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
  white-space: nowrap;
}

.extended-price-container {
  border: 1px solid #EBEEF5;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  background-color: #f9fafc;
}

.no-extensions {
  color: #909399;
  text-align: center;
  margin-bottom: 20px;
  padding: 20px 0;
  font-size: 14px;
}

.extension-item {
  border: 1px solid #EBEEF5;
  border-radius: 6px;
  padding: 16px;
  margin-bottom: 16px;
  background-color: #fff;
}

.extension-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.extension-label {
  font-weight: 500;
  color: #303133;
  font-size: 15px;
}

.add-extension-btn {
  width: 100%;
  margin-top: 10px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.extended-type-dropdown {
  min-width: 220px;
  padding: 15px;
}

:deep(.el-dropdown-menu__item.is-disabled) {
  color: #C0C4CC;
  cursor: not-allowed;
}

.ext-price-label::after {
  content: ": ";
}

.temp-tag {
  margin-left: 4px;
  font-size: 0.65rem !important;
  height: 18px;
  line-height: 16px;
  padding: 0 4px;
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
