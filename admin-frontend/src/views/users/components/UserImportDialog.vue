<script setup lang="ts">
import { ref, shallowRef } from "vue";
import type { UploadFile, UploadUserFile } from "element-plus";
import { Download, UploadFilled } from "@element-plus/icons-vue";
import { message } from "@/utils/message";
import {
  downloadUserImportTemplate,
  importAdminUsers,
  type BatchCreateUsersResult
} from "@/api/admin/users";

defineOptions({ name: "UserImportDialog" });

const emit = defineEmits<{ imported: [] }>();
const visible = defineModel<boolean>({ default: false });

const MAX_FILE_SIZE = 5 * 1024 * 1024;

const fileList = ref<UploadUserFile[]>([]);
const selectedFile = shallowRef<File | null>(null);
const importing = shallowRef(false);
const downloading = shallowRef(false);
const result = shallowRef<BatchCreateUsersResult | null>(null);

function clearSelectedFile() {
  fileList.value = [];
  selectedFile.value = null;
}

function handleFileChange(uploadFile: UploadFile) {
  const file = uploadFile.raw;
  if (!file) return;
  if (!file.name.toLowerCase().endsWith(".xlsx")) {
    message("仅支持 .xlsx 格式的 Excel 文件", { type: "warning" });
    clearSelectedFile();
    return;
  }
  if (file.size > MAX_FILE_SIZE) {
    message("Excel 文件大小不能超过 5MB", { type: "warning" });
    clearSelectedFile();
    return;
  }
  selectedFile.value = file;
}

function handleFileRemove() {
  selectedFile.value = null;
}

function handleExceed() {
  message("一次只能选择一个 Excel 文件", { type: "warning" });
}

async function handleDownloadTemplate() {
  if (downloading.value) return;
  downloading.value = true;
  try {
    const blob = await downloadUserImportTemplate();
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = "用户导入模板.xlsx";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    downloading.value = false;
  }
}

async function submitImport() {
  if (!selectedFile.value || importing.value) return;
  importing.value = true;
  try {
    const res = await importAdminUsers(selectedFile.value);
    result.value = res.data ?? null;
    emit("imported");
  } finally {
    importing.value = false;
  }
}

function resetDialog() {
  clearSelectedFile();
  result.value = null;
}
</script>

<template>
  <el-dialog
    v-model="visible"
    title="Excel 导入用户"
    width="580px"
    style="max-width: calc(100vw - 32px)"
    :close-on-click-modal="!importing"
    :close-on-press-escape="!importing"
    :show-close="!importing"
    @closed="resetDialog"
  >
    <template v-if="!result">
      <div
        class="mb-5 rounded-lg border border-blue-200 bg-blue-50 px-4 py-3 text-sm leading-6 text-slate-600"
      >
        <p class="font-medium text-slate-700">按模板填写用户资料</p>
        <p>学号、姓名和用户名必须填写，邮箱可以留空。</p>
        <p>新账号统一创建为普通用户，初始密码与学号相同。</p>
      </div>

      <div
        class="mb-4 flex items-center justify-between rounded-md border border-gray-200 px-4 py-3"
      >
        <div>
          <p class="font-medium text-slate-700">先下载标准模板</p>
          <p class="mt-0.5 text-xs text-gray-400">
            请保留工作表名称和第一行表头，每次最多导入 500 人
          </p>
        </div>
        <el-button
          type="primary"
          link
          :icon="Download"
          :loading="downloading"
          @click="handleDownloadTemplate"
        >
          下载模板
        </el-button>
      </div>

      <el-upload
        v-model:file-list="fileList"
        drag
        accept=".xlsx"
        :auto-upload="false"
        :limit="1"
        :on-change="handleFileChange"
        :on-remove="handleFileRemove"
        :on-exceed="handleExceed"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将填写好的 Excel 拖到这里，或<em>点击选择</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            仅支持 .xlsx 文件，大小不超过 5MB。任一行失败时整批数据都会回滚。
          </div>
        </template>
      </el-upload>
    </template>

    <el-result
      v-else
      icon="success"
      title="用户导入完成"
      sub-title="用户列表已经刷新"
    >
      <template #extra>
        <div
          class="mx-auto min-w-40 rounded-lg border border-gray-200 bg-gray-50 px-3 py-4 text-center"
        >
          <div class="text-2xl font-semibold text-blue-600">
            {{ result.created }}
          </div>
          <div class="mt-1 text-xs text-gray-500">新建账号</div>
        </div>
      </template>
    </el-result>

    <template #footer>
      <el-button v-if="!result" :disabled="importing" @click="visible = false">
        取消
      </el-button>
      <el-button
        v-if="!result"
        type="primary"
        :loading="importing"
        :disabled="!selectedFile"
        @click="submitImport"
      >
        开始导入
      </el-button>
      <el-button v-else type="primary" @click="visible = false">完成</el-button>
    </template>
  </el-dialog>
</template>
