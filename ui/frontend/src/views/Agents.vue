<template>
  <div style="padding: 20px">
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">智能体管理</n-h2>
      <n-button type="primary" @click="showCreateModal = true">
        新建智能体
      </n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="agents"
      :loading="loading"
      :pagination="pagination"
      :scroll-x="1000"
    />

    <!-- Create/Edit Modal -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      :title="isEdit ? '编辑智能体' : '新建智能体'"
      style="width: 700px"
    >
      <n-form
        ref="formRef"
        :model="formValue"
        :rules="rules"
        label-placement="left"
        label-width="120"
      >
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formValue.name" placeholder="智能体名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formValue.description" placeholder="描述" />
        </n-form-item>
        <n-form-item label="类型" path="type">
          <n-select
            v-model:value="formValue.type"
            :options="typeOptions"
            placeholder="选择类型"
          />
        </n-form-item>
        <n-form-item
          v-if="formValue.type !== 'external'"
          label="模型"
          path="modelId"
        >
          <n-select
            v-model:value="formValue.modelId"
            :options="modelOptions"
            placeholder="选择 AI 模型 (留空使用默认)"
            clearable
          />
        </n-form-item>
        <n-form-item
          v-if="formValue.type !== 'external'"
          label="关联工具"
          path="toolIds"
        >
          <n-select
            v-model:value="formValue.toolIds"
            multiple
            :options="toolOptions"
            placeholder="选择工具"
          />
        </n-form-item>
        <n-form-item
          v-if="formValue.type !== 'external'"
          label="关联 MCP 服务"
          path="mcpServerIds"
        >
          <n-select
            v-model:value="formValue.mcpServerIds"
            multiple
            :options="mcpOptions"
            placeholder="选择 MCP 服务"
          />
        </n-form-item>
        <n-form-item
          v-if="formValue.type !== 'external'"
          label="系统提示词"
          path="systemPrompt"
        >
          <n-input
            v-model:value="formValue.systemPrompt"
            type="textarea"
            :autosize="{ minRows: 5, maxRows: 10 }"
            placeholder="你是一个有用的助手..."
          />
        </n-form-item>
        <n-form-item
          v-if="formValue.type === 'external'"
          label="外部地址"
          path="externalUrl"
        >
          <n-input
            v-model:value="formValue.externalUrl"
            placeholder="例如 http://localhost:18080/a2a 或 /a2a/stream"
          />
        </n-form-item>
        <n-form-item
          v-if="formValue.type === 'external'"
          label="外部类型"
          path="externalType"
        >
          <n-select
            v-model:value="formValue.externalType"
            :options="externalTypeOptions"
            placeholder="选择外部 Agent 类型"
          />
        </n-form-item>
        <n-form-item
          v-if="formValue.type === 'external'"
          label="调用参数(JSON)"
          path="externalParams"
        >
          <n-input
            v-model:value="formValue.externalParams"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 8 }"
            placeholder='例如: {"headers":{"X-Env":"dev"},"query":{"projectId":"${projectId}"}}'
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="closeModal">取消</n-button>
          <n-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? "更新" : "创建" }}
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from "vue";
import { NButton, NSpace, useMessage, useDialog, NTag } from "naive-ui";
import {
  ListAgents,
  CreateAgent,
  UpdateAgent,
  DeleteAgent,
  ListAIModels,
  ListTools,
  ListMCPServers,
} from "../../wailsjs/go/main/App";

const message = useMessage();
const dialog = useDialog();

const agents = ref([]);
const modelOptions = ref([]);
const toolOptions = ref([]);
const mcpOptions = ref([]);
const loading = ref(false);
const showCreateModal = ref(false);
const submitting = ref(false);
const isEdit = ref(false);
const editingId = ref(null);

const formValue = ref({
  name: "",
  description: "",
  systemPrompt: "",
  type: "custom",
  modelId: null,
  toolIds: [],
  mcpServerIds: [],
  externalUrl: "",
  externalType: "",
  externalParams: "",
});

const rules = {
  name: { required: true, message: "必填", trigger: "blur" },
  type: { required: true, message: "必填", trigger: "blur" },
};

const pagination = { pageSize: 10 };

const typeOptions = [
  { label: "普通 Agent", value: "custom" },
  { label: "外部 Agent", value: "external" },
];

const externalTypeOptions = [
  { label: "http_a2a", value: "http_a2a" },
  { label: "http_sse", value: "http_sse" },
];

