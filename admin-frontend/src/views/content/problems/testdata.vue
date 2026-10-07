<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessageBox } from "element-plus";
import { Delete, Download, Upload } from "@element-plus/icons-vue";
import { message } from "@/utils/message";
import { PureTableBar } from "@/components/RePureTableBar";
import {
  deleteTestDataFiles,
  downloadTestDataFiles,
  getProblem,
  getTestDataContent,
  getTestDataFiles,
  replaceTestDataZip,
  uploadTestDataFile,
  type TestDataFile
} from "@/api/admin/problems";

defineOptions({ name: "AdminProblemTestData" });

const route = useRoute();
const router = useRouter();
const problemId = computed(() => String(route.params.id ?? ""));
const problemName = ref("");
const remoteOJ = ref("");
const dataList = ref<TestDataFile[]>([]);
const selected = ref<TestDataFile[]>([]);
const loading = ref(false);
const uploading = ref(false);
const downloading = ref(false);
const deleting = ref(false);
const fileInput = ref<HTMLInputElement>();
const zipInput = ref<HTMLInputElement>();

const previewVisible = ref(false);
const previewLoading = ref(false);
const previewName = ref("");
const previewContent = ref("");
const previewTruncated = ref(false);
let previewSeq = 0;

const columns: TableColumnList = [
  { type: "selection", width: 50, align: "center" },
  { label: "文件名", slot: "name", minWidth: 260 },
  { label: "大小", slot: "size", width: 140, align: "center" }
];

function formatSize(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  return `${(size / 1024 / 1024).toFixed(1)} MB`;
}

