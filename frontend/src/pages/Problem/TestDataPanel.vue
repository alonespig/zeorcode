<template>
  <div class="panel page-container">
    <div class="header">
      <div class="title">测试数据</div>

      <div class="actions">
        <el-button @click="goBackToProblem">返回题目</el-button>

        <!-- 远程题无本地测试数据，隐藏上传/下载/删除 -->
        <template v-if="ready && !isRemote">
          <el-upload
            v-model:file-list="testDataUploadFiles"
            :action="uploadUrl"
            :with-credentials="true"
            accept=".in,.out,.zip"
            multiple
            :show-file-list="false"
            :before-upload="beforeTestDataUpload"
            :on-success="onUploadSuccess"
            :on-error="onUploadError"
          >
            <el-button class="btn" type="primary">
              <span class="iconfont icon-shangchuan"></span>
              <span>上传文件</span>
            </el-button>
          </el-upload>

          <el-upload
            v-model:file-list="zipUploadFiles"
            :http-request="zipUploadReq"
            accept=".zip"
            :show-file-list="false"
          >
            <el-button class="btn" type="primary" plain>
              <span class="iconfont icon-shangchuan"></span>
              <span>上传测试包(整包替换)</span>
            </el-button>
          </el-upload>

          <el-button class="btn" :disabled="selected.length === 0" @click="download">
            <span class="iconfont icon-xiazai"></span>
            <span>下载</span>
          </el-button>

          <el-button class="btn" type="danger" :disabled="selected.length === 0" @click="remove">
            <span class="iconfont icon-shanchu"></span>
            <span>删除</span>
          </el-button>
        </template>
      </div>
    </div>

    <div v-if="uploadFiles.length" class="upload-progress-list" aria-live="polite">
      <div v-for="file in uploadFiles" :key="file.uid" class="upload-progress-item">
        <div class="upload-progress-meta">
          <span class="upload-file-name" :title="file.name">{{ file.name }}</span>
          <span class="upload-status-text">{{ uploadStatusText(file) }}</span>
        </div>
        <el-progress
          :percentage="uploadPercentage(file)"
          :status="uploadProgressStatus(file)"
          :stroke-width="6"
          :show-text="false"
        />
      </div>
    </div>

    <!-- 远程题：无本地测试点 -->
    <el-empty v-if="ready && isRemote" :image-size="90">
      <template #description>
        <p class="empty-title">远程评测题目</p>
        <p class="empty-sub">本题由远程 OJ（{{ oj }}）评测，平台不保存测试点数据</p>
      </template>
    </el-empty>

    <!-- 本地题：测试点文件 -->
    <template v-else-if="ready">
      <el-table :data="files" @selection-change="selected = $event">
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" label="文件名">
          <template #default="{ row }">
            <button type="button" class="file-name-link" @click="previewFile(row)">
              {{ row.name }}
            </button>
          </template>
        </el-table-column>
        <el-table-column label="大小" width="120">
          <template #default="{ row }">
            <span class="file-size">{{ format(row.size) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <el-dialog
      v-model="previewVisible"
      :title="previewName"
      width="min(760px, calc(100vw - 32px))"
      destroy-on-close
    >
      <div v-loading="previewLoading" class="file-preview">
        <pre v-if="previewContent" class="file-preview-content">{{ previewContent }}</pre>
        <el-empty v-else-if="!previewLoading" description="文件内容为空" :image-size="72" />
      </div>
      <div v-if="previewTruncated" class="preview-hint">
        文件内容较大，仅显示前 2 MB
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import {
  fileList,
  fileDelete,
  getTestDataContent,
  uploadTestcaseZip,
  getProblem,
} from '@/api/problems'
import { ref, computed, onMounted } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'

// 统一的后端地址（和 utils/request.js 保持一致）
const API_BASE = import.meta.env.VITE_API_BASE_URL

const files = ref([])
const selected = ref([])
const testDataUploadFiles = ref([])
const zipUploadFiles = ref([])
const previewVisible = ref(false)
const previewLoading = ref(false)
const previewName = ref('')
const previewContent = ref('')
const previewTruncated = ref(false)
const oj = ref('')            // 空=本地题；非空=远程题（如 "LibreOJ"）
const ready = ref(false)      // 题目信息拉到后再决定渲染哪种视图
const isRemote = computed(() => !!oj.value)
const route = useRoute()
const router = useRouter()

const problemId = computed(() => route.params.id)

const uploadUrl = computed(() => `${API_BASE}/problems/${problemId.value}/testdata`)
const uploadFiles = computed(() => [
  ...testDataUploadFiles.value,
  ...zipUploadFiles.value,
])

async function loadFiles() {
  const res = await fileList(problemId.value)
  files.value = res.data || []
}

function onUploadSuccess(_response, uploadFile) {
  ElMessage.success('上传成功')
  loadFiles()
  scheduleUploadRemoval(uploadFile.uid, testDataUploadFiles)
}

function onUploadError() {
  ElMessage.error('上传失败')
}

function beforeTestDataUpload(file) {
  const name = file.name || ''
  const allowed = name.endsWith('.in') || name.endsWith('.out') || name.endsWith('.zip')
  if (!allowed) {
    ElMessage.warning('只能上传 .in、.out 或 .zip 文件')
  }
  return allowed
}

async function previewFile(file) {
  previewVisible.value = true
  previewLoading.value = true
  previewName.value = file.name
  previewContent.value = ''
  previewTruncated.value = false

  try {
    const res = await getTestDataContent(problemId.value, file.name)
    previewContent.value = res.data?.content || ''
    previewTruncated.value = !!res.data?.truncated
  } catch (err) {
    console.error(err)
    previewVisible.value = false
  } finally {
    previewLoading.value = false
  }
}

function uploadPercentage(file) {
  return Math.min(100, Math.max(0, Math.round(file.percentage || 0)))
}

function uploadProgressStatus(file) {
  if (file.status === 'success') return 'success'
  if (file.status === 'fail') return 'exception'
  return undefined
}

function uploadStatusText(file) {
  if (file.status === 'success') return '上传完成'
  if (file.status === 'fail') return '上传失败'
  if (file.status === 'ready') return '等待上传'
  return `上传中 ${uploadPercentage(file)}%`
}

function scheduleUploadRemoval(uid, fileListRef) {
  window.setTimeout(() => {
    fileListRef.value = fileListRef.value.filter(file => file.uid !== uid)
  }, 1500)
}

// zip 上传走 axios（统一鉴权 + 错误提示），拿回导入的测试点数
async function zipUploadReq(options) {
  const { file, onProgress } = options
  const fd = new FormData()
  fd.append('file', file)
  try {
    const res = await uploadTestcaseZip(problemId.value, fd, {
      onUploadProgress(event) {
        if (!event.total) return
        onProgress({ percent: Math.round((event.loaded / event.total) * 100) })
      },
    })
    const { count = 0, missing = [] } = res.data || {}
    if (missing.length) {
      ElMessage.warning(`已导入 ${count} 个测试点；${missing.length} 个未成对文件已忽略`)
    } else {
      ElMessage.success(`已导入 ${count} 个测试点`)
    }
    loadFiles()
    scheduleUploadRemoval(file.uid, zipUploadFiles)
  } catch (err) {
    console.error(err) // request 拦截器已弹出错误
    throw err
  }
}

function goBackToProblem() {
  router.push(`/problem/${problemId.value}`)
}

function download() {
  const names = selected.value.map(f => f.name).join(',')
  window.open(`${API_BASE}/problems/${problemId.value}/testdata/download?files=${encodeURIComponent(names)}`)
}

function remove() {
  ElMessageBox.confirm(
    `确认删除 ${selected.value.length} 个文件？`,
    '警告',
    { type: 'warning' }
  ).then(async () => {
    try {
      await fileDelete(problemId.value, {
        files: selected.value.map(f => f.name),
      })
      ElMessage.success('删除成功')
      selected.value = []
      loadFiles()
    } catch (err) {
      console.error(err)
    }
  })
}

function format(size) {
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' K'
  return (size / 1024 / 1024).toFixed(1) + ' M'
}

async function init() {
  try {
    const res = await getProblem(problemId.value)
    oj.value = res.data?.oj || ''
  } catch (err) {
    console.error(err)
  }
  // 本地题才拉测试点文件；远程题无本地数据，直接进空状态
  if (!isRemote.value) await loadFiles()
  ready.value = true
}

onMounted(init)
</script>

<style scoped>
.btn .iconfont {
  margin-right: 2px;
}

.panel {
  width: 880px;
  background: #fff;
  padding: 16px;
}

.header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}

.upload-progress-list {
  display: grid;
  gap: 10px;
  margin-bottom: 14px;
  padding: 12px 14px;
  border: 1px solid var(--color-border-soft);
  border-radius: var(--radius-base);
  background: var(--color-surface);
}

.upload-progress-item {
  min-width: 0;
}

.upload-progress-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 6px;
  font-size: 13px;
}

