<script setup>
import { ref, computed, watch, onMounted } from "vue";
import { NTag, NCheckbox, NButton, NIcon, NInput, NEmpty, NScrollbar, useMessage } from "naive-ui";
import { TrashOutline, AddOutline, RefreshOutline as RefreshIcon } from "@vicons/ionicons5";
import { api } from "../api";

const props = defineProps({
  sessionId: { type: Number, required: true },
});

const message = useMessage();
const tasks = ref([]);
const loading = ref(false);
const newTaskContent = ref("");

async function fetchTasks() {
  if (!props.sessionId) return;
  loading.value = true;
  try {
    tasks.value = await api.listTasks(props.sessionId);
  } catch (e) {
    message.error("获取任务列表失败");
  } finally {
    loading.value = false;
  }
}

watch(() => props.sessionId, () => {
  fetchTasks();
});

onMounted(() => {
  fetchTasks();
});

async function handleAdd() {
  if (!newTaskContent.value.trim()) return;
  try {
    const task = await api.createTask(props.sessionId, newTaskContent.value, "medium");
    tasks.value.push(task);
    newTaskContent.value = "";
    message.success("添加成功");
  } catch (e) {
    message.error("添加失败");
  }
}

async function handleCheck(task, checked) {
  const newStatus = checked ? "completed" : "pending";
  // Optimistic update
  const oldStatus = task.status;
  task.status = newStatus;
  
  try {
    await api.updateTask(task.id, newStatus);
  } catch (e) {
    task.status = oldStatus; // Revert on error
    message.error("更新状态失败");
  }
}

async function handleDelete(id) {
  try {
    await api.deleteTask(id);
    tasks.value = tasks.value.filter(t => t.id !== id);
    message.success("删除成功");
  } catch (e) {
    message.error("删除失败");
  }
}

function priorityType(p) {
  switch (p) {
    case "high": return "error";
    case "medium": return "warning";
    case "low": return "info";
    default: return "default";
  }
}

const sortedTasks = computed(() => {
  return [...tasks.value].sort((a, b) => {
    if (a.status === "completed" && b.status !== "completed") return 1;
    if (a.status !== "completed" && b.status === "completed") return -1;
    return b.id - a.id; // Newest first
  });
});

const completedCount = computed(() => tasks.value.filter(t => t.status === 'completed').length);

// Expose refresh method for parent
defineExpose({ refresh: fetchTasks });
</script>

<template>
  <div class="task-list-container">
    <div class="header">
      <div style="font-weight: bold;">任务列表</div>
      <n-tag v-if="tasks.length > 0" size="small" round :bordered="false">
        {{ completedCount }} / {{ tasks.length }}
      </n-tag>
      <n-button text size="tiny" @click="fetchTasks" :loading="loading">
        <template #icon><n-icon><RefreshIcon /></n-icon></template>
      </n-button>
    </div>

    <div class="add-box">
      <n-input
        v-model:value="newTaskContent"
        placeholder="添加新任务..."
        size="small"
        @keyup.enter="handleAdd"
      >
        <template #suffix>
          <n-button text size="tiny" @click="handleAdd">
            <template #icon><n-icon><AddOutline /></n-icon></template>
          </n-button>
        </template>
      </n-input>
    </div>

    <n-scrollbar style="flex: 1">
      <div v-if="tasks.length === 0" class="empty-state">
        <n-empty description="暂无任务" size="small" />
      </div>
      <div v-else class="list">
        <div
          v-for="task in sortedTasks"
          :key="task.id"
          class="task-item"
          :class="{ completed: task.status === 'completed' }"
        >
          <n-checkbox
            :checked="task.status === 'completed'"
            @update:checked="(v) => handleCheck(task, v)"
            style="margin-right: 8px"
          />
          <div class="content">
            <div class="text">{{ task.content }}</div>
            <div class="meta">
              <n-tag size="tiny" :type="priorityType(task.priority)" :bordered="false" style="margin-right: 4px; padding: 0 4px; height: 16px; font-size: 10px;">
                {{ task.priority }}
              </n-tag>
              <n-button size="tiny" text type="error" @click="handleDelete(task.id)" class="del-btn">
                <template #icon><n-icon><TrashOutline /></n-icon></template>
              </n-button>
            </div>
          </div>
        </div>
      </div>
    </n-scrollbar>
  </div>
</template>

<style scoped>
/* Same styles as before, kept for brevity */
.task-list-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #fff;
  /* Border handled by parent */
}

.header {
  padding: 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #f0f0f0;
}

.add-box {
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.list {
  padding: 8px 0;
}

.task-item {
  display: flex;
  align-items: flex-start;
  padding: 8px 12px;
  transition: background-color 0.2s;
}

.task-item:hover {
  background-color: #f9f9f9;
}

.task-item:hover .del-btn {
  opacity: 1;
}

.del-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.content {
  flex: 1;
  min-width: 0;
}

.text {
  font-size: 13px;
  line-height: 1.4;
  margin-bottom: 4px;
  word-break: break-word;
}

.completed .text {
  text-decoration: line-through;
  color: #999;
}

.meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.empty-state {
  padding: 20px;
  display: flex;
  justify-content: center;
}
</style>

