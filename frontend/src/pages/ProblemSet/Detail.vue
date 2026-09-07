<template>
  <div class="problemset-detail page-container" v-loading="loading">
    <template v-if="detail">
      <!-- 顶部只放标题/标签/题数/作者：不放整单进度，下面的题目列表本身就逐题标了过没过 -->
      <div class="f-panel mb-3.5 px-5 py-4">
        <h1 class="flex items-center gap-2 text-xl font-semibold text-gray-800">
          {{ detail.title }}
          <el-tooltip v-if="detail.visibility === 1" content="需要邀请码" placement="top">
            <el-icon class="text-gray-400" :size="16">
              <Lock />
            </el-icon>
          </el-tooltip>
        </h1>
        <div class="mt-2.5 flex flex-wrap items-center gap-2 text-[12.5px] text-gray-400">
          <el-tag v-for="t in detail.tags" :key="t.id" effect="light" :style="tagStyle(t)">{{ t.name }}</el-tag>
          <span v-if="detail.tags?.length" class="text-gray-300">·</span>
          <span>共 <b class="text-gray-700">{{ detail.problemCount }}</b> 题</span>
          <span class="text-gray-300">·</span>
          <span>{{ detail.author || '-' }}</span>
          <span class="text-gray-300">·</span>
          <span>更新于 {{ detail.updatedAt }}</span>
        </div>
      </div>

      <div v-if="detail.description" class="f-panel desc-panel mb-3.5 px-5 py-4">
        <!-- 用项目的 MdEditor 包装组件（全局自动注册），它统一了 mdHeadingId，
             不要直接引 md-editor-v3 —— 详见 CLAUDE.md -->
        <MdEditor :model-value="detail.description" preview-only editor-id="problemset-desc" />
      </div>

      <!-- 邀请码题单未解锁：标题描述照常可见，只把题目列表换成解锁框 -->
      <div v-if="detail.locked" class="f-panel px-5 py-12 text-center">
        <el-icon class="text-gray-300" :size="34">
          <Lock />
        </el-icon>
        <h3 class="mt-3 mb-1.5 text-base font-medium text-gray-700">该题单需要邀请码</h3>
        <p class="mb-4 text-[13px] text-gray-400">输入邀请码后即可查看全部 {{ detail.problemCount }} 道题目，只需输入一次</p>
        <div class="flex justify-center gap-2.5">
          <el-input v-model="inviteCode" placeholder="请输入邀请码" class="w-52!" @keyup.enter="submitUnlock" />
          <el-button type="primary" :loading="unlocking" @click="submitUnlock">解锁</el-button>
        </div>
      </div>

      <div v-else class="f-panel">
        <div class="border-b border-gray-100 px-5 py-3.5">
          <span class="text-[15px] font-semibold text-gray-700">题目列表</span>
        </div>
        <el-table :data="detail.problems" style="width: 100%">
          <el-table-column label="#" width="60" align="center">
            <template #default="{ $index }">
              <span class="text-[13px] text-gray-500">{{ $index + 1 }}</span>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="70" align="center">
            <template #default="{ row }">
              <!-- status: 1=已通过；其他非空值是首次提交的判题结果码(2=WA/3=TLE…)，
                   都算「交过没过」；null=没做过。不能只判 ===2，那样 TLE 等会漏显示 -->
              <el-icon v-if="row.status === 1" color="#2f9e44" :size="17">
                <Select />
              </el-icon>
              <el-icon v-else-if="row.status != null" color="#e5484d" :size="17">
                <CloseBold />
              </el-icon>
            </template>
          </el-table-column>

          <el-table-column prop="id" label="题号" width="110" align="center" />

          <el-table-column label="题目" min-width="220">
            <template #default="{ row }">
              <router-link class="font-medium text-blue-500 hover:text-blue-400" :to="`/problem/${row.id}`">
                {{ row.name }}
              </router-link>
            </template>
          </el-table-column>

          <el-table-column label="难度" width="90" align="center">
            <template #default="{ row }">
              <span class="text-[14px] font-medium" :class="diffTextClass(row.difficulty)">
                {{ DIFF_TEXT[row.difficulty] || '-' }}
              </span>
            </template>
          </el-table-column>

          <el-table-column label="算法标签" min-width="180">
            <template #default="{ row }">
              <div class="flex flex-wrap gap-1.5">
                <el-tag v-for="t in row.tags" :key="t.id" effect="light" :style="tagStyle(t)">{{ t.name }}</el-tag>
              </div>
            </template>
          </el-table-column>

          <el-table-column label="通过率" width="140" align="center">
            <template #default="{ row }">
              <div v-if="row.submitCount > 0" class="flex items-center gap-2">
                <div class="h-1.5 w-[74px] overflow-hidden rounded bg-gray-100">
                  <div class="h-full rounded"
                    :style="{ width: rate(row) + '%', background: diffColor(row.difficulty) }"></div>
                </div>
                <span class="w-9 text-[12.5px] tabular-nums text-gray-500">{{ rate(row) }}%</span>
              </div>
              <span v-else class="text-gray-300">-</span>
            </template>
          </el-table-column>

          <template #empty>
            <el-empty description="该题单还没有添加题目" :image-size="80" />
          </template>
        </el-table>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import { Lock, Select, CloseBold } from "@element-plus/icons-vue";
import { getProblemSetDetail, unlockProblemSet } from "@/api/problemset";
import { tagStyle } from "@/utils/tag";

const route = useRoute();
const detail = ref(null);
const loading = ref(false);
const inviteCode = ref("");
const unlocking = ref(false);

// 与题库列表保持一致的难度文案与配色
const DIFF_TEXT = { 1: "简单", 2: "中等", 3: "困难" };
const diffTextClass = (d) => ({ 1: "text-emerald-600", 2: "text-amber-500", 3: "text-red-500" }[d] || "text-gray-400");
const diffColor = (d) => ({ 1: "#2f9e44", 2: "#e8930c", 3: "#e5484d" }[d] || "#c0c4cc");

const rate = (row) => (row.submitCount ? Math.floor((100 * row.acceptedCount) / row.submitCount) : 0);

const loadDetail = async () => {
  loading.value = true;
  try {
    const res = await getProblemSetDetail(route.params.id);
    detail.value = res.data;
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const submitUnlock = async () => {
  if (!inviteCode.value.trim()) {
    ElMessage.warning("请输入邀请码");
    return;
  }
  unlocking.value = true;
  try {
    await unlockProblemSet(route.params.id, inviteCode.value.trim());
    ElMessage.success("解锁成功");
    inviteCode.value = "";
    await loadDetail();
  } catch (err) {
    console.error(err);
  } finally {
    unlocking.value = false;
  }
};

onMounted(loadDetail);
</script>

<style scoped>
.problemset-detail {
  padding-top: 8px;
  padding-bottom: 24px;
}

/* md-editor 预览里首个标题自带 margin-top，叠上面板 padding 后顶部空一大块 */
.desc-panel :deep(.md-editor-preview > :first-child) {
  margin-top: 0;
}
</style>
