<template>
  <div
    v-if="show"
    ref="modalRef"
    class="script-docs-float"
    :class="{ minimized: isMinimized }"
    :style="{
      left: position.x + 'px',
      top: position.y + 'px',
      width: isMinimized ? '300px' : size.width + 'px',
      height: isMinimized ? '48px' : size.height + 'px',
    }"
  >
    <!-- Header (Draggable) -->
    <div
      class="docs-header"
      @mousedown="startDrag"
    >
      <div class="header-left">
        <n-icon size="18" class="icon-blue">
          <CodeSlashOutline />
        </n-icon>
        <span class="header-title">Script API Docs</span>
      </div>
      <div class="header-right" @mousedown.stop>
        <!-- Search Button (only when not minimized) -->
        <n-popover v-if="!isMinimized" trigger="click" placement="bottom-end">
          <template #trigger>
            <n-button quaternary circle size="tiny">
              <template #icon>
                <n-icon><SearchOutline /></n-icon>
              </template>
            </n-button>
          </template>
          <n-input 
            v-model:value="searchQuery" 
            placeholder="Search API..." 
            size="small" 
            class="search-input" 
            autofocus
            clearable
          />
        </n-popover>

        <!-- Minimize/Restore -->
        <n-button quaternary circle size="tiny" @click="toggleMinimize">
          <template #icon>
            <n-icon>
              <RemoveOutline v-if="!isMinimized" />
              <ExpandOutline v-else />
            </n-icon>
          </template>
        </n-button>

        <!-- Maximize/Restore Size -->
        <n-button quaternary circle size="tiny" @click="toggleMaximize" v-if="!isMinimized">
          <template #icon>
            <n-icon>
              <ResizeOutline v-if="!isMaximized" />
              <ContractOutline v-else />
            </n-icon>
          </template>
        </n-button>

        <!-- Close -->
        <n-button quaternary circle size="tiny" type="error" @click="$emit('close')">
          <template #icon>
            <n-icon><CloseOutline /></n-icon>
          </template>
        </n-button>
      </div>
    </div>

    <!-- Content -->
    <div v-show="!isMinimized" class="docs-body">
      <!-- Sidebar -->
      <div class="docs-sidebar">
        <ul class="module-list">
          <li v-for="mod in filteredModules" :key="mod.name">
            <a
              href="#"
              class="module-link"
              :class="{ active: activeModule === mod.name }"
              @click.prevent="scrollToModule(mod.name)"
            >
              {{ mod.name }}
            </a>
          </li>
        </ul>
      </div>

      <!-- Main Content -->
      <div id="docs-content" class="docs-main">
        <div v-if="loading" class="loading-state">
          <n-spin size="medium" />
        </div>
        <div v-else-if="error" class="error-state">
          {{ error }}
          <n-button size="small" class="retry-btn" @click="refreshDocs">Retry</n-button>
        </div>
        <div v-else class="content-list">
          <div
            v-for="mod in filteredModules"
            :key="mod.name"
            :id="`float-module-${mod.name}`"
            class="module-section"
          >
            <h3 class="module-title">
              <span class="hash">#</span>{{ mod.name }}
            </h3>
            <p class="module-desc">{{ mod.desc }}</p>
            
            <div class="functions-list">
              <div v-for="fn in mod.functions" :key="fn.name" class="function-item">
                <div class="function-header">
                  <code class="function-name">
                    {{ mod.name }}.{{ fn.name }}
                  </code>
                  <span class="function-args">
                    ({{ fn.params.map(p => p.name).join(', ') }})
                  </span>
                </div>
                <p class="function-desc">{{ fn.desc }}</p>
                
                <div class="function-details">
                  <div v-if="fn.params.length > 0" class="params-list">
                    <div v-for="param in fn.params" :key="param.name" class="param-row">
                      <div class="param-name" :title="param.name">{{ param.name }}</div>
                      <div class="param-info">
                        <span class="param-type">{{ param.type }}</span>
                        {{ param.desc }}
                      </div>
                    </div>
                  </div>
                  <div class="return-row">
                    <div class="return-label">Returns</div>
                    <div class="return-value">{{ fn.returns }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div v-if="filteredModules.length === 0" class="empty-state">
            No results found for "{{ searchQuery }}"
          </div>
        </div>
      </div>
    </div>

    <!-- Resize Handle (Bottom Right) -->
    <div
      v-if="!isMinimized && !isMaximized"
      class="resize-handle"
      @mousedown="startResize"
    >
      <div class="resize-icon"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import { NIcon, NButton, NInput, NPopover, NSpin } from "naive-ui";
import { 
  CodeSlashOutline, 
  CloseOutline, 
  RemoveOutline, 
  ExpandOutline,
  ResizeOutline,
  ContractOutline,
  SearchOutline 
} from "@vicons/ionicons5";
import { GetScriptAPIDocs } from '../../wailsjs/go/main/App';

const props = defineProps({
  show: Boolean
});

const emit = defineEmits(['close']);

// State
const modules = ref([]);
const loading = ref(true);
const error = ref(null);
const activeModule = ref('');
const searchQuery = ref('');

// Window State
const position = ref({ x: window.innerWidth - 620, y: 100 });
const size = ref({ width: 600, height: 500 });
const isMinimized = ref(false);
const isMaximized = ref(false);
const preMaximizeSize = ref({ ...size.value });
const preMaximizePos = ref({ ...position.value });

// Dragging
const modalRef = ref(null);
const isDragging = ref(false);
const dragOffset = ref({ x: 0, y: 0 });

// Resizing
const isResizing = ref(false);
const resizeStart = ref({ x: 0, y: 0 });
const initialSize = ref({ width: 0, height: 0 });

// Filtering
const filteredModules = computed(() => {
  if (!searchQuery.value) return modules.value;
  const query = searchQuery.value.toLowerCase();
  
  return modules.value.map(mod => {
    // Check if module matches
    const modMatch = mod.name.toLowerCase().includes(query) || 
                     mod.desc.toLowerCase().includes(query);
    
    // Check functions
    const matchedFunctions = mod.functions.filter(fn => 
      fn.name.toLowerCase().includes(query) || 
      fn.desc.toLowerCase().includes(query)
    );
    
    if (modMatch || matchedFunctions.length > 0) {
      return {
        ...mod,
        functions: matchedFunctions.length > 0 ? matchedFunctions : mod.functions
      };
    }
    return null;
  }).filter(Boolean);
});

// Actions
const refreshDocs = async () => {
  loading.value = true;
  error.value = null;
  try {
    const res = await GetScriptAPIDocs();
    if (res.code === 200) {
      modules.value = res.data;
      if (modules.value.length > 0) {
        activeModule.value = modules.value[0].name;
      }
    } else {
      error.value = res.msg;
    }
  } catch (err) {
    error.value = "Failed to load documentation: " + err;
  } finally {
    loading.value = false;
  }
};

const scrollToModule = (name) => {
  activeModule.value = name;
  const el = document.getElementById(`float-module-${name}`);
  const container = document.getElementById('docs-content');
  if (el && container) {
    el.scrollIntoView({ behavior: 'smooth' });
  }
};

// Window Controls
const toggleMinimize = () => {
  isMinimized.value = !isMinimized.value;
};

const toggleMaximize = () => {
  if (isMaximized.value) {
    // Restore
    size.value = { ...preMaximizeSize.value };
    position.value = { ...preMaximizePos.value };
    isMaximized.value = false;
  } else {
    // Maximize
    preMaximizeSize.value = { ...size.value };
    preMaximizePos.value = { ...position.value };
    position.value = { x: 20, y: 20 };
    size.value = { width: window.innerWidth - 40, height: window.innerHeight - 40 };
    isMaximized.value = true;
  }
};

// Drag Logic
const startDrag = (e) => {
  if (isMaximized.value) return;
  isDragging.value = true;
  dragOffset.value = {
    x: e.clientX - position.value.x,
    y: e.clientY - position.value.y
  };
  document.addEventListener('mousemove', onDrag);
  document.addEventListener('mouseup', stopDrag);
};

const onDrag = (e) => {
  if (!isDragging.value) return;
  position.value = {
    x: e.clientX - dragOffset.value.x,
    y: e.clientY - dragOffset.value.y
  };
};

const stopDrag = () => {
  isDragging.value = false;
  document.removeEventListener('mousemove', onDrag);
  document.removeEventListener('mouseup', stopDrag);
};

// Resize Logic
const startResize = (e) => {
  isResizing.value = true;
  resizeStart.value = { x: e.clientX, y: e.clientY };
  initialSize.value = { ...size.value };
  document.addEventListener('mousemove', onResize);
  document.addEventListener('mouseup', stopResize);
};

const onResize = (e) => {
  if (!isResizing.value) return;
  const dx = e.clientX - resizeStart.value.x;
  const dy = e.clientY - resizeStart.value.y;
  
  size.value = {
    width: Math.max(300, initialSize.value.width + dx),
    height: Math.max(200, initialSize.value.height + dy)
  };
};

const stopResize = () => {
  isResizing.value = false;
  document.removeEventListener('mousemove', onResize);
  document.removeEventListener('mouseup', stopResize);
};

onMounted(() => {
  refreshDocs();
});

watch(() => props.show, (newVal) => {
  if (newVal && modules.value.length === 0) {
    refreshDocs();
  }
});
</script>

<style scoped>
.script-docs-float {
  position: fixed;
  z-index: 1000;
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2), 0 0 0 1px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  transition: box-shadow 0.2s;
  overflow: hidden;
  color: #333;
}

