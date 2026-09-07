<template>
  <el-dialog
    v-model="visible"
    title="导入学生名单"
    width="580px"
    style="max-width: calc(100vw - 32px)"
    :close-on-click-modal="!importing"
    :close-on-press-escape="!importing"
    :show-close="!importing"
    @closed="resetDialog"
  >
    <template v-if="!result">
      <div class="mb-5 rounded-lg border border-blue-100 bg-blue-50 px-4 py-3 text-sm leading-6 text-gray-600">
        <div class="font-medium text-gray-800">系统只按邮箱判断账号是否存在</div>
        <div>已有邮箱：直接将账号加入团队，不覆盖原有姓名、学号等资料。</div>
        <div>新邮箱：填写学号、姓名和用户名，系统会创建普通账号，初始密码为学号。</div>
      </div>

      <div class="mb-4 flex items-center justify-between rounded border border-gray-200 px-4 py-3">
        <div>
          <div class="text-sm font-medium text-gray-700">先下载标准模板</div>
          <div class="mt-0.5 text-xs text-gray-400">请保留工作表名称和第一行表头，每次最多导入 500 人</div>
        </div>
        <el-link :href="templateUrl" type="primary" download="学生名单导入模板.xlsx" :underline="false">
          <el-icon class="mr-1"><Download /></el-icon>
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

    <el-result v-else icon="success" title="学生名单导入完成" sub-title="成员列表已经刷新">
      <template #extra>
        <div class="grid w-full grid-cols-1 gap-3 sm:grid-cols-3">
          <div class="result-stat">
            <div class="result-value">{{ result.createdUsers }}</div>
            <div class="result-label">新建账号</div>
          </div>
          <div class="result-stat">
            <div class="result-value">{{ result.addedMembers }}</div>
            <div class="result-label">新增成员</div>
          </div>
          <div class="result-stat">
            <div class="result-value">{{ result.skippedMembers }}</div>
            <div class="result-label">已在团队</div>
          </div>
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

<script setup>
import { computed, ref, shallowRef } from "vue";
import { ElMessage } from "element-plus";
import { Download, UploadFilled } from "@element-plus/icons-vue";
import { importTeamStudents } from "@/api/team";

const props = defineProps({
  teamId: {
    type: [String, Number],
    required: true,
  },
});
const emit = defineEmits(["imported"]);
const visible = defineModel({ type: Boolean, default: false });

const fileList = ref([]);
const selectedFile = shallowRef(null);
const importing = shallowRef(false);
const result = shallowRef(null);

const templateUrl = computed(
  () => `${import.meta.env.VITE_API_BASE_URL}/team/${props.teamId}/member/import/template`,
);

const handleFileChange = (uploadFile) => {
  const file = uploadFile.raw;
  if (!file) return;
  if (!file.name.toLocaleLowerCase().endsWith(".xlsx")) {
    ElMessage.warning("仅支持 .xlsx 格式的 Excel 文件");
    fileList.value = [];
    selectedFile.value = null;
    return;
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.warning("Excel 文件大小不能超过 5MB");
    fileList.value = [];
    selectedFile.value = null;
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
    const res = await importTeamStudents(props.teamId, selectedFile.value);
    result.value = res.data;
    emit("imported", res.data);
  } catch (err) {
    console.error(err);
  } finally {
    importing.value = false;
  }
};

const resetDialog = () => {
  fileList.value = [];
  selectedFile.value = null;
  result.value = null;
};
</script>

<style scoped>
.result-stat {
  min-width: 120px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 14px 10px;
  background: #f9fafb;
  text-align: center;
}

.result-value {
  color: #2563eb;
  font-size: 24px;
  font-weight: 600;
  line-height: 1.2;
}

.result-label {
  margin-top: 5px;
  color: #6b7280;
  font-size: 12px;
}
</style>