.upload-file-name {
  min-width: 0;
  overflow: hidden;
  color: var(--color-text);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upload-status-text {
  flex: none;
  color: #909399;
}

.file-name-link,
.file-size {
  font-family: "Cascadia Mono", "Segoe UI Mono", Consolas, monospace;
  font-size: 14px;
  font-variant-numeric: tabular-nums;
}

.file-name-link {
  padding: 0;
  border: 0;
  color: var(--color-primary);
  background: transparent;
  cursor: pointer;
  transition: color 0.15s ease;
}

.file-name-link:hover,
.file-name-link:focus-visible {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.file-name-link:focus-visible {
  outline: 2px solid color-mix(in srgb, var(--color-primary) 35%, transparent);
  outline-offset: 3px;
  border-radius: var(--radius-small);
}

.file-preview {
  min-height: 180px;
  max-height: 60vh;
  overflow: auto;
  border: 1px solid var(--color-border-soft);
  border-radius: var(--radius-base);
  background: #f7f8fa;
}

.file-preview-content {
  min-width: max-content;
  margin: 0;
  padding: 16px;
  color: var(--color-text);
  font-family: "Cascadia Mono", "Segoe UI Mono", Consolas, monospace;
  font-size: 13px;
  line-height: 1.65;
  white-space: pre;
  tab-size: 4;
}

.preview-hint {
  margin-top: 10px;
  color: #909399;
  font-size: 13px;
}

.title {
  font-size: 16px;
  font-weight: 600;
}

.actions {
  display: flex;
  gap: 10px;
}

.empty-title {
  font-size: 15px;
  font-weight: 600;
  color: #4b5563;
  margin: 0 0 4px;
}

.empty-sub {
  font-size: 13px;
  color: #9aa1ab;
  margin: 0;
}
</style>
