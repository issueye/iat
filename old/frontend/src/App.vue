<script setup>
import { NConfigProvider, zhCN, dateZhCN } from "naive-ui";
</script>

<template>
  <n-config-provider :locale="zhCN" :date-locale="dateZhCN">
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <router-view />
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<style scoped>
.n-config-provider {
  height: 100%;
}
</style>
