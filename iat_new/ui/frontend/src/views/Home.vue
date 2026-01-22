<template>
  <div>
    <h1>欢迎使用 iat</h1>
    <p>你的 AI Agent 工具</p>
    <p>Engine Status: {{ status }}</p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { api } from '../api';

const status = ref('Checking...');

onMounted(async () => {
  try {
    const res = await api.checkHealth();
    status.value = res ? 'Online' : 'Error';
  } catch (e) {
    console.error(e);
    status.value = 'Offline';
  }
});
</script>
