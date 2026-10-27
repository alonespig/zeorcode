<script setup lang="ts">
import { computed, shallowRef } from "vue";
import { UploadFilled } from "@element-plus/icons-vue";
import type { UploadFile, UploadInstance } from "element-plus";
import { message } from "@/utils/message";
import {
  importProblemArchive,
  type ProblemArchiveImportResult
} from "@/api/admin/problems";

const visible = defineModel<boolean>({ required: true });
const emit = defineEmits<{
  imported: [];
}>();

const uploadRef = shallowRef<UploadInstance>();
const selectedFile = shallowRef<File>();
const importing = shallowRef(false);
const result = shallowRef<ProblemArchiveImportResult>();

const canImport = computed(() => Boolean(selectedFile.value) && !importing.value);

function handleFileChange(file: UploadFile) {
  const raw = file.raw;
  if (!raw) return;
  if (!raw.name.toLowerCase().endsWith(".zip")) {
    message("请选择 ZeorCode 导出的 ZIP 归档", { type: "warning" });
    uploadRef.value?.clearFiles();
    selectedFile.value = undefined;
    return;
  }
  selectedFile.value = raw;
  result.value = undefined;
}

function resetDialog() {
  uploadRef.value?.clearFiles();
  selectedFile.value = undefined;
  result.value = undefined;
}

async function submitImport() {
  if (!selectedFile.value) return;
  importing.value = true;
  try {
    const response = await importProblemArchive(selectedFile.value);
    if (!response.data) {
      message("导入完成，但服务器未返回统计结果", { type: "warning" });
      return;
    }
    result.value = response.data;
    message(`成功导入 ${response.data.imported} 道题目`, {
      type: "success"
    });
    emit("imported");
  } finally {
    importing.value = false;
  }
}
</script>

<template>
  <el-dialog
    v-model="visible"
    title="批量导入题目"
    width="560px"
    destroy-on-close
    :close-on-click-modal="!importing"
    :close-on-press-escape="!importing"
    @closed="resetDialog"
  >
    <div class="archive-note">
      <div class="archive-note__title">导入说明</div>
      <ul class="archive-note__list">
        <li>请选择后台“一键导出”生成的 ZIP，题面、标签、样例和测试数据会一起导入。</li>
        <li>已存在的题号会自动跳过，不会覆盖当前题目和测试数据。</li>
        <li>导入前会完整校验归档版本、文件路径、大小和 SHA-256。</li>
      </ul>
    </div>

    <el-upload
      ref="uploadRef"
      drag
      action="#"
      accept=".zip,application/zip"
      :auto-upload="false"
      :limit="1"
      :disabled="importing"
      :on-change="handleFileChange"
    >
      <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
      <div class="el-upload__text">将题目归档拖到这里，或<em>点击选择</em></div>
      <template #tip>
        <div class="el-upload__tip">仅支持 ZeorCode 题目归档，最大 2GB</div>
      </template>
    </el-upload>

    <el-alert
      v-if="result"
      class="archive-result"
      type="success"
      :closable="false"
      show-icon
    >
      <template #title>
        导入完成：新增 {{ result.imported }} 道，跳过 {{ result.skipped }} 道
      </template>
      <div v-if="result.skippedIds.length" class="archive-result__ids">
        已跳过题号：{{ result.skippedIds.join("、") }}
      </div>
    </el-alert>

    <template #footer>
      <el-button :disabled="importing" @click="visible = false">关闭</el-button>
      <el-button
        type="primary"
        :loading="importing"
        :disabled="!canImport"
        @click="submitImport"
      >
        开始导入
      </el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.archive-note {
  margin-bottom: 18px;
  padding: 14px 16px;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
}

.archive-note__title {
  margin-bottom: 6px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.archive-note__list {
  padding-left: 18px;
  margin: 0;
  line-height: 1.8;
}

.archive-result {
  margin-top: 18px;
}

.archive-result__ids {
  max-height: 72px;
  margin-top: 6px;
  overflow: auto;
  line-height: 1.6;
  word-break: break-all;
}
</style>
