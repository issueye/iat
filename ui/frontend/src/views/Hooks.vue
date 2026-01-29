<template>
  <div class="hooks-view">
    <div class="header-actions">
      <n-button type="primary" @click="showCreateModal = true">
        <template #icon>
          <n-icon><AddIcon /></n-icon>
        </template>
        添加 Hook
      </n-button>
    </div>

    <n-data-table
      :columns="columns"
      :data="hooks"
      :loading="loading"
      :pagination="pagination"
      class="hooks-table"
    />

    <!-- Create/Edit Modal -->
    <n-modal
      v-model:show="showCreateModal"
      preset="card"
      :title="editingHook ? '编辑 Hook' : '新建 Hook'"
      style="width: 600px"
    >
      <n-form
        ref="formRef"
        :model="formModel"
        :rules="rules"
        label-placement="left"
        label-width="100"
        require-mark-placement="right-hanging"
      >
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formModel.name" placeholder="请输入 Hook 名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input
            v-model:value="formModel.description"
            type="textarea"
            placeholder="请输入描述"
          />
        </n-form-item>
        <n-form-item label="触发类型" path="type">
          <n-select v-model:value="formModel.type" :options="typeOptions" />
        </n-form-item>
        <n-form-item label="目标类型" path="targetType">
          <n-select
            v-model:value="formModel.targetType"
            :options="targetTypeOptions"
          />
        </n-form-item>
        <n-form-item
          label="目标 ID"
          path="targetId"
          v-if="formModel.targetType === 'agent'"
        >
          <n-select
            v-model:value="formModel.targetId"
            :options="agentOptions"
            placeholder="选择智能体"
            filterable
          />
        </n-form-item>
        <n-form-item label="动作类型" path="action">
          <n-select v-model:value="formModel.action" :options="actionOptions" />
        </n-form-item>
        <n-form-item label="内容" path="content">
          <n-input
            v-model:value="formModel.content"
            type="textarea"
            :rows="5"
            placeholder="JS 脚本或 HTTP URL"
            font-family="monospace"
          />
        </n-form-item>
        <n-form-item label="启用" path="enabled">
          <n-switch v-model:value="formModel.enabled" />
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="closeModal">取消</n-button>
          <n-button type="primary" @click="handleSubmit">保存</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, h } from "vue";
import {
  NButton,
  NIcon,
  NDataTable,
  NModal,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NSwitch,
  NSpace,
  NTag,
  useMessage,
  useDialog,
} from "naive-ui";
import {
  AddOutline as AddIcon,
  CreateOutline as EditIcon,
  TrashOutline as DeleteIcon,
} from "@vicons/ionicons5";
import { api } from "../api";

const message = useMessage();
const dialog = useDialog();

const hooks = ref([]);
const agents = ref([]);
const loading = ref(false);
const showCreateModal = ref(false);
const editingHook = ref(null);

const pagination = { pageSize: 10 };

const formModel = ref({
  name: "",
  description: "",
  type: "pre_chat",
  targetType: "global",
  targetId: 0,
  action: "script",
  content: "",
  enabled: true,
});

const rules = {
  name: { required: true, message: "请输入名称", trigger: "blur" },
  type: { required: true, message: "请选择触发类型", trigger: "blur" },
  targetType: { required: true, message: "请选择目标类型", trigger: "blur" },
  action: { required: true, message: "请选择动作类型", trigger: "blur" },
  content: { required: true, message: "请输入内容", trigger: "blur" },
};

const typeOptions = [
  { label: "对话开始前 (pre_chat)", value: "pre_chat" },
  { label: "对话结束后 (post_chat)", value: "post_chat" },
  { label: "工具调用前 (pre_tool)", value: "pre_tool" },
  { label: "工具调用后 (post_tool)", value: "post_tool" },
];

const targetTypeOptions = [
  { label: "全局 (Global)", value: "global" },
  { label: "智能体 (Agent)", value: "agent" },
];

const actionOptions = [
  { label: "JS 脚本 (Script)", value: "script" },
  { label: "HTTP 请求 (Webhook)", value: "http" },
];

