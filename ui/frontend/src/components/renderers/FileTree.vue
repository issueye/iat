<template>
  <div class="file-tree-container">
    <div class="tree-header">
      <n-icon><FolderOpenOutline /></n-icon>
      <span>文件目录</span>
    </div>
    <n-tree
      block-line
      expand-on-click
      :data="treeData"
      :render-label="renderLabel"
      :render-prefix="renderPrefix"
      class="tree-content"
    />
  </div>
</template>

<script setup>
import { computed, h } from 'vue'
import { NTree, NIcon } from 'naive-ui'
import { 
  FolderOpenOutline, 
  DocumentTextOutline, 
  ChevronForwardOutline 
} from '@vicons/ionicons5'

const props = defineProps({
  files: {
    type: Array, // Array of { path: string, type: 'file'|'dir' }
    required: true
  }
})

const treeData = computed(() => {
  const root = []
  const map = {}

  props.files.forEach(file => {
    const parts = file.path.split(/[/\\]/)
    let current = root
    let fullPath = ''

    parts.forEach((part, index) => {
      fullPath += (fullPath ? '/' : '') + part
      const isLast = index === parts.length - 1
      const key = fullPath

      if (!map[key]) {
        const node = {
          label: part,
          key: key,
          type: isLast ? file.type : 'dir',
          children: isLast && file.type === 'file' ? undefined : []
        }
        map[key] = node
        current.push(node)
      }
      current = map[key].children
    })
  })

  // Sort: dirs first, then alphabetical
  const sortTree = (nodes) => {
    if (!nodes) return
    nodes.sort((a, b) => {
      if (a.type !== b.type) return a.type === 'dir' ? -1 : 1
      return a.label.localeCompare(b.label)
    })
    nodes.forEach(n => sortTree(n.children))
  }
  
  sortTree(root)
  return root
})

const renderPrefix = ({ option }) => {
  return h(NIcon, null, {
    default: () => h(option.type === 'dir' ? FolderOpenOutline : DocumentTextOutline)
  })
}

const renderLabel = ({ option }) => {
  return h('span', { class: 'tree-node-label' }, option.label)
}
</script>

<style scoped>
.file-tree-container {
  border: 1px solid #eee;
  border-radius: 4px;
  background: #fff;
  max-height: 400px;
  display: flex;
  flex-direction: column;
}

.tree-header {
  padding: 8px 12px;
  background: #f6f8fa;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  font-size: 13px;
}

.tree-content {
  padding: 8px;
  overflow-y: auto;
  flex: 1;
}

:deep(.tree-node-label) {
  font-size: 12px;
  font-family: 'Fira Code', monospace;
}
</style>
