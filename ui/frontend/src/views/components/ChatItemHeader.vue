<template>
  <div class="message-header">
    <div class="header-content">
      <span>{{ header }}</span>
      <span class="message-time">{{ formatTime(item.createdAt) }}</span>
    </div>
    <n-button
      quaternary
      circle
      size="tiny"
      class="delete-btn"
      @click="$emit('delete')"
    >
      <template #icon>
        <n-icon><TrashOutline /></n-icon>
      </template>
    </n-button>
  </div>
</template>
<script setup>
import { computed } from "vue";
import { NButton, NIcon } from "naive-ui";
import { TrashOutline } from "@vicons/ionicons5";
import { RoleTypes } from "@/constants/chat";

const props = defineProps({
  item: {
    type: Object,
    default: () => ({}),
  },
});

defineEmits(["delete"]);

const formatTime = (time) => {
  return new Date(time).toLocaleTimeString();
};

const header = computed(() => {
  const role = RoleTypes[props.item.role];
  return `${role.title}`;
});
</script>
<style scoped>
.message-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: var(--base-gap-sm);
}

.header-content {
  display: flex;
  align-items: center;
  gap: var(--base-gap-sm);
}

.message-time {
  margin-left: var(--base-gap-sm);
  font-size: var(--base-font-size-sm);
  color: var(--color-grey-text);
}

.delete-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.message-header:hover .delete-btn {
  opacity: 1;
}
</style>
