<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from "vue";
import { ElMessageBox, type FormInstance, type FormRules } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import {
  createTag,
  deleteTag,
  getAdminTagList,
  updateTag,
  type AdminTagItem
} from "@/api/admin/tags";

defineOptions({ name: "AdminTags" });

type TagRow = AdminTagItem & { _busy?: boolean };

const dataList = ref<TagRow[]>([]);
const loading = ref(false);
const dialogVisible = ref(false);
const submitting = ref(false);
const editingId = ref<number>();
const formRef = ref<FormInstance>();
const form = reactive({ name: "" });

const rules: FormRules = {
  name: [
    { required: true, message: "请输入标签名称", trigger: "blur" },
    { max: 64, message: "标签名称不能超过 64 个字符", trigger: "blur" }
  ]
};

const columns: TableColumnList = [
  { label: "编号", prop: "id", width: 100, align: "center" },
  { label: "标签名称", prop: "name", minWidth: 220 },
  {
    label: "关联题目",
    prop: "problemCount",
    width: 130,
    align: "center"
  },
  {
    label: "关联题单",
    prop: "problemSetCount",
    width: 130,
    align: "center"
  },
  {
    label: "操作",
    slot: "operation",
    width: 140,
    fixed: "right",
    align: "center"
  }
];

async function loadTags() {
  loading.value = true;
  try {
    const res = await getAdminTagList();
    dataList.value = res.data?.tags ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingId.value = undefined;
  form.name = "";
  dialogVisible.value = true;
  nextTick(() => formRef.value?.clearValidate());
}

function openEdit(row: TagRow) {
  editingId.value = row.id;
  form.name = row.name;
  dialogVisible.value = true;
  nextTick(() => formRef.value?.clearValidate());
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid || submitting.value) return;

  const name = form.name.trim();
  if (!name) {
    message("请输入标签名称", { type: "warning" });
    return;
  }

  submitting.value = true;
  try {
    if (editingId.value === undefined) {
      await createTag({ name });
      message("标签已创建", { type: "success" });
    } else {
      await updateTag(editingId.value, { name });
      message("标签已更新", { type: "success" });
    }
    dialogVisible.value = false;
    await loadTags();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    submitting.value = false;
  }
}

async function remove(row: TagRow) {
  const relationText =
    row.problemCount || row.problemSetCount
      ? `该标签将从 ${row.problemCount} 道题目和 ${row.problemSetCount} 个题单中移除。`
      : "";
  try {
    await ElMessageBox.confirm(
      `确认删除标签「${row.name}」？${relationText}删除后不可恢复。`,
      "删除标签",
      { type: "warning" }
    );
  } catch {
    return;
  }

  if (row._busy) return;
  row._busy = true;
  try {
    await deleteTag(row.id);
    message("标签已删除", { type: "success" });
    await loadTags();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._busy = false;
  }
}

onMounted(loadTags);
</script>

<template>
  <div>
    <PureTableBar :columns="columns" @refresh="loadTags">
      <template #title>
        <div class="flex items-center">
          <span>算法标签</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            共 {{ dataList.length }} 个
          </span>
        </div>
      </template>
      <template #buttons>
        <el-button type="primary" :icon="Plus" @click="openCreate">
          新建标签
        </el-button>
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <PureTable
          border
          adaptive
          table-layout="auto"
          align-whole="center"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
        >
          <template #operation="{ row }">
            <el-button link size="small" type="primary" @click="openEdit(row)">
              编辑
            </el-button>
            <el-button
              link
              size="small"
              type="danger"
              :loading="row._busy"
              @click="remove(row)"
            >
              删除
            </el-button>
          </template>
        </PureTable>
      </template>
    </PureTableBar>

    <el-dialog
      v-model="dialogVisible"
      :title="editingId === undefined ? '新建标签' : '编辑标签'"
      width="440px"
      destroy-on-close
      @closed="formRef?.resetFields()"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="82px">
        <el-form-item label="标签名称" prop="name">
          <el-input
            v-model="form.name"
            maxlength="64"
            show-word-limit
            placeholder="请输入标签名称"
            @keyup.enter="submit"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submit">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