const agentOptions = computed(() => {
  return agents.value.map((a) => ({ label: a.name, value: a.id }));
});

const columns = [
  { title: "ID", key: "id", width: 60 },
  { title: "名称", key: "name", width: 150 },
  {
    title: "触发类型",
    key: "type",
    width: 150,
    render(row) {
      return h(NTag, { type: "info", size: "small" }, { default: () => row.type });
    },
  },
  {
    title: "目标",
    key: "targetType",
    width: 120,
    render(row) {
      if (row.targetType === "global") {
        return h(NTag, { type: "success", size: "small" }, { default: () => "Global" });
      }
      const agent = agents.value.find((a) => a.id === row.targetId);
      return h(
        NTag,
        { type: "warning", size: "small" },
        { default: () => (agent ? `Agent: ${agent.name}` : `Agent #${row.targetId}`) }
      );
    },
  },
  {
    title: "动作",
    key: "action",
    width: 100,
    render(row) {
      return h(
        NTag,
        { bordered: false, size: "small" },
        { default: () => (row.action === "script" ? "JS Script" : "HTTP") }
      );
    },
  },
  {
    title: "状态",
    key: "enabled",
    width: 80,
    render(row) {
      return h(
        NSwitch,
        {
          value: row.enabled,
          size: "small",
          onUpdateValue: (val) => toggleEnabled(row, val),
        }
      );
    },
  },
  {
    title: "操作",
    key: "actions",
    width: 120,
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(
            NButton,
            {
              size: "tiny",
              secondary: true,
              type: "info",
              onClick: () => editHook(row),
            },
            { icon: () => h(NIcon, null, { default: () => h(EditIcon) }) }
          ),
          h(
            NButton,
            {
              size: "tiny",
              secondary: true,
              type: "error",
              onClick: () => confirmDelete(row),
            },
            { icon: () => h(NIcon, null, { default: () => h(DeleteIcon) }) }
          ),
        ],
      });
    },
  },
];

async function fetchData() {
  loading.value = true;
  try {
    const [hData, aData] = await Promise.all([
      api.listHooks(),
      api.listAgents(),
    ]);
    hooks.value = hData || [];
    agents.value = aData || [];
  } catch (e) {
    message.error("加载数据失败: " + e.message);
  } finally {
    loading.value = false;
  }
}

async function toggleEnabled(row, val) {
  try {
    const updated = { ...row, enabled: val };
    await api.updateHook(updated);
    row.enabled = val;
    message.success(val ? "已启用" : "已禁用");
  } catch (e) {
    message.error("操作失败");
  }
}

function closeModal() {
  showCreateModal.value = false;
  editingHook.value = null;
  resetForm();
}

function resetForm() {
  formModel.value = {
    name: "",
    description: "",
    type: "pre_chat",
    targetType: "global",
    targetId: 0,
    action: "script",
    content: "",
    enabled: true,
  };
}

function editHook(row) {
  editingHook.value = row;
  formModel.value = { ...row };
  showCreateModal.value = true;
}

function confirmDelete(row) {
  dialog.warning({
    title: "确认删除",
    content: `确定要删除 Hook "${row.name}" 吗？`,
    positiveText: "删除",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await api.deleteHook(row.id);
        message.success("删除成功");
        fetchData();
      } catch (e) {
        message.error("删除失败: " + e.message);
      }
    },
  });
}

async function handleSubmit() {
  // Validate form manually or via ref if needed, simple check here
  if (!formModel.value.name || !formModel.value.content) {
    message.error("请填写必填项");
    return;
  }

  try {
    if (editingHook.value) {
      await api.updateHook(formModel.value);
      message.success("更新成功");
    } else {
      await api.createHook(formModel.value);
      message.success("创建成功");
    }
    closeModal();
    fetchData();
  } catch (e) {
    message.error("保存失败: " + e.message);
  }
}

onMounted(() => {
  fetchData();
});
</script>

<style scoped>
.hooks-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.header-actions {
  display: flex;
  justify-content: flex-end;
}
.hooks-table {
  flex: 1;
}
</style>
