import { ref, onMounted, onUnmounted, computed } from "vue";
import { api } from "./api";

export function useEngineStatus() {
  const engineStatus = ref("Checking");

  const engineStatusLabel = computed(() => {
    if (engineStatus.value === "Online") return "Online";
    if (engineStatus.value === "Offline") return "Offline";
    if (engineStatus.value === "Error") return "Error";
    return "Checking";
  });

  async function refreshEngineStatus() {
    try {
      const res = await api.checkHealth();
      engineStatus.value = res ? "Online" : "Error";
    } catch (e) {
      engineStatus.value = "Offline";
    }
  }

  let timer: number | null = null;

  onMounted(() => {
    refreshEngineStatus();
    timer = window.setInterval(refreshEngineStatus, 10_000);
  });

  onUnmounted(() => {
    if (timer) window.clearInterval(timer);
    timer = null;
  });

  return {
    engineStatus,
    engineStatusLabel,
    refreshEngineStatus,
  };
}

