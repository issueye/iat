<template>
  <div style="padding: 20px">
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">工具管理</n-h2>
      <n-button type="primary" @click="showCreateModal = true">
        新建工具
      </n-button>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="tools"
      :loading="loading"
      :pagination="pagination"
    />

    <!-- Create/Edit Modal -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      :title="isEdit ? '编辑工具' : '新建工具'"
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
          <n-input v-model:value="formValue.name" placeholder="工具名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="formValue.description" placeholder="描述" />
        </n-form-item>
        <n-form-item label="类型" path="type">
           <n-select
            v-model:value="formValue.type"
            :options="typeOptions"
            placeholder="选择工具类型"
          />
        </n-form-item>
        <n-form-item label="内容/配置" path="content">
          <n-input
            v-model:value="formValue.content"
            type="textarea"
            :autosize="{ minRows: 5, maxRows: 10 }"
            placeholder="脚本内容或 API 配置"
          />
        </n-form-item>
        <n-form-item label="参数定义" path="parameters">
           <n-input
            v-model:value="formValue.parameters"
            type="textarea"
            :autosize="{ minRows: 3, maxRows: 6 }"
            placeholder="JSON 格式的参数定义"
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
import { NButton, NSpace, useMessage, useDialog, NTag, NInput, NForm, NFormItem, NModal, NSelect, NH2, NDataTable } from "naive-ui";
import {
  ListTools,
  CreateTool,
  UpdateTool,
  DeleteTool,
} from "../../wailsjs/go/main/App";

const message = useMessage();
const dialog = useDialog();

const tools = ref([]);
const loading = ref(false);
const showCreateModal = ref(false);
const submitting = ref(false);
const isEdit = ref(false);
const editingId = ref(null);

const formValue = ref({
  name: "",
  description: "",
  type: "script",
  content: "",
  parameters: "{}",
});

const typeOptions = [
    { label: '脚本 (Script)', value: 'script' },
    { label: 'API', value: 'api' },
    { label: '函数 (Function)', value: 'function' }
]

const rules = {
  name: { required: true, message: "必填", trigger: "blur" },
  type: { required: true, message: "必填", trigger: "blur" },
};

const pagination = { pageSize: 10 };

const columns = [
  { title: "名称", key: "name", width: 150 },
  { title: "类型", key: "type", width: 100, 
    render(row) {
        const type = row.type === 'builtin' ? 'primary' : 'info';
        return h(NTag, { type: type, bordered: false }, { default: () => row.type })
    }
  },
  { title: "描述", key: "description" },
  {
    title: "操作",
    key: "actions",
    width: 150,
    render(row) {
      if (row.type === 'builtin') {
        return null;
      }
      return h(NSpace, null, {
        default: () => [
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
    const res = await ListTools();
    if (res.code === 200) {
      tools.value = res.data || [];
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
    type: row.type,
    content: row.content,
    parameters: row.parameters,
  };
  showCreateModal.value = true;
}

function handleDelete(row) {
  dialog.warning({
    title: "确认删除",
    content: `确认删除工具 "${row.name}"?`,
    positiveText: "确认",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        const res = await DeleteTool(row.id);
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
    type: "script",
    content: "",
    parameters: "{}",
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
    const toolData = {
        ...formValue.value,
        id: isEdit.value ? editingId.value : 0
    }
    
    if (isEdit.value) {
      res = await UpdateTool(toolData);
    } else {
      res = await CreateTool(toolData);
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
