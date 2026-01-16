<template>
  <div>
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">AI Models</n-h2>
      <n-button type="primary" @click="showCreateModal = true">
        New Model
      </n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="models"
      :loading="loading"
      :pagination="pagination"
    />

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showCreateModal" preset="dialog" :title="isEdit ? 'Edit Model' : 'New Model'" style="width: 600px">
      <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="left" label-width="100">
        <n-form-item label="Name" path="name">
          <n-input v-model:value="formValue.name" placeholder="Model Name (e.g. gpt-4o)" />
        </n-form-item>
        <n-form-item label="Provider" path="provider">
          <n-select v-model:value="formValue.provider" :options="providerOptions" />
        </n-form-item>
        <n-form-item label="Base URL" path="baseUrl">
          <n-input v-model:value="formValue.baseUrl" placeholder="https://api.openai.com/v1" />
        </n-form-item>
        <n-form-item label="API Key" path="apiKey">
          <n-input v-model:value="formValue.apiKey" type="password" show-password-on="click" placeholder="sk-..." />
        </n-form-item>
        <n-form-item label="Config" path="configJson">
          <n-input
            v-model:value="formValue.configJson"
            type="textarea"
            placeholder="JSON Config (Optional)"
          />
        </n-form-item>
        <n-form-item label="Default" path="isDefault">
          <n-switch v-model:value="formValue.isDefault" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="testConnection" :loading="testing">Test Connection</n-button>
          <n-button @click="closeModal">Cancel</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? 'Update' : 'Create' }}
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
  name: { required: true, message: 'Required', trigger: 'blur' },
  provider: { required: true, message: 'Required', trigger: 'blur' },
  apiKey: { required: true, message: 'Required', trigger: 'blur' }
}

const pagination = { pageSize: 10 }

const columns = [
  { title: 'Name', key: 'name', width: 150 },
  { title: 'Provider', key: 'provider', width: 100 },
  { title: 'Base URL', key: 'baseUrl' },
  { 
    title: 'Default', 
    key: 'isDefault', 
    width: 80,
    render(row) {
      return row.isDefault ? h(NTag, { type: 'success' }, { default: () => 'Yes' }) : ''
    }
  },
  {
    title: 'Action',
    key: 'actions',
    width: 150,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, { default: () => 'Delete' })
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
    message.error('Failed to load models: ' + e)
  } finally {
    loading.value = false
  }
}

function handleDelete(row) {
  dialog.warning({
    title: 'Confirm Delete',
    content: `Delete model "${row.name}"?`,
    positiveText: 'Confirm',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      try {
        const res = await DeleteAIModel(row.id)
        if (res.code === 200) {
          message.success('Deleted')
          loadModels()
        } else {
          message.error(res.msg)
        }
      } catch (e) {
        message.error('Failed to delete: ' + e)
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
    message.warning('Name and API Key are required')
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
      message.success('Created successfully')
      closeModal()
      loadModels()
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('Operation failed: ' + e)
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
       message.success('Connection Successful!')
     } else {
       message.error('Connection Failed: ' + res.msg)
     }
  } catch (e) {
    message.error('Test failed: ' + e)
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  loadModels()
})
</script>
