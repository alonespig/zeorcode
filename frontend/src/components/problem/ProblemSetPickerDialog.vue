<script setup>
import { computed, ref } from "vue";
import { Lock, Search } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import {
  getProblemSetDetail,
  getProblemSetList,
  unlockProblemSet,
} from "@/api/problemset";

const props = defineProps({
  modelValue: {
    type: Boolean,
    required: true,
  },
  existingProblemIds: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits(["update:modelValue", "add"]);

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});
const existingIdSet = computed(() => new Set(props.existingProblemIds.map(String)));

const query = ref("");
const problemSets = ref([]);
const page = ref(1);
const pageSize = 8;
const total = ref(0);
const listLoading = ref(false);
const detailLoading = ref(false);
const activeId = ref("");
const detail = ref(null);
const inviteCode = ref("");
const unlocking = ref(false);
let detailRequestId = 0;

const isExisting = (id) => existingIdSet.value.has(String(id));
const importableProblems = computed(() =>
  (detail.value?.problems || []).filter((problem) => !isExisting(problem.id))
);
const duplicateCount = computed(() =>
  (detail.value?.problems || []).filter((problem) => isExisting(problem.id)).length
);

const loadDetail = async (id) => {
  const requestId = ++detailRequestId;
  activeId.value = String(id);
  detail.value = null;
  inviteCode.value = "";
  detailLoading.value = true;
  try {
    const res = await getProblemSetDetail(id);
    if (requestId !== detailRequestId) return;
    detail.value = res.data;
  } catch (err) {
    console.error(err);
  } finally {
    if (requestId === detailRequestId) detailLoading.value = false;
  }
};

const loadProblemSets = async (targetPage = 1) => {
  page.value = targetPage;
  listLoading.value = true;
  try {
    const res = await getProblemSetList({
      page: targetPage,
      pageSize,
      q: query.value.trim() || undefined,
    });
    problemSets.value = res.data?.list || [];
    total.value = res.data?.total || 0;
    const activeStillVisible = problemSets.value.some((item) => String(item.id) === activeId.value);
    if (!activeStillVisible) {
      if (problemSets.value.length) await loadDetail(problemSets.value[0].id);
      else {
        detailRequestId += 1;
        activeId.value = "";
        detail.value = null;
        detailLoading.value = false;
      }
    }
  } catch (err) {
    console.error(err);
  } finally {
    listLoading.value = false;
  }
};

const handleOpen = () => {
  query.value = "";
  activeId.value = "";
  detail.value = null;
  inviteCode.value = "";
  loadProblemSets(1);
};

const handleClose = () => {
  detailRequestId += 1;
};

const submitUnlock = async () => {
  if (!inviteCode.value.trim()) {
    ElMessage.warning("请输入邀请码");
    return;
  }
  unlocking.value = true;
  try {
    await unlockProblemSet(activeId.value, inviteCode.value.trim());
    ElMessage.success("解锁成功");
    await loadDetail(activeId.value);
  } catch (err) {
    console.error(err);
  } finally {
    unlocking.value = false;
  }
};

const confirm = () => {
  if (!detail.value || detail.value.locked) {
    ElMessage.warning("请先选择可查看的题单");
    return;
  }
  if (!detail.value.problems?.length) {
    ElMessage.warning("该题单暂无题目");
    return;
  }
  if (!importableProblems.value.length) {
    ElMessage.warning("该题单中的题目已全部添加");
    return;
  }
  emit("add", importableProblems.value);
  visible.value = false;
};
</script>

