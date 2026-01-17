<template>
  <div style="padding: 20px">
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">MCP 服务器管理</n-h2>
      <n-button type="primary" @click="showCreateModal = true">
        新建 MCP 服务器
      </n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="servers"
      :loading="loading"
      :pagination="pagination"
      :scroll-x="1000"
    />

    <!-- Create/Edit Modal -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      :title="isEdit ? '编辑 MCP 服务器' : '新建 MCP 服务器'"
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
          <n-input v-model:value="formValue.name" placeholder="服务器名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formValue.description" placeholder="描述" />
        </n-form-item>
        <n-form-item label="启用" path="enabled">
          <n-switch v-model:value="formValue.enabled" />
        </n-form-item>
        <n-form-item label="类型" path="type">
          <n-radio-group v-model:value="formValue.type">
            <n-radio-button value="stdio">Stdio (命令行)</n-radio-button>
            <n-radio-button value="sse">SSE (HTTP)</n-radio-button>
          </n-radio-group>
        </n-form-item>

        <template v-if="formValue.type === 'stdio'">
          <n-form-item label="命令 (Command)" path="command">
            <n-input v-model:value="formValue.command" placeholder="e.g. npx, python, /path/to/executable" />
          </n-form-item>
          <n-form-item label="参数 (Args)" path="args">
            <n-input
              v-model:value="formValue.args"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 5 }"
              placeholder='JSON 数组, e.g. ["-y", "@modelcontextprotocol/server-filesystem", "/users/me/files"]'
            />
          </n-form-item>
          <n-form-item label="环境变量 (Env)" path="env">
            <n-input
              v-model:value="formValue.env"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 5 }"
              placeholder='JSON 对象, e.g. {"PATH": "/usr/bin"}'
            />
          </n-form-item>
        </template>

        <template v-if="formValue.type === 'sse'">
          <n-form-item label="URL" path="url">
            <n-input v-model:value="formValue.url" placeholder="e.g. http://localhost:8000/sse" />
          </n-form-item>
        </template>
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

    <!-- Tools Modal -->
    <n-modal
      v-model:show="showToolsModal"
      preset="card"
      :title="`工具列表 - ${currentServerName}`"
      style="width: 800px; height: 600px;"
    >
      <div style="height: 100%; display: flex; flex-direction: column;">
        <div v-if="toolsLoading" style="display: flex; justify-content: center; padding: 20px;">
          加载中...
        </div>
        <div v-else style="flex: 1; overflow: auto; padding-right: 10px;">
           <n-data-table
            :columns="[
              { title: '名称', key: 'Name', width: 200 },
              { title: '描述', key: 'Desc' },
            ]"
            :data="currentTools"
          />
          <div v-if="currentTools.length === 0" style="text-align: center; padding: 20px; color: #999;">
            暂无工具或连接失败
          </div>
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, h } from "vue";
import {
  NButton,
  NSpace,
  useMessage,
  useDialog,
  NTag,
  NInput,
  NForm,
  NFormItem,
  NModal,
  NRadioGroup,
  NRadioButton,
  NSwitch,
  NH2,
  NDataTable,
} from "naive-ui";
import {
  ListMCPServers,
  CreateMCPServer,
  UpdateMCPServer,
  DeleteMCPServer,
  ListMCPTools,
} from "../../wailsjs/go/main/App";

const message = useMessage();
const dialog = useDialog();

const servers = ref([]);
const loading = ref(false);
const showCreateModal = ref(false);
const showToolsModal = ref(false);
const currentTools = ref([]);
const currentServerName = ref("");
const toolsLoading = ref(false);

const submitting = ref(false);
const isEdit = ref(false);
const editingId = ref(null);

const formValue = ref({
  name: "",
  description: "",
  type: "stdio",
  command: "",
  args: "[]",
  env: "{}",
  url: "",
  enabled: true,
});

