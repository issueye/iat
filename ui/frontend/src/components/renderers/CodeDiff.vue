<template>
  <div class="code-diff-container">
    <div class="diff-header" v-if="fileName">
      <n-icon><FileTrayFullOutline /></n-icon>
      <span class="file-name">{{ fileName }}</span>
    </div>
    <div class="diff-content">
      <div 
        v-for="(line, index) in parsedLines" 
        :key="index"
        :class="['diff-line', line.type]"
      >
        <span class="line-num">{{ line.oldLine || '' }}</span>
        <span class="line-num">{{ line.newLine || '' }}</span>
        <span class="line-prefix">{{ line.prefix }}</span>
        <pre class="line-code"><code v-html="line.html"></code></pre>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { NIcon } from 'naive-ui'
import { FileTrayFullOutline } from '@vicons/ionicons5'
import hljs from 'highlight.js'

const props = defineProps({
  diff: {
    type: String,
    required: true
  },
  fileName: String,
  language: {
    type: String,
    default: 'javascript'
  }
})

const parsedLines = computed(() => {
  const lines = props.diff.split('\n')
  const result = []
  let oldLine = 0
  let newLine = 0

  lines.forEach(line => {
    if (line.startsWith('---') || line.startsWith('+++')) return
    
    if (line.startsWith('@@')) {
      const match = line.match(/@@ -(\d+),?\d* \+(\d+),?\d* @@/)
      if (match) {
        oldLine = parseInt(match[1])
        newLine = parseInt(match[2])
        result.push({ type: 'info', prefix: ' ', text: line, html: line })
      }
      return
    }

    let type = 'normal'
    let prefix = ' '
    let oNum = null
    let nNum = null

    if (line.startsWith('+')) {
      type = 'add'
      prefix = '+'
      nNum = newLine++
    } else if (line.startsWith('-')) {
      type = 'remove'
      prefix = '-'
      oNum = oldLine++
    } else {
      oNum = oldLine++
      nNum = newLine++
    }

    const content = line.slice(1)
    const highlighted = hljs.highlight(content, { language: props.language }).value

    result.push({
      type,
      prefix,
      oldLine: oNum,
      newLine: nNum,
      text: content,
      html: highlighted
    })
  })

  return result
})
</script>

<style scoped>
.code-diff-container {
  border: 1px solid #eee;
  border-radius: 4px;
  overflow: hidden;
  font-family: 'Fira Code', monospace;
  font-size: 12px;
  background: #fff;
}

.diff-header {
  background: #f6f8fa;
  padding: 8px 12px;
  border-bottom: 1px solid #eee;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.diff-content {
  overflow-x: auto;
}

.diff-line {
  display: flex;
  line-height: 1.5;
  white-space: pre;
}

.diff-line.add { background-color: #e6ffec; }
.diff-line.remove { background-color: #ffebe9; }
.diff-line.info { background-color: #f1f8ff; color: #005cc5; }

.line-num {
  width: 40px;
  text-align: right;
  padding-right: 10px;
  color: #999;
  user-select: none;
  border-right: 1px solid #eee;
}

.line-prefix {
  width: 20px;
  text-align: center;
  user-select: none;
}

.line-code {
  margin: 0;
  padding: 0 10px;
  flex: 1;
}

code {
  font-family: inherit;
}
</style>
