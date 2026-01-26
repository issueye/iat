<template>
  <div>
    <n-space justify="space-between" align="center" style="margin-bottom: 16px">
      <n-h2 style="margin: 0">项目列表</n-h2>
      <n-space>
        <n-button secondary @click="handleIndexAll" :loading="indexingAll">
          索引全部
        </n-button>
        <n-button type="primary" @click="showCreateModal = true">
          新建项目
        </n-button>
      </n-space>
    </n-space>

    <n-data-table
      :columns="columns"
      :data="projects"
      :loading="loading"
      :pagination="pagination"
      :scroll-x="1000"
    />

    <!-- Create/Edit Modal -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      :title="isEdit ? '编辑项目' : '新建项目'"
    >
      <n-form ref="formRef" :model="formValue" :rules="rules">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="formValue.name" placeholder="项目名称" />
        </n-form-item>
        <n-form-item label="路径" path="path">
          <n-input-group>
            <n-input
              v-model:value="formValue.path"
              placeholder="项目路径"
              readonly
            />
            <n-button type="primary" @click="handleSelectDir">选择</n-button>
          </n-input-group>
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input
            v-model:value="formValue.description"
            type="textarea"
            placeholder="描述"
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
import { NButton, NSpace, useMessage, useDialog, NInputGroup } from "naive-ui";
import {
  ListProjects,
  CreateProject,
  UpdateProject,
  DeleteProject,
  IndexProject,
  IndexAllProjects,
  SelectDirectory,
} from "../../wailsjs/go/main/App";

const message = useMessage();
const dialog = useDialog();

const projects = ref([]);
const loading = ref(false);
const showCreateModal = ref(false);
const submitting = ref(false);
const indexingAll = ref(false);
const indexingProjectId = ref(null);
const isEdit = ref(false);
const editingId = ref(null);

const formValue = ref({
  name: "",
  description: "",
  path: "",
});

const rules = {
  name: {
    required: true,
    message: "请输入项目名称",
    trigger: "blur",
  },
};

const pagination = {
  pageSize: 10,
};

const columns = [
  {
    title: "ID",
    key: "id",
    width: 80,
  },
  {
    title: "名称",
    key: "name",
    width: 200,
  },
  {
    title: "路径",
    key: "path",
    width: 200,
    ellipsis: {
      tooltip: true,
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
    title: "创建时间",
    key: "createdAt",
    width: 200,
    render(row) {
      return new Date(row.createdAt).toLocaleString();
    },
  },
  {
    title: "操作",
    key: "actions",
    width: 220,
    fixed: "right",
    render(row) {
      return h(NSpace, null, {
        default: () => [
          h(
            NButton,
            {
              size: "small",
              secondary: true,
              loading: indexingProjectId.value === row.id,
              onClick: () => handleIndex(row),
            },
            { default: () => "索引" }
          ),
          h(
            NButton,
            {
              size: "small",
              onClick: () => handleEdit(row),
            },
            { default: () => "编辑" }
          ),
          h(
            NButton,
            {
              size: "small",
              type: "error",
              onClick: () => handleDelete(row),
            },
            { default: () => "删除" }
          ),
        ],
      });
    },
  },
];

async function loadProjects() {
  loading.value = true;
  try {
    const res = await ListProjects();
    if (res.code === 200) {
      projects.value = res.data || [];
    } else {
      message.error(res.msg);
    }
  } catch (e) {
    message.error("加载项目失败: " + e);
  } finally {
    loading.value = false;
  }
}

async function handleIndex(row) {
  indexingProjectId.value = row.id;
  try {
    const res = await IndexProject(row.id);
    if (res.code === 200) {
      const indexed = res.data?.indexed ?? 1;
      const files = res.data?.files;
      const dbPath = res.data?.dbPath ? `（${res.data.dbPath}）` : "";
      message.success(
        `索引完成：项目${indexed}${files != null ? ` 文件${files}` : ""}${dbPath}`
      );
    } else {
      message.error(res.msg || "索引失败");
    }
  } catch (e) {
    message.error("索引失败: " + e);
  } finally {
    indexingProjectId.value = null;
  }
}

async function handleIndexAll() {
  indexingAll.value = true;
  try {
    const res = await IndexAllProjects();
    if (res.code === 200) {
      const indexed = res.data?.indexed ?? 0;
      const files = res.data?.files;
      const dbPath = res.data?.dbPath ? `（${res.data.dbPath}）` : "";
      message.success(
        `索引完成：项目${indexed}${files != null ? ` 文件${files}` : ""}${dbPath}`
      );
    } else {
      message.error(res.msg || "索引失败");
    }
  } catch (e) {
    message.error("索引失败: " + e);
  } finally {
    indexingAll.value = false;
  }
}

function handleEdit(row) {
  isEdit.value = true;
  editingId.value = row.id;
  formValue.value = {
    name: row.name,
    description: row.description,
    path: row.path,
  };
  showCreateModal.value = true;
}

async function handleSelectDir() {
  try {
    const res = await SelectDirectory();
    if (res.code === 200) {
      if (res.data) {
        formValue.value.path = res.data;
      }
    } else {
      message.error(res.msg);
    }
  } catch (e) {
    message.error("选择目录失败: " + e);
  }
}

function handleDelete(row) {
  dialog.warning({
    title: "确认删除",
    content: `确定要删除项目 "${row.name}" 吗？`,
    positiveText: "确认",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        const res = await DeleteProject(row.id);
        if (res.code === 200) {
          message.success("删除成功");
          loadProjects();
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
  formValue.value = { name: "", description: "", path: "" };
  isEdit.value = false;
  editingId.value = null;
}

async function handleSubmit() {
  if (!formValue.value.name) {
    message.warning("请输入名称");
    return;
  }

  submitting.value = true;
  try {
    let res;
    if (isEdit.value) {
      res = await UpdateProject(
        editingId.value,
        formValue.value.name,
        formValue.value.description,
        formValue.value.path
      );
    } else {
      res = await CreateProject(
        formValue.value.name,
        formValue.value.description,
        formValue.value.path
      );
    }

    if (res.code === 200) {
      message.success(isEdit.value ? "更新成功" : "创建成功");
      closeModal();
      loadProjects();
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
  loadProjects();
});
</script>
