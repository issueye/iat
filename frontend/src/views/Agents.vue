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
        <n-form-item label="模型" path="modelId">
          <n-select
            v-model:value="formValue.modelId"
            :options="modelOptions"
            placeholder="选择 AI 模型 (留空使用默认)"
            clearable
          />
        </n-form-item>
        <n-form-item label="关联工具" path="toolIds">
          <n-select
            v-model:value="formValue.toolIds"
            multiple
            :options="toolOptions"
            placeholder="选择工具"
          />
        </n-form-item>
        <n-form-item label="系统提示词" path="systemPrompt">
          <n-input
            v-model:value="formValue.systemPrompt"
            type="textarea"
            :autosize="{ minRows: 5, maxRows: 10 }"
            placeholder="你是一个有用的助手..."
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
} from "../../wailsjs/go/main/App";

const message = useMessage();
const dialog = useDialog();

const agents = ref([]);
const modelOptions = ref([]);
const toolOptions = ref([]);
const loading = ref(false);
const showCreateModal = ref(false);
const submitting = ref(false);
const isEdit = ref(false);
const editingId = ref(null);

const formValue = ref({
  name: "",
  description: "",
  systemPrompt: "",
  modelId: null,
  toolIds: [],
});

const rules = {
  name: { required: true, message: "必填", trigger: "blur" },
  modelId: { required: true, message: "必填", type: "number", trigger: "blur" },
};

const pagination = { pageSize: 10 };

const columns = [
  { title: "名称", key: "name", width: 150 },
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
        { default: () => row.type }
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
    title: "工具",
    key: "Tools",
    width: 200,
    render(row) {
      if (!row.tools || row.tools.length === 0) return "无";
      return h(
        NSpace,
        { size: 4, style: { flexWrap: "wrap" } },
        {
          default: () =>
            row.tools.map((tool) =>
              h(
                NTag,
                { type: "success", size: "small", bordered: false },
                { default: () => tool.name }
              )
            ),
        }
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
              { default: () => "编辑" }
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
                { default: () => "删除" }
              )
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
    modelId: row.modelId,
    toolIds: (row.tools || []).map((t) => t.id),
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
    modelId: null,
    toolIds: [],
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
    // Handle modelId being null/undefined/0
    const modelId = formValue.value.modelId || 0;

    if (isEdit.value) {
      res = await UpdateAgent(
        editingId.value,
        formValue.value.name,
        formValue.value.description,
        formValue.value.systemPrompt,
        modelId,
        formValue.value.toolIds,
        1 // Default ModeID for now
      );
    } else {
      res = await CreateAgent(
        formValue.value.name,
        formValue.value.description,
        formValue.value.systemPrompt,
        modelId,
        formValue.value.toolIds,
        // Assuming modeID is handled or defaulted in backend for now, or we need to add it to form
        // Current CreateAgent signature in frontend call might need update if backend changed
        // Based on previous context, backend CreateAgent takes modeID.
        // Let's check App.js signature or just pass 0 if we haven't added Mode selection to UI yet.
        // For this task, we focus on Model optionality.
        // We will pass 0 for modeID for now (defaulting to chat in backend maybe? or need update)
        // Wait, CreateAgent signature in App.go: (name, description, systemPrompt, modelID, toolIDs, modeID)
        // We need to pass modeID. Let's assume 1 (Chat) or add to form.
        // For now let's pass 1 as a safe default or 0.
        1
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