const columns = [
  { title: "名称", key: "name", width: 150 },
  {
    title: "状态",
    key: "status",
    width: 100,
    render(row) {
      let type = "default";
      if (row.status === "online") type = "success";
      if (row.status === "busy") type = "warning";
      return h(
        NTag,
        {
          type: type,
          bordered: false,
        },
        { default: () => row.status || "offline" },
      );
    },
  },
  {
    title: "类型",
    key: "type",
    width: 100,
    render(row) {
      return h(
        NTag,
        {
          type: row.type === "builtin" ? "primary" : "default",
          bordered: false,
        },
        { default: () => row.type },
      );
    },
  },
  {
    title: "描述",
    key: "description",
    key: "description",
    minWidth: 200,
    ellipsis: {
      tooltip: true,
    },
  },
  {
    title: "模型",
    key: "Model",
    width: 150,
    render(row) {
      return row.Model
        ? h(NTag, { type: "info" }, { default: () => row.Model.name })
        : "无";
    },
  },
  {
    title: "工具/MCP",
    key: "Tools",
    width: 200,
    render(row) {
      const tools = row.tools || [];
      const mcps = row.mcpServers || [];
      if (tools.length === 0 && mcps.length === 0) return "无";

      const tags = [];
      tools.forEach((t) => {
        tags.push(
          h(
            NTag,
            { type: "success", size: "small", bordered: false },
            { default: () => t.name },
          ),
        );
      });
      mcps.forEach((m) => {
        tags.push(
          h(
            NTag,
            { type: "warning", size: "small", bordered: false },
            { default: () => `MCP:${m.name}` },
          ),
        );
      });

      return h(
        NSpace,
        { size: 4, style: { flexWrap: "wrap" } },
        { default: () => tags },
      );
    },
  },
  {
    title: "操作",
    key: "actions",
    width: 150,
    fixed: "right",
    render(row) {
      return h(NSpace, null, {
        default: () => {
          const btns = [
            h(
              NButton,
              { size: "small", onClick: () => handleEdit(row) },
              { default: () => "编辑" },
            ),
          ];

          if (row.type !== "builtin") {
            btns.push(
              h(
                NButton,
                {
                  size: "small",
                  type: "error",
                  onClick: () => handleDelete(row),
                },
                { default: () => "删除" },
              ),
            );
          }
          return btns;
        },
      });
    },
  },
];

async function loadData() {
  loading.value = true;
  try {
    // Load Models
    const modelRes = await ListAIModels();
    if (modelRes.code === 200) {
      modelOptions.value = (modelRes.data || []).map((m) => ({
        label: m.name,
        value: m.id,
      }));
    }

    // Load Tools
    const toolRes = await ListTools();
    if (toolRes.code === 200) {
      toolOptions.value = (toolRes.data || []).map((t) => ({
        label: `${t.name} (${t.type})`,
        value: t.id,
      }));
    }

    // Load MCP Servers
    const mcpRes = await ListMCPServers();
    if (mcpRes.code === 200) {
      mcpOptions.value = (mcpRes.data || []).map((m) => ({
        label: `${m.name} (${m.enabled ? "已启用" : "已禁用"})`,
        value: m.id,
        disabled: !m.enabled,
      }));
    }

    // Load Agents
    const res = await ListAgents();
    if (res.code === 200) {
      agents.value = res.data || [];
    } else {
      message.error(res.msg);
    }
  } catch (e) {
    message.error("加载数据失败: " + e);
  } finally {
    loading.value = false;
  }
}

function handleEdit(row) {
  isEdit.value = true;
  editingId.value = row.id;
  formValue.value = {
    name: row.name,
    description: row.description,
    systemPrompt: row.systemPrompt,
    type: row.type || "custom",
    modelId: row.modelId,
    toolIds: (row.tools || []).map((t) => t.id),
    mcpServerIds: (row.mcpServers || []).map((m) => m.id),
    externalUrl: row.externalUrl || "",
    externalType: row.externalType || "",
    externalParams: row.externalParams || "",
  };
  showCreateModal.value = true;
}

function handleDelete(row) {
  dialog.warning({
    title: "确认删除",
    content: `确认删除智能体 "${row.name}"?`,
    positiveText: "确认",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        const res = await DeleteAgent(row.id);
        if (res.code === 200) {
          message.success("已删除");
          loadData();
        } else {
          message.error(res.msg);
        }
      } catch (e) {
        message.error("删除失败: " + e);
      }
    },
  });
}

function closeModal() {
  showCreateModal.value = false;
  formValue.value = {
    name: "",
    description: "",
    systemPrompt: "",
    type: "custom",
    modelId: null,
    toolIds: [],
    mcpServerIds: [],
    externalUrl: "",
    externalType: "",
    externalParams: "",
  };
  isEdit.value = false;
  editingId.value = null;
}

async function handleSubmit() {
  if (!formValue.value.name) {
    message.warning("名称为必填项");
    return;
  }

  submitting.value = true;
  try {
    let res;
    const modelId = formValue.value.modelId || 0;
    const type = formValue.value.type || "custom";

    if (isEdit.value) {
      res = await UpdateAgent(
        editingId.value,
        formValue.value.name,
        formValue.value.description,
        formValue.value.systemPrompt,
        type,
        formValue.value.externalUrl,
        formValue.value.externalType,
        formValue.value.externalParams,
        modelId,
        formValue.value.toolIds,
        formValue.value.mcpServerIds,
        1, // Default ModeID for now
        formValue.value.status || "offline",
        formValue.value.capabilities || "",
      );
    } else {
      res = await CreateAgent(
        formValue.value.name,
        formValue.value.description,
        formValue.value.systemPrompt,
        type,
        formValue.value.externalUrl,
        formValue.value.externalType,
        formValue.value.externalParams,
        modelId,
        formValue.value.toolIds,
        formValue.value.mcpServerIds,
        1,
        formValue.value.status || "offline",
        formValue.value.capabilities || "",
      );
    }

    if (res.code === 200) {
      message.success(isEdit.value ? "更新成功" : "创建成功");
      closeModal();
      loadData();
    } else {
      message.error(res.msg);
    }
  } catch (e) {
    message.error("操作失败: " + e);
  } finally {
    submitting.value = false;
  }
}

onMounted(() => {
  loadData();
});
</script>
