<template>
  <div>
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">模式管理</n-h2>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="modes"
      :loading="loading"
      :pagination="pagination"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { ListModes } from '../../wailsjs/go/main/App'

const message = useMessage()

const modes = ref([])
const loading = ref(false)

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
    title: '标识 (Key)',
    key: 'key',
    width: 150
  },
  {
    title: '名称',
    key: 'name',
    width: 200
  },
  {
    title: '描述',
    key: 'description'
  },
  {
    title: '系统提示词',
    key: 'systemPrompt',
    ellipsis: {
      tooltip: true
    }
  }
]

async function loadModes() {
  loading.value = true
  try {
    const res = await ListModes()
    if (res.code === 200) {
      modes.value = res.data || []
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('加载模式失败: ' + e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadModes()
})
</script>
