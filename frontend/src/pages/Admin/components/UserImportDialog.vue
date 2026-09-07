<script setup>
import { computed, ref, shallowRef } from "vue";
import { ElMessage } from "element-plus";
import { Download, UploadFilled } from "@element-plus/icons-vue";
import { importAdminUsers } from "@/api/admin";

const emit = defineEmits(["imported"]);
const visible = defineModel({ type: Boolean, default: false });

const fileList = ref([]);
const selectedFile = shallowRef(null);
const importing = shallowRef(false);
const result = shallowRef(null);

const templateUrl = computed(
  () => `${import.meta.env.VITE_API_BASE_URL}/admin/users/import/template`,
);

const clearSelectedFile = () => {
  fileList.value = [];
  selectedFile.value = null;
};

const handleFileChange = (uploadFile) => {
  const file = uploadFile.raw;
  if (!file) return;
  if (!file.name.toLocaleLowerCase().endsWith(".xlsx")) {
    ElMessage.warning("仅支持 .xlsx 格式的 Excel 文件");
    clearSelectedFile();
    return;
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.warning("Excel 文件大小不能超过 5MB");
    clearSelectedFile();
    return;
  }
  selectedFile.value = file;
};

const handleFileRemove = () => {
  selectedFile.value = null;
};

const handleExceed = () => {
  ElMessage.warning("一次只能选择一个 Excel 文件");
};

const submitImport = async () => {
  if (!selectedFile.value || importing.value) return;
  importing.value = true;
  try {
    const res = await importAdminUsers(selectedFile.value);
    result.value = res.data;
    emit("imported", res.data);
  } catch (err) {
    console.error(err);
  } finally {
    importing.value = false;
  }
};

const resetDialog = () => {
  clearSelectedFile();
  result.value = null;
};
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
      <div class="import-notice">
        <div class="import-notice__title">按模板填写用户资料</div>
        <div>邮箱、学号、姓名和用户名均为必填项。</div>
        <div>新账号统一创建为普通用户，初始密码与学号相同。</div>
      </div>

      <div class="template-row">
        <div>
          <div class="template-row__title">先下载标准模板</div>
          <div class="template-row__hint">请保留工作表名称和第一行表头，每次最多导入 500 人</div>
        </div>
        <el-link :href="templateUrl" type="primary" download="用户导入模板.xlsx" :underline="false">
          <el-icon class="download-icon"><Download /></el-icon>
          下载模板
        </el-link>
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
        <div class="el-upload__text">将填写好的 Excel 拖到这里，或<em>点击选择</em></div>
        <template #tip>
          <div class="el-upload__tip">仅支持 .xlsx 文件，大小不超过 5MB。任一行失败时整批数据都会回滚。</div>
        </template>
      </el-upload>
    </template>

    <el-result v-else icon="success" title="用户导入完成" sub-title="用户列表已经刷新">
      <template #extra>
        <div class="result-stat">
          <div class="result-value">{{ result.created }}</div>
          <div class="result-label">新建账号</div>
        </div>
      </template>
    </el-result>

    <template #footer>
      <el-button v-if="!result" :disabled="importing" @click="visible = false">取消</el-button>
      <el-button v-if="!result" type="primary" :loading="importing" :disabled="!selectedFile" @click="submitImport">
        开始导入
      </el-button>
      <el-button v-else type="primary" @click="visible = false">完成</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.import-notice {
  margin-bottom: 20px;
  border: 1px solid #dbeafe;
  border-radius: 8px;
  padding: 12px 16px;
  background: #eff6ff;
  color: #606b7d;
  font-size: 14px;
  line-height: 24px;
}

.import-notice__title,
.template-row__title {
  color: #374151;
  font-weight: 500;
}

.template-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  padding: 12px 16px;
}

.template-row__hint {
  margin-top: 2px;
  color: #9ca3af;
  font-size: 12px;
}

.download-icon {
  margin-right: 4px;
}

.result-stat {
  min-width: 160px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 16px 12px;
  background: #f9fafb;
  text-align: center;
}

.result-value {
  color: #2563eb;
  font-size: 26px;
  font-weight: 600;
  line-height: 1.2;
}

.result-label {
  margin-top: 5px;
  color: #6b7280;
  font-size: 12px;
}
</style>