async function loadFiles() {
  if (!problemId.value || remoteOJ.value) return;
  loading.value = true;
  try {
    const res = await getTestDataFiles(problemId.value);
    dataList.value = res.data ?? [];
    selected.value = [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
}

async function initialize() {
  loading.value = true;
  try {
    const res = await getProblem(problemId.value);
    problemName.value = res.data?.name ?? "";
    remoteOJ.value = res.data?.oj ?? "";
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loading.value = false;
  }
  await loadFiles();
}

async function upload(file: File, replace: boolean) {
  if (uploading.value) return;
  const lowerName = file.name.toLowerCase();
  if (replace && !lowerName.endsWith(".zip")) {
    message("请选择 .zip 测试数据包", { type: "warning" });
    return;
  }
  if (!replace && !/\.(in|out|zip)$/.test(lowerName)) {
    message("只能上传 .in、.out 或 .zip 文件", { type: "warning" });
    return;
  }

  if (replace) {
    try {
      await ElMessageBox.confirm(
        "整包上传会替换当前全部测试数据，确认继续？",
        "替换测试数据",
        { type: "warning" }
      );
    } catch {
      return;
    }
  }

  const data = new FormData();
  data.append("file", file);
  uploading.value = true;
  try {
    const res = replace
      ? await replaceTestDataZip(problemId.value, data)
      : await uploadTestDataFile(problemId.value, data);
    const result = res.data;
    if (result?.missing?.length) {
      message(
        `已导入 ${result.count} 个测试点，另有 ${result.missing.length} 个文件未成对`,
        { type: "warning" }
      );
    } else {
      message(result ? `已导入 ${result.count} 个测试点` : "文件已上传", {
        type: "success"
      });
    }
    await loadFiles();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    uploading.value = false;
  }
}

function onFileSelected(event: Event, replace: boolean) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (file) upload(file, replace);
}

async function preview(row: TestDataFile) {
  const seq = ++previewSeq;
  previewVisible.value = true;
  previewLoading.value = true;
  previewName.value = row.name;
  previewContent.value = "";
  previewTruncated.value = false;
  try {
    const res = await getTestDataContent(problemId.value, row.name);
    if (seq !== previewSeq) return;
    previewContent.value = res.data?.content ?? "";
    previewTruncated.value = Boolean(res.data?.truncated);
  } catch {
    if (seq === previewSeq) previewVisible.value = false;
  } finally {
    if (seq === previewSeq) previewLoading.value = false;
  }
}

async function downloadSelected() {
  if (!selected.value.length || downloading.value) return;
  downloading.value = true;
  try {
    const blob = await downloadTestDataFiles(
      problemId.value,
      selected.value.map(item => item.name)
    );
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${problemId.value}-testdata.zip`;
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    downloading.value = false;
  }
}

async function removeSelected() {
  if (!selected.value.length || deleting.value) return;
  try {
    await ElMessageBox.confirm(
      `确认删除选中的 ${selected.value.length} 个测试数据文件？`,
      "删除测试数据",
      { type: "warning" }
    );
  } catch {
    return;
  }
  deleting.value = true;
  try {
    const res = await deleteTestDataFiles(
      problemId.value,
      selected.value.map(item => item.name)
    );
    const failed = res.data?.failed ?? [];
    if (failed.length) {
      message(`${failed.length} 个文件删除失败`, { type: "warning" });
    } else {
      message("测试数据已删除", { type: "success" });
    }
    await loadFiles();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    deleting.value = false;
  }
}

function onPreviewClosed() {
  previewSeq += 1;
  previewContent.value = "";
  previewLoading.value = false;
}

onMounted(initialize);
</script>

<template>
  <div>
    <PureTableBar :columns="columns" @refresh="loadFiles">
      <template #title>
        <div class="flex items-center">
          <span>测试数据</span>
          <span class="ml-2 text-xs font-normal text-gray-400">
            {{ problemId }} {{ problemName }}
          </span>
        </div>
      </template>
      <template #buttons>
        <div class="flex flex-wrap gap-2">
          <el-button @click="router.push('/content/problems')"
            >返回列表</el-button
          >
          <template v-if="!remoteOJ">
            <el-button
              type="primary"
              :icon="Upload"
              :loading="uploading"
              @click="fileInput?.click()"
            >
              上传文件
            </el-button>
            <el-button
              :icon="Upload"
              :loading="uploading"
              @click="zipInput?.click()"
            >
              整包替换
            </el-button>
            <el-button
              :icon="Download"
              :disabled="!selected.length"
              :loading="downloading"
              @click="downloadSelected"
            >
              下载
            </el-button>
            <el-button
              type="danger"
              :icon="Delete"
              :disabled="!selected.length"
              :loading="deleting"
              @click="removeSelected"
            >
              删除
            </el-button>
          </template>
        </div>
        <input
          ref="fileInput"
          hidden
          type="file"
          accept=".in,.out,.zip"
          @change="event => onFileSelected(event, false)"
        />
        <input
          ref="zipInput"
          hidden
          type="file"
          accept=".zip"
          @change="event => onFileSelected(event, true)"
        />
      </template>
      <template v-slot="{ size, dynamicColumns }">
        <el-alert
          v-if="remoteOJ"
          :title="`该题使用 ${remoteOJ} 远程评测，平台不保存本地测试数据。`"
          type="info"
          show-icon
          :closable="false"
        />
        <PureTable
          v-else
          border
          adaptive
          table-layout="auto"
          :loading="loading"
          :size="size"
          :data="dataList"
          :columns="dynamicColumns"
          @selection-change="selected = $event"
        >
          <template #name="{ row }">
            <el-button link type="primary" @click="preview(row)">
              {{ row.name }}
            </el-button>
          </template>
          <template #size="{ row }">{{ formatSize(row.size) }}</template>
        </PureTable>
      </template>
    </PureTableBar>

    <el-dialog
      v-model="previewVisible"
      :title="previewName"
      width="760px"
      destroy-on-close
      @closed="onPreviewClosed"
    >
      <div v-loading="previewLoading" class="preview-box">
        <pre v-if="previewContent">{{ previewContent }}</pre>
        <el-empty
          v-else-if="!previewLoading"
          description="文件内容为空"
          :image-size="72"
        />
      </div>
      <p v-if="previewTruncated" class="mt-2 text-xs text-orange-500">
        文件较大，仅显示前 2 MB。
      </p>
    </el-dialog>
  </div>
</template>

<style scoped>
.preview-box {
  min-height: 180px;
  max-height: 60vh;
  overflow: auto;
  background: var(--el-fill-color-lighter);
  border: 1px solid var(--el-border-color-lighter);
}

.preview-box pre {
  min-width: max-content;
  padding: 16px;
  margin: 0;
  font-family: "Cascadia Mono", Consolas, monospace;
  font-size: 13px;
  line-height: 1.65;
  white-space: pre;
}
</style>
