<template>
  <div class="workflow-canvas" ref="canvasRef">
    <div v-if="tasks.length === 0" class="empty-state">等待生成工作流...</div>
    <div v-else class="dag-container">
      <!-- Simple SVG DAG for now -->
      <svg width="100%" height="100%" class="dag-svg">
        <defs>
          <marker
            id="arrowhead"
            markerWidth="10"
            markerHeight="7"
            refX="0"
            refY="3.5"
            orient="auto"
          >
            <polygon points="0 0, 10 3.5, 0 7" fill="#999" />
          </marker>
        </defs>

        <!-- Edges -->
        <g v-for="edge in edges" :key="edge.id">
          <path
            :d="edge.path"
            stroke="#999"
            stroke-width="2"
            fill="none"
            marker-end="url(#arrowhead)"
          />
        </g>

        <!-- Nodes -->
        <foreignObject
          v-for="node in nodes"
          :key="node.id"
          :x="node.x"
          :y="node.y"
          width="180"
          height="80"
        >
          <div
            class="task-node"
            :class="node.status"
            @click="handleSelectTask(node.id)"
          >
            <div class="node-title">{{ node.title }}</div>
            <div class="node-capability">{{ node.capability }}</div>
            <div class="node-status-icon">
              <n-spin v-if="node.status === 'running'" :size="12" />
              <n-icon v-else-if="node.status === 'completed'" color="#52c41a"
                ><CheckmarkCircleOutline
              /></n-icon>
              <n-icon v-else-if="node.status === 'failed'" color="#f5222d"
                ><CloseCircleOutline
              /></n-icon>
            </div>
          </div>
        </foreignObject>
      </svg>
    </div>

    <!-- Task Detail Drawer -->
    <n-drawer v-model:show="showDrawer" :width="400" placement="right">
      <n-drawer-content :title="selectedTask?.title || '任务详情'" closable>
        <n-descriptions label-placement="top" :column="1">
          <n-descriptions-item label="状态">
            <n-tag :type="getStatusType(selectedTask?.status)">
              {{ selectedTask?.status }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="能力需求">
            <code>{{ selectedTask?.capability }}</code>
          </n-descriptions-item>
          <n-descriptions-item label="描述">
            {{ selectedTask?.description }}
          </n-descriptions-item>
          <n-descriptions-item v-if="selectedTask?.error" label="错误信息">
            <n-alert type="error" :show-icon="false">
              {{ selectedTask.error }}
            </n-alert>
          </n-descriptions-item>
          <n-descriptions-item label="执行输出">
            <ResultRenderer
              v-if="selectedTask?.output"
              :content="selectedTask.output"
              :type="getRendererType(selectedTask)"
              :metadata="getRendererMetadata(selectedTask)"
            />
            <div v-else class="empty-output">暂无输出</div>
          </n-descriptions-item>
        </n-descriptions>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, watch } from "vue";
import {
  NSpin,
  NIcon,
  NDrawer,
  NDrawerContent,
  NDescriptions,
  NDescriptionsItem,
  NTag,
  NAlert,
} from "naive-ui";
import {
  CheckmarkCircleOutline,
  CloseCircleOutline,
  PlayCircleOutline,
} from "@vicons/ionicons5";
import ResultRenderer from "../renderers/ResultRenderer.vue";

const props = defineProps({
  tasks: {
    type: Array,
    default: () => [],
  },
});

const canvasRef = ref(null);
const showDrawer = ref(false);
const selectedTaskId = ref(null);

const selectedTask = computed(() => {
  return props.tasks.find((t) => t.id === selectedTaskId.value);
});

const handleSelectTask = (id) => {
  selectedTaskId.value = id;
  showDrawer.value = true;
};

const getStatusType = (status) => {
  switch (status) {
    case "completed":
      return "success";
    case "running":
      return "info";
    case "failed":
      return "error";
    default:
      return "default";
  }
};

const getRendererType = (task) => {
  if (!task.output) return "text";
  if (task.capability === "list_files") return "tree";
  if (task.capability === "diff_file") return "diff";
  try {
    JSON.parse(task.output);
    return "code";
  } catch (e) {
    return "text";
  }
};

const getRendererMetadata = (task) => {
  if (task.capability === "diff_file") return { language: "diff" };
  try {
    JSON.parse(task.output);
    return { language: "json" };
  } catch (e) {}
  return {};
};

// Simple layout algorithm
const nodes = computed(() => {
  const layout = [];
  const levels = {}; // taskID -> level

  // 1. Calculate levels (simple BFS)
  const taskMap = {};
  props.tasks.forEach((t) => {
    taskMap[t.id] = t;
    levels[t.id] = 0;
  });

  let changed = true;
  while (changed) {
    changed = false;
    props.tasks.forEach((t) => {
      if (t.dependsOn) {
        t.dependsOn.forEach((depId) => {
          if (levels[t.id] <= levels[depId]) {
            levels[t.id] = levels[depId] + 1;
            changed = true;
          }
        });
      }
    });
  }

  // 2. Position nodes
  const levelCounts = {};
  props.tasks.forEach((t) => {
    const lvl = levels[t.id];
    levelCounts[lvl] = (levelCounts[lvl] || 0) + 1;

    layout.push({
      id: t.id,
      title: t.title,
      capability: t.capability,
      status: t.status || "pending",
      level: lvl,
      order: levelCounts[lvl] - 1,
      x: lvl * 250 + 20,
      y: (levelCounts[lvl] - 1) * 120 + 20,
    });
  });

  return layout;
});

const edges = computed(() => {
  const result = [];
  const nodeMap = {};
  nodes.value.forEach((n) => (nodeMap[n.id] = n));

  props.tasks.forEach((t) => {
    if (t.dependsOn) {
      t.dependsOn.forEach((depId) => {
        const from = nodeMap[depId];
        const to = nodeMap[t.id];
        if (from && to) {
          const x1 = from.x + 180;
          const y1 = from.y + 40;
          const x2 = to.x;
          const y2 = to.y + 40;
          result.push({
            id: `${depId}-${t.id}`,
            path: `M ${x1} ${y1} C ${x1 + 50} ${y1}, ${x2 - 50} ${y2}, ${x2} ${y2}`,
          });
        }
      });
    }
  });
  return result;
});
</script>

<style scoped>
.workflow-canvas {
  width: 100%;
  height: 300px;
  background: #f9f9f9;
  border: 1px solid #eee;
  border-radius: 8px;
  overflow: auto;
  position: relative;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #999;
}

.dag-container {
  min-width: 1000px;
  min-height: 500px;
}

.task-node {
  background: #fff;
  border: 2px solid #ddd;
  border-radius: 8px;
  padding: 10px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  cursor: pointer;
  transition: all 0.3s;
  height: 60px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  position: relative;
}

.task-node:hover {
  border-color: #1890ff;
  transform: translateY(-2px);
}

.task-node.running {
  border-color: #1890ff;
  background: #e6f7ff;
}

.task-node.completed {
  border-color: #52c41a;
  background: #f6ffed;
}

.task-node.failed {
  border-color: #f5222d;
  background: #fff1f0;
}

.node-title {
  font-weight: bold;
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.node-capability {
  font-size: 11px;
  color: #666;
  margin-top: 4px;
}

.node-status-icon {
  position: absolute;
  top: 5px;
  right: 5px;
}

.empty-output {
  color: #999;
  font-style: italic;
  font-size: 12px;
  padding: 10px;
  background: #f9f9f9;
  border-radius: 4px;
}

.dag-svg {
  pointer-events: none;
}

foreignObject {
  pointer-events: auto;
}
</style>
