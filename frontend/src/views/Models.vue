<template>
  <div>
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">AI 模型管理</n-h2>
      <n-button type="primary" @click="showCreateModal = true">
        新建模型
      </n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="models"
      :loading="loading"
      :pagination="pagination"
    />

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showCreateModal" preset="dialog" :title="isEdit ? '编辑模型' : '新建模型'" style="width: 600px">
      <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="left" label-width="100">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formValue.name" placeholder="模型名称 (如 gpt-4o)" />
        </n-form-item>
        <n-form-item label="提供商" path="provider">
          <n-select v-model:value="formValue.provider" :options="providerOptions" />
        </n-form-item>
        <n-form-item label="接口地址" path="baseUrl">
          <n-input v-model:value="formValue.baseUrl" placeholder="https://api.openai.com/v1" />
        </n-form-item>
        <n-form-item label="API 密钥" path="apiKey">
          <n-input v-model:value="formValue.apiKey" type="password" show-password-on="click" placeholder="sk-..." />
        </n-form-item>
        <n-form-item label="配置" path="configJson">
          <n-input
            v-model:value="formValue.configJson"
            type="textarea"
            placeholder="JSON 配置 (可选)"
          />
        </n-form-item>
        <n-form-item label="默认" path="isDefault">
          <n-switch v-model:value="formValue.isDefault" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="testConnection" :loading="testing">测试连接</n-button>
          <n-button @click="closeModal">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? '更新' : '创建' }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from 'vue'
import { NButton, NSpace, useMessage, useDialog, NTag } from 'naive-ui'
import { ListAIModels, CreateAIModel, DeleteAIModel, TestAIModel } from '../../wailsjs/go/main/App'

const message = useMessage()
const dialog = useDialog()

const models = ref([])
const loading = ref(false)
const showCreateModal = ref(false)
const submitting = ref(false)
const testing = ref(false)
const isEdit = ref(false)
const editingId = ref(null)

const formValue = ref({
  name: '',
  provider: 'openai',
  baseUrl: '',
  apiKey: '',
  configJson: '',
  isDefault: false
})

const providerOptions = [
  { label: 'OpenAI', value: 'openai' },
  { label: 'DeepSeek', value: 'deepseek' },
  { label: 'Ollama', value: 'ollama' }
]

const rules = {
  name: { required: true, message: '必填', trigger: 'blur' },
  provider: { required: true, message: '必填', trigger: 'blur' },
  apiKey: { required: true, message: '必填', trigger: 'blur' }
}

const pagination = { pageSize: 10 }

const columns = [
  { title: '名称', key: 'name', width: 150 },
  { title: '提供商', key: 'provider', width: 100 },
  { title: '接口地址', key: 'baseUrl' },
  { 
    title: '默认', 
    key: 'isDefault', 
    width: 80,
    render(row) {
      return row.isDefault ? h(NTag, { type: 'success' }, { default: () => '是' }) : ''
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 150,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, { default: () => '删除' })
        ]
      })
    }
  }
]

async function loadModels() {
  loading.value = true
  try {
    const res = await ListAIModels()
    if (res.code === 200) {
      models.value = res.data || []
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('加载模型失败: ' + e)
  } finally {
    loading.value = false
  }
}

function handleDelete(row) {
  dialog.warning({
    title: '确认删除',
    content: `确认删除模型 "${row.name}"?`,
    positiveText: '确认',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await DeleteAIModel(row.id)
        if (res.code === 200) {
          message.success('已删除')
          loadModels()
        } else {
          message.error(res.msg)
        }
      } catch (e) {
        message.error('删除失败: ' + e)
      }
    }
  })
}

function closeModal() {
  showCreateModal.value = false
  formValue.value = {
    name: '',
    provider: 'openai',
    baseUrl: '',
    apiKey: '',
    configJson: '',
    isDefault: false
  }
  isEdit.value = false
  editingId.value = null
}

async function handleSubmit() {
  if (!formValue.value.name || !formValue.value.apiKey) {
    message.warning('名称和 API Key 为必填项')
    return
  }
  
  submitting.value = true
  try {
    // For now, CreateAIModel handles both create (ID=0) and update (ID!=0) if we passed ID.
    // But currently CreateAIModel in App.go calls repo.Create which forces new entry.
    // We only implemented Create and Delete in App.go for models for now (based on previous turn).
    // Let's check App.go again.
    // Yes, CreateAIModel, ListAIModels, DeleteAIModel, TestAIModel. No UpdateAIModel.
    // So we only support Create for now.
    
    const modelData = { ...formValue.value }
    // Ensure numeric/bool types are correct if needed, but JS to Go handling in Wails is usually good.
    
    const res = await CreateAIModel(modelData)
    
    if (res.code === 200) {
      message.success('创建成功')
      closeModal()
      loadModels()
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('操作失败: ' + e)
  } finally {
    submitting.value = false
  }
}

async function testConnection() {
  testing.value = true
  try {
     const modelData = { ...formValue.value }
     const res = await TestAIModel(modelData)
     if (res.code === 200) {
       message.success('连接成功！')
     } else {
       message.error('连接失败: ' + res.msg)
     }
  } catch (e) {
    message.error('测试失败: ' + e)
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  loadModels()
})
</script>
