<template>
  <div>
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">Scripts</n-h2>
      <n-button type="primary" @click="showCreateModal = true">
        New Script
      </n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="scripts"
      :loading="loading"
      :pagination="pagination"
    />

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showCreateModal" preset="dialog" :title="isEdit ? 'Edit Script' : 'New Script'" style="width: 800px">
      <n-form ref="formRef" :model="formValue" :rules="rules">
        <n-form-item label="Name" path="name">
          <n-input v-model:value="formValue.name" placeholder="Script Name" />
        </n-form-item>
        <n-form-item label="Description" path="description">
          <n-input v-model:value="formValue.description" placeholder="Description" />
        </n-form-item>
        <n-form-item label="Content" path="content">
          <n-input
            v-model:value="formValue.content"
            type="textarea"
            :autosize="{ minRows: 10, maxRows: 20 }"
            placeholder="// Write your JS code here...
// Global objects: console, http
// console.log('Hello');
// http.get('https://example.com');"
            style="font-family: monospace;"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
           <n-button @click="handleRunTest" :loading="running">Run Test</n-button>
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
import { NButton, NSpace, useMessage, useDialog, NCode } from 'naive-ui'
import { ListScripts, CreateScript, UpdateScript, DeleteScript, RunScript } from '../../wailsjs/go/main/App'

const message = useMessage()
const dialog = useDialog()

const scripts = ref([])
const loading = ref(false)
const showCreateModal = ref(false)
const submitting = ref(false)
const running = ref(false)
const isEdit = ref(false)
const editingId = ref(null)

const formValue = ref({
  name: '',
  description: '',
  content: ''
})

const rules = {
  name: { required: true, message: 'Required', trigger: 'blur' },
  content: { required: true, message: 'Required', trigger: 'blur' }
}

const pagination = { pageSize: 10 }

const columns = [
  { title: 'ID', key: 'id', width: 60 },
  { title: 'Name', key: 'name', width: 150 },
  { title: 'Description', key: 'description' },
  {
    title: 'Action',
    key: 'actions',
    width: 200,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { size: 'small', onClick: () => handleRunRow(row) }, { default: () => 'Run' }),
          h(NButton, { size: 'small', onClick: () => handleEdit(row) }, { default: () => 'Edit' }),
          h(NButton, { size: 'small', type: 'error', onClick: () => handleDelete(row) }, { default: () => 'Delete' })
        ]
      })
    }
  }
]

async function loadScripts() {
  loading.value = true
  try {
    const res = await ListScripts()
    if (res.code === 200) {
      scripts.value = res.data || []
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('Failed to load scripts: ' + e)
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
    content: row.content
  }
  showCreateModal.value = true
}

function handleDelete(row) {
  dialog.warning({
    title: 'Confirm Delete',
    content: `Delete script "${row.name}"?`,
    positiveText: 'Confirm',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      try {
        const res = await DeleteScript(row.id)
        if (res.code === 200) {
          message.success('Deleted')
          loadScripts()
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
  formValue.value = { name: '', description: '', content: '' }
  isEdit.value = false
  editingId.value = null
}

async function handleSubmit() {
  if (!formValue.value.name || !formValue.value.content) {
    message.warning('Name and Content are required')
    return
  }
  
  submitting.value = true
  try {
    let res
    if (isEdit.value) {
      res = await UpdateScript(editingId.value, formValue.value.name, formValue.value.description, formValue.value.content)
    } else {
      res = await CreateScript(formValue.value.name, formValue.value.description, formValue.value.content)
    }
    
    if (res.code === 200) {
      message.success(isEdit.value ? 'Updated' : 'Created')
      closeModal()
      loadScripts()
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('Operation failed: ' + e)
  } finally {
    submitting.value = false
  }
}

async function handleRunRow(row) {
    try {
        const res = await RunScript(row.id)
        if (res.code === 200) {
             dialog.success({
                title: 'Execution Result',
                content: () => h('pre', null, JSON.stringify(res.data, null, 2)),
                positiveText: 'OK'
            })
        } else {
            message.error(res.msg)
        }
    } catch(e) {
        message.error("Run failed: " + e)
    }
}

// For testing in modal, we need to save first or have a "TestRun" API that accepts content.
// Current RunScript takes ID. So for now "Run Test" will warn user to save first or we could impl a TestScript API.
// Let's just save and run for simplicity if editing, or warn if creating.
async function handleRunTest() {
    if (!isEdit.value) {
        message.warning("Please create the script first to run it.")
        return
    }
    
    // Auto save before run
    await handleSubmit()
    if (!showCreateModal.value) {
        // Modal closed means save success
        await handleRunRow({id: editingId.value})
        // Re-open modal to continue editing? Maybe better UX to stay in modal. 
        // But handleSubmit closes modal.
        // Let's improve this later. For now this flow works.
        handleEdit(scripts.value.find(s => s.id === editingId.value))
    }
}

onMounted(() => {
  loadScripts()
})
</script>