.script-docs-float.minimized {
  overflow: hidden;
}

/* Header */
.docs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background-color: #f3f4f6;
  border-bottom: 1px solid #e5e7eb;
  cursor: move;
  user-select: none;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-blue {
  color: #2563eb;
}

.header-title {
  font-weight: 700;
  color: #374151;
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-input {
  width: 180px;
}

/* Body */
.docs-body {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* Sidebar */
.docs-sidebar {
  width: 25%;
  min-width: 120px;
  background-color: #f9fafb;
  border-right: 1px solid #e5e7eb;
  overflow-y: auto;
}

.module-list {
  padding: 8px 0;
  list-style: none;
  margin: 0;
}

.module-link {
  display: block;
  padding: 8px 12px;
  font-size: 13px;
  font-weight: 500;
  color: #4b5563;
  text-decoration: none;
  transition: all 0.2s;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.module-link:hover {
  background-color: #eff6ff;
  color: #2563eb;
}

.module-link.active {
  background-color: #dbeafe;
  color: #1d4ed8;
  border-right: 3px solid #2563eb;
}

/* Main Content */
.docs-main {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background-color: white;
  scroll-behavior: smooth;
}

.loading-state, .error-state, .empty-state {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #6b7280;
  text-align: center;
}

.error-state {
  color: #ef4444;
}

.retry-btn {
  margin-top: 8px;
}

/* Module Section */
.module-section {
  scroll-margin-top: 16px;
  margin-bottom: 32px;
}

.module-title {
  font-size: 18px;
  font-weight: 700;
  color: #1f2937;
  display: flex;
  align-items: center;
  border-bottom: 1px solid #e5e7eb;
  padding-bottom: 8px;
  margin-bottom: 12px;
  position: sticky;
  top: 0;
  background-color: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(4px);
  z-index: 10;
}

.hash {
  color: #2563eb;
  margin-right: 6px;
}

.module-desc {
  font-size: 14px;
  color: #6b7280;
  margin-bottom: 16px;
}

/* Functions */
.functions-list {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.function-item {
  /* group equivalent */
}

.function-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin-bottom: 4px;
  flex-wrap: wrap;
}

.function-name {
  font-size: 14px;
  font-weight: 700;
  color: #7e22ce;
  background-color: #f3e8ff;
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid #f3e8ff;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

.function-args {
  font-size: 12px;
  color: #9ca3af;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

.function-desc {
  font-size: 13px;
  color: #4b5563;
  margin-bottom: 8px;
  padding-left: 4px;
}

.function-details {
  padding-left: 12px;
  border-left: 2px solid #f3f4f6;
  margin-left: 4px;
}

.params-list {
  margin-bottom: 8px;
}

.param-row {
  font-size: 12px;
  display: grid;
  grid-template-columns: 100px 1fr;
  gap: 8px;
  margin-bottom: 4px;
}

.param-name {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  color: #6b7280;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.param-info {
  color: #4b5563;
}

.param-type {
  color: #16a34a;
  background-color: #dcfce7;
  padding: 0 4px;
  border-radius: 2px;
  font-size: 10px;
  margin-right: 4px;
}

.return-row {
  font-size: 12px;
  display: grid;
  grid-template-columns: 100px 1fr;
  gap: 8px;
}

.return-label {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  color: #9ca3af;
}

.return-value {
  color: #ea580c;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}

/* Resize Handle */
.resize-handle {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 16px;
  height: 16px;
  cursor: nwse-resize;
  z-index: 50;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.5;
}

.resize-handle:hover {
  opacity: 1;
}

.resize-icon {
  width: 8px;
  height: 8px;
  border-right: 2px solid #9ca3af;
  border-bottom: 2px solid #9ca3af;
}

/* Custom Scrollbar */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}
::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
</style>
