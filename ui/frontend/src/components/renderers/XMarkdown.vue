<template>
  <div class="markdown-body" v-html="html" @click="handleCopy"></div>
</template>

<script setup>
import { computed } from "vue";
import { renderMarkdown } from "../../utils/markdown";
import { useMessage } from "naive-ui";

const message = useMessage();
const props = defineProps({
  markdown: {
    type: String,
    default: "",
  },
});

const html = computed(() => renderMarkdown(props.markdown));

const handleCopy = (e) => {
  const target = e.target;
  if (target.classList.contains("copy-btn")) {
    const code = target.getAttribute("data-code");
    if (code) {
      navigator.clipboard.writeText(decodeURIComponent(code)).then(() => {
        message.success("代码已复制");
      });
    }
  }
};
</script>

<style scoped>
.markdown-body {
  font-size: 14px;
  line-height: 1.6;
  word-wrap: break-word;
  color: #24292e;
}

:deep(pre.hljs) {
  padding: 16px;
  border-radius: 8px;
  background: #1e1e1e;
  color: #d4d4d4;
  overflow-x: auto;
  margin: 16px 0;
  font-family: "Fira Code", "JetBrains Mono", monospace;
  font-size: 13px;
  line-height: 1.5;
}

:deep(.code-block-wrapper) {
  margin: 16px 0;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #e1e4e8;
  background: #f6f8fa;
}

:deep(.code-block-header) {
  background: #f6f8fa;
  padding: 6px 12px;
  font-size: 12px;
  color: #586069;
  border-bottom: 1px solid #e1e4e8;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

:deep(.copy-btn) {
  cursor: pointer;
  padding: 2px 8px;
  border-radius: 4px;
  background: #fff;
  border: 1px solid #d1d5da;
  font-size: 11px;
  transition: all 0.2s;
}

:deep(.copy-btn:hover) {
  background: #f3f4f6;
  border-color: #1b1f23;
}

:deep(code) {
  font-family: "Fira Code", "JetBrains Mono", monospace;
  padding: 0.2em 0.4em;
  background-color: rgba(27, 31, 23, 0.05);
  border-radius: 3px;
  font-size: 85%;
}

:deep(pre code) {
  padding: 0;
  background-color: transparent;
  font-size: inherit;
}

:deep(h1, h2, h3, h4, h5, h6) {
  margin-top: 24px;
  margin-bottom: 16px;
  font-weight: 600;
  line-height: 1.25;
}

:deep(h1) {
  font-size: 1.5em;
  border-bottom: 1px solid #eaecef;
  padding-bottom: 0.3em;
}
:deep(h2) {
  font-size: 1.25em;
  border-bottom: 1px solid #eaecef;
  padding-bottom: 0.3em;
}

:deep(ul, ol) {
  padding-left: 2em;
  margin-bottom: 16px;
}

:deep(blockquote) {
  padding: 0 1em;
  color: #6a737d;
  border-left: 0.25em solid #dfe2e5;
  margin: 0 0 16px 0;
}

:deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 16px;
}

:deep(table th, table td) {
  padding: 6px 13px;
  border: 1px solid #dfe2e5;
}

:deep(table tr:nth-child(2n)) {
  background-color: #f6f8fa;
}
</style>
