<template>
  <div>
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">Agents</n-h2>
      <n-button type="primary" @click="showCreateModal = true">
        New Agent
      </n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="agents"
      :loading="loading"
      :pagination="pagination"
    />

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showCreateModal" preset="dialog" :title="isEdit ? 'Edit Agent' : 'New Agent'" style="width: 700px">
      <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="left" label-width="120">
        <n-form-item label="Name" path="name">
          <n-input v-model:value="formValue.name" placeholder="Agent Name" />
        </n-form-item>
        <n-form-item label="Description" path="description">
          <n-input v-model:value="formValue.description" placeholder="Description" />
        </n-form-item>
        <n-form-item label="Model" path="modelId">
            <n-select v-model:value="formValue.modelId" :options="modelOptions" placeholder="Select AI Model" />
        </n-form-item>
        <n-form-item label="System Prompt" path="systemPrompt">
          <n-input
            v-model:value="formValue.systemPrompt"
            type="textarea"
            :autosize="{ minRows: 5, maxRows: 10 }"
            placeholder="You are a helpful assistant..."
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
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
import { ListAgents, CreateAgent, UpdateAgent, DeleteAgent, ListAIModels } from '../../wailsjs/go/main/App'

const message = useMessage()
const dialog = useDialog()

const agents = ref([])
const modelOptions = ref([])
const loading = ref(false)
const showCreateModal = ref(false)
const submitting = ref(false)
const isEdit = ref(false)
const editingId = ref(null)

const formValue = ref({
  name: '',
  description: '',
  systemPrompt: '',
  modelId: null
})

const rules = {
  name: { required: true, message: 'Required', trigger: 'blur' },
  modelId: { required: true, message: 'Required', type: 'number', trigger: 'blur' }
}

const pagination = { pageSize: 10 }

const columns = [
  { title: 'Name', key: 'name', width: 150 },
  { title: 'Description', key: 'description' },
  { 
      title: 'Model', 
      key: 'Model', 
      width: 150,
      render(row) {
          return row.Model ? h(NTag, { type: 'info' }, { default: () => row.Model.name }) : 'N/A'
      }
  },
  {
    title: 'Action',
    key: 'actions',
    width: 150,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => handleEdit(row) }, { default: () => 'Edit' }),
          h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, { default: () => 'Delete' })
        ]
      })
    }
  }
]

async function loadData() {
  loading.value = true
  try {
    // Load Models first for select options
    const modelRes = await ListAIModels()
    if (modelRes.code === 200) {
        modelOptions.value = (modelRes.data || []).map(m => ({ label: m.name, value: m.id }))
    }

    // Load Agents
    const res = await ListAgents()
    if (res.code === 200) {
      agents.value = res.data || []
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('Failed to load data: ' + e)
  } finally {
    loading.value = false
  }
}

function handleEdit(row) {
  isEdit.value = true
  editingId.value = row.id
  formValue.value = {
    name: row.name,
    description: row.description,
    systemPrompt: row.systemPrompt,
    modelId: row.modelId
  }
  showCreateModal.value = true
}

function handleDelete(row) {
  dialog.warning({
    title: 'Confirm Delete',
    content: `Delete agent "${row.name}"?`,
    positiveText: 'Confirm',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      try {
        const res = await DeleteAgent(row.id)
        if (res.code === 200) {
          message.success('Deleted')
          loadData()
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
  formValue.value = { name: '', description: '', systemPrompt: '', modelId: null }
  isEdit.value = false
  editingId.value = null
}

async function handleSubmit() {
  if (!formValue.value.name || !formValue.value.modelId) {
    message.warning('Name and Model are required')
    return
  }
  
  submitting.value = true
  try {
    let res
    if (isEdit.value) {
      res = await UpdateAgent(editingId.value, formValue.value.name, formValue.value.description, formValue.value.systemPrompt, formValue.value.modelId)
    } else {
      res = await CreateAgent(formValue.value.name, formValue.value.description, formValue.value.systemPrompt, formValue.value.modelId)
    }
    
    if (res.code === 200) {
      message.success(isEdit.value ? 'Updated' : 'Created')
      closeModal()
      loadData()
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('Operation failed: ' + e)
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
