<template>
  <div class="message-footer">
    <n-button text color="#8a2be2" @click="showPromptModalFn(item)">
      {{ item.role === "tool" ? "输出" : "输入" }}
    </n-button>
    <n-button
      text
      size="tiny"
      style="margin-left: 8px"
      @click="showDebugDrawer = true"
    >
      调试
    </n-button>

    <!-- token -->
    <span class="token-usage" v-if="item.role !== 'tool'">
      TOKEN {{ item.tokenCount }}
    </span>
  </div>
  <!-- Debug Console Drawer -->
  <n-drawer v-model:show="showDebugDrawer" :width="500" placement="right">
    <n-drawer-content title="系统调试日志" closable>
      <template #header-extra>
        <n-button size="tiny" secondary @click="debugStore.clear()">
          清空
        </n-button>
      </template>
      <div class="debug-logs">
        <div v-for="log in logs" :key="log.id" class="log-item">
          <div class="log-meta">
            <span class="log-time">{{ log.time }}</span>
            <n-tag size="small" :bordered="false" type="info">{{
              log.type
            }}</n-tag>
          </div>
          <pre class="log-data">{{ log.data }}</pre>
        </div>
        <div v-if="logs.length === 0" class="empty-logs">暂无实时日志</div>
      </div>
    </n-drawer-content>
  </n-drawer>
  <n-modal
    v-model:show="showPromptModal"
    preset="dialog"
    :title="
      viewType === 'diff'
        ? '文件差异'
        : viewType === 'tree'
          ? '文件目录'
          : '详细信息'
    "
    style="width: 80%"
  >
    <div style="max-height: 80vh; overflow: auto">
      <ResultRenderer
        :content="currentViewPrompt"
        :type="viewType"
        :metadata="viewMetadata"
      />
    </div>
  </n-modal>
</template>
<script setup>
import { ref, computed } from "vue";
import { useDebugStore } from "@/store/debug";
import ResultRenderer from "@/components/renderers/ResultRenderer.vue";

defineProps({
  item: {
    type: Object,
    default: () => ({}),
  },
});

const showDebugDrawer = ref(false);
const showPromptModal = ref(false);
const debugStore = useDebugStore();
const logs = computed(() => debugStore.logs);
const viewType = ref("text");
const viewMetadata = ref({});
const currentViewPrompt = ref("");

const showPromptModalFn = (item) => {
  console.log("showPromptModalFn", item);
  viewType.value = "text";
  viewMetadata.value = {};

  switch (item.role) {
    case "tool":
      {
        currentViewPrompt.value = item.toolOutput || "";
        const toolName = item.toolName;

        if (toolName === "list_files") {
          viewType.value = "tree";
        } else if (toolName === "diff_file") {
          viewType.value = "diff";
          viewMetadata.value = { language: "diff" };
        } else if (toolName === "read_file" || toolName === "read_file_range") {
          viewType.value = "code";
          // Try to get language from path in arguments
          try {
            const args = JSON.parse(item.toolArguments || "{}");
            const path = args.path || "";
            const ext = path.split(".").pop();
            viewMetadata.value = { path, language: ext };
          } catch (e) {}
        } else {
          // Check if it's JSON
          try {
            JSON.parse(currentViewPrompt.value);
            viewType.value = "code";
            viewMetadata.value = { language: "json" };
          } catch (e) {}
        }
      }
      break;
    default:
      {
        currentViewPrompt.value = item.prompt || "";
        if (currentViewPrompt.value) {
          try {
            const parsed = JSON.parse(currentViewPrompt.value);
            currentViewPrompt.value = JSON.stringify(parsed, null, 2);
            viewType.value = "code";
            viewMetadata.value = { language: "json" };
          } catch (e) {}
        }
      }
      break;
  }

  showPromptModal.value = true;
};
</script>
<style scoped>
.debug-logs {
  font-family: "Fira Code", monospace;
  font-size: var(--base-font-size-sm);
}
.log-item {
  margin-bottom: var(--base-gap);
  border-bottom: 1px solid var(--color-grey);
  padding-bottom: var(--base-gap);
}
.log-meta {
  display: flex;
  align-items: center;
  gap: var(--base-gap);
  margin-bottom: var(--base-gap-sm);
}
.log-time {
  color: var(--color-grey-text);
}
.log-data {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  color: var(--color-text);
  background: var(--color-grey-light);
  padding: var(--base-padding-sm);
  border-radius: var(--base-radius-sm);
}
.empty-logs {
  text-align: center;
  color: var(--color-grey-text);
  padding: var(--base-padding) 0;
}

.message-footer {
  display: flex;
  gap: var(--base-gap);
}

.token-usage {
  font-size: var(--base-font-size-sm);
  color: var(--color-grey-text);
}
</style>
