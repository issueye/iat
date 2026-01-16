<template>
  <div
    v-if="show"
    ref="modalRef"
    class="fixed z-50 bg-white rounded-lg shadow-2xl border border-gray-200 flex flex-col transition-all duration-200 ease-out"
    :style="{
      left: position.x + 'px',
      top: position.y + 'px',
      width: isMinimized ? '300px' : size.width + 'px',
      height: isMinimized ? '48px' : size.height + 'px',
    }"
  >
    <!-- Header (Draggable) -->
    <div
      class="flex justify-between items-center px-4 py-2 bg-gray-100 rounded-t-lg cursor-move select-none border-b border-gray-200"
      @mousedown="startDrag"
    >
      <div class="flex items-center gap-2">
        <n-icon size="18" class="text-blue-600">
          <CodeSlashOutline />
        </n-icon>
        <span class="font-bold text-gray-700 text-sm">Script API Docs</span>
      </div>
      <div class="flex items-center gap-2" @mousedown.stop>
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
            class="w-48" 
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
    <div v-show="!isMinimized" class="flex-1 overflow-hidden flex">
      <!-- Sidebar -->
      <div class="w-1/4 min-w-[120px] bg-gray-50 border-r border-gray-200 overflow-y-auto">
        <ul class="py-2">
          <li v-for="mod in filteredModules" :key="mod.name">
            <a
              href="#"
              class="block px-3 py-2 text-xs font-medium text-gray-600 hover:bg-blue-50 hover:text-blue-600 transition truncate"
              :class="{ 'bg-blue-100 text-blue-700 border-r-2 border-blue-600': activeModule === mod.name }"
              @click.prevent="scrollToModule(mod.name)"
            >
              {{ mod.name }}
            </a>
          </li>
        </ul>
      </div>

      <!-- Main Content -->
      <div id="docs-content" class="flex-1 overflow-y-auto p-4 bg-white scroll-smooth">
        <div v-if="loading" class="flex justify-center items-center h-full">
          <n-spin size="medium" />
        </div>
        <div v-else-if="error" class="text-red-500 p-4 text-center">
          {{ error }}
          <n-button size="small" class="mt-2" @click="refreshDocs">Retry</n-button>
        </div>
        <div v-else class="space-y-8">
          <div
            v-for="mod in filteredModules"
            :key="mod.name"
            :id="`float-module-${mod.name}`"
            class="scroll-mt-4"
          >
            <h3 class="text-lg font-bold text-gray-800 flex items-center border-b pb-2 mb-4 sticky top-0 bg-white/95 backdrop-blur z-10">
              <span class="text-blue-600 mr-2">#</span>{{ mod.name }}
            </h3>
            <p class="text-sm text-gray-500 mb-4">{{ mod.desc }}</p>
            
            <div class="space-y-6">
              <div v-for="fn in mod.functions" :key="fn.name" class="group">
                <div class="flex items-baseline gap-2 mb-1">
                  <code class="text-sm font-bold text-purple-700 bg-purple-50 px-1.5 py-0.5 rounded border border-purple-100">
                    {{ mod.name }}.{{ fn.name }}
                  </code>
                  <span class="text-xs text-gray-400 font-mono">
                    ({{ fn.params.map(p => p.name).join(', ') }})
                  </span>
                </div>
                <p class="text-xs text-gray-600 mb-2 pl-1">{{ fn.desc }}</p>
                
                <div class="pl-3 border-l-2 border-gray-100 ml-1">
                  <div v-if="fn.params.length > 0" class="mb-2">
                    <div v-for="param in fn.params" :key="param.name" class="text-xs grid grid-cols-[100px_1fr] gap-2 mb-1">
                      <div class="font-mono text-gray-500 truncate" :title="param.name">{{ param.name }}</div>
                      <div class="text-gray-600">
                        <span class="text-green-600 bg-green-50 px-1 rounded text-[10px] mr-1">{{ param.type }}</span>
                        {{ param.desc }}
                      </div>
                    </div>
                  </div>
                  <div class="text-xs grid grid-cols-[100px_1fr] gap-2">
                    <div class="font-mono text-gray-400">Returns</div>
                    <div class="text-orange-600 font-mono">{{ fn.returns }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div v-if="filteredModules.length === 0" class="text-center text-gray-400 py-8">
            No results found for "{{ searchQuery }}"
          </div>
        </div>
      </div>
    </div>

    <!-- Resize Handle (Bottom Right) -->
    <div
      v-if="!isMinimized && !isMaximized"
      class="absolute bottom-0 right-0 w-4 h-4 cursor-nwse-resize z-50 flex items-center justify-center opacity-50 hover:opacity-100"
      @mousedown="startResize"
    >
      <div class="w-2 h-2 border-r-2 border-b-2 border-gray-400"></div>
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
  ExpandOutline, // Use ExpandOutline for restore from minimize
  ResizeOutline, // Use ResizeOutline for maximize
  ContractOutline, // Use ContractOutline for restore from maximize
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
    // Simple scroll into view
    el.scrollIntoView({ behavior: 'smooth' });
  }
};

// Window Controls
const toggleMinimize = () => {
  isMinimized.value = !isMinimized.value;
  // Reset maximized if minimizing
  if (isMinimized.value) {
    // Keep width logic in style binding
  }
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
/* Custom Scrollbar for nicer look */
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
