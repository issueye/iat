<template>
  <div class="result-renderer">
    <!-- Code Diff -->
    <CodeDiff 
      v-if="type === 'diff'" 
      :diff="content" 
      :file-name="metadata?.path"
      :language="metadata?.language"
    />

    <!-- File Tree -->
    <FileTree 
      v-else-if="type === 'tree'" 
      :files="parsedFiles" 
    />

    <!-- JSON / Code Block -->
    <div v-else-if="type === 'code'" class="code-block">
      <div class="code-header" v-if="metadata?.title">
        <span>{{ metadata.title }}</span>
      </div>
      <pre v-highlight><code :class="metadata?.language">{{ content }}</code></pre>
    </div>

    <!-- Default Markdown -->
    <div v-else class="markdown-body" v-html="renderedMarkdown"></div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import CodeDiff from './CodeDiff.vue'
import FileTree from './FileTree.vue'
import MarkdownIt from 'markdown-it'

const md = new MarkdownIt()

const props = defineProps({
  type: {
    type: String,
    default: 'text' // diff, tree, code, text
  },
  content: {
    type: [String, Object, Array],
    required: true
  },
  metadata: {
    type: Object,
    default: () => ({})
  }
})

const renderedMarkdown = computed(() => {
  if (typeof props.content !== 'string') return JSON.stringify(props.content, null, 2)
  return md.render(props.content)
})

const parsedFiles = computed(() => {
  if (Array.isArray(props.content)) return props.content
  if (typeof props.content === 'string') {
    try {
      return JSON.parse(props.content)
    } catch (e) {
      // Fallback for newline separated paths
      return props.content.split('\n').filter(p => p.trim()).map(p => ({
        path: p.trim(),
        type: p.endsWith('/') || p.endsWith('\\') ? 'dir' : 'file'
      }))
    }
  }
  return []
})
</script>

<style scoped>
.result-renderer {
  margin: 8px 0;
  width: 100%;
}

.code-block {
  border: 1px solid #eee;
  border-radius: 4px;
  overflow: hidden;
}

.code-header {
  background: #f6f8fa;
  padding: 4px 12px;
  font-size: 11px;
  color: #666;
  border-bottom: 1px solid #eee;
}

pre {
  margin: 0;
  padding: 12px;
  background: #fafafa;
  font-size: 12px;
  overflow-x: auto;
}

.markdown-body {
  font-size: 14px;
  line-height: 1.6;
}

:deep(.markdown-body p) {
  margin-bottom: 12px;
}
</style>
