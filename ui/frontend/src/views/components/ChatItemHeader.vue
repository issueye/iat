<template>
  <div class="message-header">
    <span>{{ header }}</span>
    <span class="message-time">{{ formatTime(item.createdAt) }}</span>
  </div>
</template>
<script setup>
import { computed } from "vue";
import { RoleTypes } from "@/constants/chat";

const props = defineProps({
  item: {
    type: Object,
    default: () => ({}),
  },
});

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
  gap: var(--base-gap-sm);
}

.message-time {
  margin-left: var(--base-gap-sm);
  font-size: var(--base-font-size-sm);
  color: var(--color-grey-text);
}
</style>