const rules = {
  name: { required: true, message: "必填", trigger: "blur" },
  type: { required: true, message: "必填", trigger: "blur" },
  command: {
    required: true,
    message: "命令必填",
    trigger: "blur",
    validator: (rule, value) => {
      if (formValue.value.type === "stdio" && !value) return new Error("命令必填");
      return true;
    },
  },
  url: {
    required: true,
    message: "URL必填",
    trigger: "blur",
    validator: (rule, value) => {
      if (formValue.value.type === "sse" && !value) return new Error("URL必填");
      return true;
    },
  },
};

const pagination = { pageSize: 10 };

const columns = [
  { title: "名称", key: "name", width: 150 },
  {
    title: "类型",
    key: "type",
    width: 100,
    render(row) {
      const type = row.type === "stdio" ? "info" : "success";
      return h(
        NTag,
        { type: type, bordered: false },
        { default: () => row.type.toUpperCase() }
      );
    },
  },
  {
    title: "状态",
    key: "enabled",
    width: 100,
    render(row) {
      return h(
        NTag,
        { type: row.enabled ? "success" : "warning", bordered: false },
        { default: () => (row.enabled ? "启用" : "禁用") }
      );
    },
  },
  {
    title: "配置详情",
    key: "details",
    minWidth: 200,
    render(row) {
      if (row.type === "stdio") {
        return `${row.command} ${row.args !== "[]" ? "..." : ""}`;
      }
      return row.url;
    },
  },
  {
    title: "描述",
    key: "description",
    minWidth: 150,
    ellipsis: { tooltip: true },
  },
  {
    title: "操作",
    key: "actions",
    width: 150,
    fixed: "right",
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(
            NButton,
            { size: "small", type: "info", onClick: () => handleViewTools(row) },
            { default: () => "查看工具" }
          ),
          h(
            NButton,
            { size: "small", onClick: () => handleEdit(row) },
            { default: () => "编辑" }
          ),
          h(
            NButton,
            { size: "small", type: "error", onClick: () => handleDelete(row) },
            { default: () => "删除" }
          ),
        ],
      });
    },
  },
];

async function loadData() {
  loading.value = true;
  try {
    const res = await ListMCPServers();
    if (res.code === 200) {
      servers.value = res.data || [];
    } else {
      message.error(res.msg);
    }
  } catch (e) {
    message.error("加载数据失败: " + e);
  } finally {
    loading.value = false;
  }
}

async function handleViewTools(row) {
  currentServerName.value = row.name;
  showToolsModal.value = true;
  toolsLoading.value = true;
  currentTools.value = [];
  
  try {
    const res = await ListMCPTools(row.id);
    if (res.code === 200) {
      currentTools.value = res.data || [];
    } else {
      message.error(res.msg);
    }
  } catch (e) {
    message.error("获取工具失败: " + e);
  } finally {
    toolsLoading.value = false;

    console.log('currentTools', currentTools.value);
    
  }
}

function handleEdit(row) {
  isEdit.value = true;
  editingId.value = row.id;
  formValue.value = {
    name: row.name,
    description: row.description,
    type: row.type,
    command: row.command,
    args: row.args || "[]",
    env: row.env || "{}",
    url: row.url,
    enabled: row.enabled,
  };
  showCreateModal.value = true;
}

function handleDelete(row) {
  dialog.warning({
    title: "确认删除",
    content: `确认删除 MCP 服务器 "${row.name}"?`,
    positiveText: "确认",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        const res = await DeleteMCPServer(row.id);
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
    type: "stdio",
    command: "",
    args: "[]",
    env: "{}",
    url: "",
    enabled: true,
  };
  isEdit.value = false;
  editingId.value = null;
}

async function handleSubmit() {
  if (!formValue.value.name || !formValue.value.type) {
    message.warning("名称和类型为必填项");
    return;
  }

  submitting.value = true;
  try {
    let res;
    if (isEdit.value) {
      res = await UpdateMCPServer(
        editingId.value,
        formValue.value.name,
        formValue.value.description,
        formValue.value.type,
        formValue.value.command,
        formValue.value.args,
        formValue.value.env,
        formValue.value.url,
        formValue.value.enabled
      );
    } else {
      res = await CreateMCPServer(
        formValue.value.name,
        formValue.value.description,
        formValue.value.type,
        formValue.value.command,
        formValue.value.args,
        formValue.value.env,
        formValue.value.url
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
