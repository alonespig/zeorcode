<script setup lang="ts">
import { computed, nextTick, onMounted, reactive, ref } from "vue";
import dayjs from "dayjs";
import { ElMessageBox, type FormInstance, type FormRules } from "element-plus";
import { Plus } from "@element-plus/icons-vue";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import {
  createLanguage,
  deleteLanguage,
  getAdminLanguages,
  updateLanguage,
  type LanguageItem,
  type SaveLanguageParams
} from "@/api/admin/languages";

defineOptions({ name: "AdminLanguages" });

type LanguageRow = LanguageItem & { _busy?: boolean };

const dataList = ref<LanguageRow[]>([]);
const loading = ref(false);
const dialogVisible = ref(false);
const submitting = ref(false);
const editingId = ref<number>();
const formRef = ref<FormInstance>();
const form = reactive<SaveLanguageParams>({ name: "", status: 1, sort: 0 });

const dialogTitle = computed(() =>
  editingId.value === undefined ? "新增编程语言" : "编辑编程语言"
);

const rules: FormRules = {
  name: [
    { required: true, message: "请输入语言名称", trigger: "blur" },
    { max: 32, message: "语言名称不能超过 32 个字符", trigger: "blur" }
  ],
  sort: [{ required: true, message: "请输入排序值", trigger: "change" }]
};

const columns: TableColumnList = [
  { label: "编号", prop: "id", width: 90, align: "center" },
  { label: "名称", prop: "name", minWidth: 180 },
  { label: "状态", slot: "status", width: 120, align: "center" },
  { label: "排序", prop: "sort", width: 100, align: "center" },
  { label: "创建时间", slot: "createdAt", width: 170, align: "center" },
  { label: "更新时间", slot: "updatedAt", width: 170, align: "center" },
  {
    label: "操作",
    slot: "operation",
    width: 140,
    fixed: "right",
    align: "center"
  }
];

function formatTime(value: string) {
  return value ? dayjs(value).format("YYYY-MM-DD HH:mm") : "—";
}

async function loadLanguages() {
  loading.value = true;
  try {
    const res = await getAdminLanguages();
    dataList.value = res.data ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingId.value = undefined;
  form.name = "";
  form.status = 1;
  form.sort =
    dataList.value.reduce((max, item) => Math.max(max, item.sort), 0) + 1;
  dialogVisible.value = true;
  nextTick(() => formRef.value?.clearValidate());
}

function openEdit(row: LanguageRow) {
  editingId.value = row.id;
  form.name = row.name;
  form.status = row.status;
  form.sort = row.sort;
  dialogVisible.value = true;
  nextTick(() => formRef.value?.clearValidate());
}

async function save() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid || submitting.value) return;

  const data: SaveLanguageParams = {
    name: form.name.trim(),
    status: form.status,
    sort: form.sort
  };
  if (!data.name) {
    message("请输入语言名称", { type: "warning" });
    return;
  }

  submitting.value = true;
  try {
    if (editingId.value === undefined) {
      await createLanguage(data);
      message("编程语言已添加", { type: "success" });
    } else {
      await updateLanguage(editingId.value, data);
      message("编程语言已更新", { type: "success" });
    }
    dialogVisible.value = false;
    await loadLanguages();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    submitting.value = false;
  }
}

async function toggleStatus(
  row: LanguageRow,
  enabled: string | number | boolean
) {
  if (row._busy) return;
  const previous = row.status;
  const nextStatus: 0 | 1 = enabled ? 1 : 0;
  row.status = nextStatus;
  row._busy = true;
  try {
    await updateLanguage(row.id, {
      name: row.name,
      status: nextStatus,
      sort: row.sort
    });
    message(nextStatus === 1 ? "语言已启用" : "语言已停用", {
      type: "success"
    });
  } catch {
    row.status = previous;
  } finally {
    row._busy = false;
  }
}

async function remove(row: LanguageRow) {
  try {
    await ElMessageBox.confirm(
      `确认删除编程语言「${row.name}」？已有提交记录不会被删除。`,
      "删除编程语言",
      { type: "warning" }
    );
  } catch {
    return;
  }

  if (row._busy) return;
  row._busy = true;
  try {
    await deleteLanguage(row.id);
    message("编程语言已删除", { type: "success" });
    await loadLanguages();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    row._busy = false;
  }
}

onMounted(loadLanguages);
</script>

<template>
  <div>
    <PureTableBar :columns="columns" @refresh="loadLanguages">
      <template #title>
        <div class="flex items-center">
          <span>编程语言</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            共 {{ dataList.length }} 种
          </span>
        </div>
      </template>
      <template #buttons>
        <el-button type="primary" :icon="Plus" @click="openCreate">
          新增语言
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
          <template #status="{ row }">
            <el-switch
              :model-value="row.status === 1"
              :loading="row._busy"
              inline-prompt
              active-text="启用"
              inactive-text="停用"
              @change="value => toggleStatus(row, value)"
            />
          </template>
          <template #createdAt="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
          <template #updatedAt="{ row }">
            {{ formatTime(row.updatedAt) }}
          </template>
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
      :title="dialogTitle"
      width="480px"
      destroy-on-close
      :close-on-click-modal="false"
      @closed="formRef?.resetFields()"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input
            v-model="form.name"
            maxlength="32"
            show-word-limit
            placeholder="例如 c++、java、python"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
            active-text="启用"
            inactive-text="停用"
          />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number
            v-model="form.sort"
            :min="0"
            :max="9999"
            controls-position="right"
          />
        </el-form-item>
        <p class="ml-20 mt-1 text-xs leading-6 text-gray-400">
          名称必须与评测程序支持的语言一致，未知名称会被后端拒绝。
        </p>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="save">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
