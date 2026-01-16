<template>
  <div>
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">Projects</n-h2>
      <n-button type="primary" @click="showCreateModal = true">
        New Project
      </n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="projects"
      :loading="loading"
      :pagination="pagination"
    />

    <!-- Create/Edit Modal -->
    <n-modal v-model:show="showCreateModal" preset="dialog" :title="isEdit ? 'Edit Project' : 'New Project'">
      <n-form ref="formRef" :model="formValue" :rules="rules">
        <n-form-item label="Name" path="name">
          <n-input v-model:value="formValue.name" placeholder="Project Name" />
        </n-form-item>
        <n-form-item label="Description" path="description">
          <n-input
            v-model:value="formValue.description"
            type="textarea"
            placeholder="Description"
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
import { NButton, NSpace, useMessage, useDialog } from 'naive-ui'
import { ListProjects, CreateProject, UpdateProject, DeleteProject } from '../../wailsjs/go/main/App'

const message = useMessage()
const dialog = useDialog()

const projects = ref([])
const loading = ref(false)
const showCreateModal = ref(false)
const submitting = ref(false)
const isEdit = ref(false)
const editingId = ref(null)

const formValue = ref({
  name: '',
  description: ''
})

const rules = {
  name: {
    required: true,
    message: 'Please enter project name',
    trigger: 'blur'
  }
}

const pagination = {
  pageSize: 10
}

const columns = [
  {
    title: 'ID',
    key: 'id',
    width: 80
  },
  {
    title: 'Name',
    key: 'name',
    width: 200
  },
  {
    title: 'Description',
    key: 'description'
  },
  {
    title: 'Created At',
    key: 'createdAt',
    width: 200,
    render(row) {
      return new Date(row.createdAt).toLocaleString()
    }
  },
  {
    title: 'Action',
    key: 'actions',
    width: 150,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(
            NButton,
            {
              size: 'small',
              onClick: () => handleEdit(row)
            },
            { default: () => 'Edit' }
          ),
          h(
            NButton,
            {
              size: 'small',
              type: 'error',
              onClick: () => handleDelete(row)
            },
            { default: () => 'Delete' }
          )
        ]
      })
    }
  }
]

async function loadProjects() {
  loading.value = true
  try {
    const res = await ListProjects()
    if (res.code === 200) {
      projects.value = res.data || []
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('Failed to load projects: ' + e)
  } finally {
    loading.value = false
  }
}

function handleEdit(row) {
  isEdit.value = true
  editingId.value = row.id
  formValue.value = {
    name: row.name,
    description: row.description
  }
  showCreateModal.value = true
}

function handleDelete(row) {
  dialog.warning({
    title: 'Confirm Delete',
    content: `Are you sure to delete project "${row.name}"?`,
    positiveText: 'Confirm',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      try {
        const res = await DeleteProject(row.id)
        if (res.code === 200) {
          message.success('Deleted successfully')
          loadProjects()
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
  formValue.value = { name: '', description: '' }
  isEdit.value = false
  editingId.value = null
}

async function handleSubmit() {
  if (!formValue.value.name) {
    message.warning('Please enter name')
    return
  }
  
  submitting.value = true
  try {
    let res
    if (isEdit.value) {
      res = await UpdateProject(editingId.value, formValue.value.name, formValue.value.description)
    } else {
      res = await CreateProject(formValue.value.name, formValue.value.description)
    }
    
    if (res.code === 200) {
      message.success(isEdit.value ? 'Updated successfully' : 'Created successfully')
      closeModal()
      loadProjects()
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
  loadProjects()
})
</script>
