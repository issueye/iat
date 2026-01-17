<script setup>
import { ref, computed } from "vue";
import { NTag, NCheckbox, NButton, NIcon, NInput, NEmpty, NScrollbar, NSpace } from "naive-ui";
import { TrashOutline, AddOutline } from "@vicons/ionicons5";

const props = defineProps({
  tasks: { type: Array, default: () => [] },
  loading: Boolean,
});

const emit = defineEmits(["update:status", "delete", "add"]);

const newTaskContent = ref("");

function handleAdd() {
  if (!newTaskContent.value.trim()) return;
  emit("add", newTaskContent.value);
  newTaskContent.value = "";
}

function handleCheck(task, checked) {
  const newStatus = checked ? "completed" : "pending";
  emit("update:status", task.id, newStatus);
}

function priorityType(p) {
  switch (p) {
    case "high":
      return "error";
    case "medium":
      return "warning";
    case "low":
      return "info";
    default:
      return "default";
  }
}

const sortedTasks = computed(() => {
  return [...props.tasks].sort((a, b) => {
    // Sort by status (pending first) then id
    if (a.status === "completed" && b.status !== "completed") return 1;
    if (a.status !== "completed" && b.status === "completed") return -1;
    return a.id - b.id;
  });
});

const completedCount = computed(() => props.tasks.filter(t => t.status === 'completed').length);
</script>

<template>
  <div class="task-list-container">
    <div class="header">
      <div style="font-weight: bold;">任务列表</div>
      <n-tag v-if="tasks.length > 0" size="small" round :bordered="false">
        {{ completedCount }} / {{ tasks.length }}
      </n-tag>
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
              <n-button size="tiny" text type="error" @click="$emit('delete', task.id)" class="del-btn">
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
.task-list-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #fff;
  border-left: 1px solid #eee;
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