<template>
  <el-dialog
    v-model="visible"
    title="从题单导入"
    width="min(960px, calc(100vw - 32px))"
    top="8vh"
    destroy-on-close
    @open="handleOpen"
    @closed="handleClose"
  >
    <div class="problemset-picker">
      <aside class="problemset-list">
        <div class="problemset-search">
          <el-input
            v-model="query"
            placeholder="搜索题单"
            clearable
            @keyup.enter="loadProblemSets(1)"
            @clear="loadProblemSets(1)"
          />
          <el-button :icon="Search" aria-label="搜索题单" @click="loadProblemSets(1)" />
        </div>

        <div v-loading="listLoading" class="problemset-list__body">
          <button
            v-for="item in problemSets"
            :key="item.id"
            type="button"
            class="problemset-item"
            :class="{ 'is-active': String(item.id) === activeId }"
            @click="loadDetail(item.id)"
          >
            <span class="problemset-item__title">
              {{ item.title }}
              <el-icon v-if="item.visibility === 1" class="problemset-item__lock"><Lock /></el-icon>
            </span>
            <span class="problemset-item__meta">{{ item.problemCount || 0 }} 题</span>
          </button>
          <el-empty v-if="!listLoading && !problemSets.length" description="暂无题单" :image-size="56" />
        </div>

        <el-pagination
          v-if="total > pageSize"
          class="problemset-pagination"
          size="small"
          layout="prev, next"
          :page-size="pageSize"
          :total="total"
          :current-page="page"
          @current-change="loadProblemSets"
        />
      </aside>

      <main v-loading="detailLoading" class="problemset-problems">
        <template v-if="detail">
          <div class="problemset-problems__heading">
            <strong>{{ detail.title }}</strong>
            <span>{{ detail.problemCount || 0 }} 题</span>
          </div>

          <div v-if="detail.locked" class="locked-problemset">
            <el-icon :size="28"><Lock /></el-icon>
            <p>该题单需要邀请码</p>
            <div class="unlock-form">
              <el-input
                v-model="inviteCode"
                placeholder="请输入邀请码"
                show-password
                @keyup.enter="submitUnlock"
              />
              <el-button type="primary" :loading="unlocking" @click="submitUnlock">解锁</el-button>
            </div>
          </div>

          <el-table v-else :data="detail.problems || []" height="410" class="problemset-table">
            <el-table-column type="index" label="#" width="54" align="center" />
            <el-table-column prop="id" label="题号" width="110" align="center" />
            <el-table-column prop="name" label="题目" min-width="220" show-overflow-tooltip />
            <el-table-column label="状态" width="88" align="center">
              <template #default="{ row }">
                <span v-if="isExisting(row.id)" class="existing-label">已添加</span>
              </template>
            </el-table-column>
            <template #empty>
              <el-empty description="该题单暂无题目" :image-size="64" />
            </template>
          </el-table>
        </template>
        <el-empty v-else-if="!detailLoading" description="请选择题单" :image-size="72" />
      </main>
    </div>

    <template #footer>
      <div class="picker-footer">
        <span v-if="detail && !detail.locked">
          可导入 {{ importableProblems.length }} 道
          <template v-if="duplicateCount">，已存在 {{ duplicateCount }} 道将自动跳过</template>
        </span>
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :disabled="!importableProblems.length" @click="confirm">
          导入全部题目
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<style scoped>
.problemset-picker {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  min-height: 460px;
  overflow: hidden;
  border: 1px solid #e5eaf1;
  border-radius: 4px;
}

.problemset-list {
  display: flex;
  min-width: 0;
  flex-direction: column;
  border-right: 1px solid #e5eaf1;
  background: #fafbfc;
}

.problemset-search {
  display: flex;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid #e5eaf1;
}

.problemset-search :deep(.el-button) {
  margin-left: 0;
}

.problemset-list__body {
  min-height: 0;
  flex: 1;
  padding: 6px;
}

.problemset-item {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 10px;
  padding: 11px 10px;
  border: 0;
  border-radius: 4px;
  color: #334155;
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.problemset-item:hover {
  background: #f0f5ff;
}

.problemset-item.is-active {
  color: #2563eb;
  background: #eaf2ff;
}

.problemset-item__title {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  font-size: 14px;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.problemset-item__lock {
  margin-left: 4px;
  color: #94a3b8;
  vertical-align: -2px;
}

.problemset-item__meta {
  flex: none;
  color: #94a3b8;
  font-size: 12px;
}

.problemset-pagination {
  justify-content: center;
  padding: 10px;
  border-top: 1px solid #e5eaf1;
}

.problemset-problems {
  min-width: 0;
  background: #fff;
}

.problemset-problems__heading {
  display: flex;
  height: 49px;
  align-items: center;
  gap: 10px;
  padding: 0 16px;
  border-bottom: 1px solid #e5eaf1;
}

.problemset-problems__heading strong {
  min-width: 0;
  overflow: hidden;
  color: #1f2937;
  font-size: 15px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.problemset-problems__heading span {
  color: #94a3b8;
  font-size: 12px;
}

.locked-problemset {
  display: flex;
  height: 410px;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  color: #94a3b8;
}

.locked-problemset p {
  margin: 10px 0 16px;
  color: #64748b;
}

.unlock-form {
  display: flex;
  width: min(360px, calc(100% - 32px));
  gap: 8px;
}

.existing-label {
  color: #94a3b8;
  font-size: 12px;
}

.picker-footer {
  display: flex;
  width: 100%;
  align-items: center;
}

.picker-footer > span {
  margin-right: auto;
  color: #64748b;
  font-size: 13px;
}

@media (max-width: 680px) {
  .problemset-picker {
    grid-template-columns: 1fr;
  }

  .problemset-list {
    max-height: 240px;
    border-right: 0;
    border-bottom: 1px solid #e5eaf1;
  }

  .problemset-picker {
    min-height: 0;
  }
}
</style>
