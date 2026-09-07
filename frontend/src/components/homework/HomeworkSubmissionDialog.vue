<script setup>
import { computed } from "vue";
import {
  AcceptedCode,
  JUDGE_STATUS_CLASS,
  JUDGE_STATUS_TEXT,
  PendingCode,
} from "@/constants/index";
import { formatMemory, formatTime } from "@/utils/format";
import CodeViewer from "@/components/CodeViewer.vue";

const visible = defineModel({ type: Boolean, default: false });
const props = defineProps({
  submission: {
    type: Object,
    default: () => ({}),
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const dialogTitle = computed(() => props.submission.id ? `提交详情 #${props.submission.id}` : "提交详情");
const statusText = computed(() => JUDGE_STATUS_TEXT[props.submission.status] || `状态 ${props.submission.status}`);
const statusClass = computed(() => JUDGE_STATUS_CLASS[props.submission.status] || "text-gray-500");
const caseResults = computed(() => props.submission.caseResults || []);
const passedCaseCount = computed(() => caseResults.value.filter(({ status }) => status === AcceptedCode).length);
const emptyCaseText = computed(() => props.submission.status === PendingCode
  ? "正在评测，测试点结果将在评测完成后显示"
  : "暂无测试点信息");
const memberName = computed(() => {
  return props.submission.realName || props.submission.username || "-";
});
const memberTitle = computed(() => props.submission.studentNo
  ? `${memberName.value}（${props.submission.studentNo}）`
  : memberName.value);
const memberInitial = computed(() => memberName.value.slice(0, 1).toUpperCase());
</script>

<template>
  <el-dialog v-model="visible" :title="dialogTitle" width="min(900px, 92vw)" :close-on-click-modal="false"
    destroy-on-close>
    <div v-loading="loading" class="min-h-56">
      <template v-if="submission.id">
        <section class="rounded border border-gray-200 bg-gray-50 px-4 py-3" aria-label="提交摘要">
          <div class="mb-3 flex flex-wrap items-center gap-2 border-b border-gray-200 pb-3">
            <strong class="text-lg font-semibold" :class="statusClass">{{ statusText }}</strong>
            <span class="text-sm text-gray-400">提交号 #{{ submission.id }}</span>
            <el-tag v-if="!submission.inWindow" size="small" type="info" effect="plain">时间窗外，不计入排行榜</el-tag>
          </div>

          <dl class="grid grid-cols-2 gap-x-6 gap-y-3 text-sm md:grid-cols-3">
            <div class="min-w-0">
              <dt class="text-xs text-gray-400">成员</dt>
              <dd class="mt-1 flex min-w-0 items-center gap-2 text-gray-700" :title="memberTitle">
                <el-avatar :size="30" :src="submission.avatar || undefined"
                  class="shrink-0 bg-blue-50 text-xs font-semibold text-blue-600">
                  {{ memberInitial }}
                </el-avatar>
                <span class="min-w-0">
                  <span class="block truncate">{{ memberName }}</span>
                  <span class="block truncate text-xs text-gray-400">{{ submission.studentNo || "暂无学号" }}</span>
                </span>
              </dd>
            </div>
            <div class="min-w-0">
              <dt class="text-xs text-gray-400">题目</dt>
              <dd class="mt-1 truncate text-gray-700" :title="`${submission.problemId} ${submission.problemName}`">
                <span class="font-mono text-blue-500">{{ submission.problemId }}</span>
                {{ submission.problemName }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-gray-400">语言</dt>
              <dd class="mt-1 font-mono text-gray-700">{{ submission.language || "-" }}</dd>
            </div>
            <div>
              <dt class="text-xs text-gray-400">得分</dt>
              <dd class="mt-1 font-mono font-semibold tabular-nums"
                :class="submission.score >= 100 ? 'text-emerald-600' : submission.score > 0 ? 'text-amber-600' : 'text-gray-500'">
                {{ submission.score ?? "-" }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-gray-400">用时 / 内存</dt>
              <dd class="mt-1 font-mono tabular-nums text-gray-700">
                {{ formatTime(submission.timeUsed) }} / {{ formatMemory(submission.memoryUsed) }}
              </dd>
            </div>
            <div>
              <dt class="text-xs text-gray-400">提交时间</dt>
              <dd class="mt-1 font-mono text-gray-700">{{ submission.createdAt || "-" }}</dd>
            </div>
          </dl>
        </section>

        <section class="mt-5" aria-label="测试点">
          <div class="mb-2 flex items-center justify-between border-b border-gray-200 pb-2">
            <h3 class="text-[15px] font-semibold text-gray-700">测试点</h3>
            <span v-if="caseResults.length" class="text-xs tabular-nums text-gray-500">
              通过 {{ passedCaseCount }} / {{ caseResults.length }}
            </span>
          </div>

          <div v-if="submission.oj"
            class="rounded border border-blue-100 bg-blue-50 px-4 py-3 text-sm text-blue-700">
            远程评测（{{ submission.oj }}）由原 OJ 完成，不提供测试点明细。
          </div>
          <div v-else-if="caseResults.length"
            class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
            <article v-for="item in caseResults" :key="item.id"
              class="flex min-w-0 items-center gap-3 rounded border border-gray-200 bg-white px-3 py-2.5 text-sm">
              <span class="shrink-0 font-mono font-semibold text-gray-500">#{{ item.id }}</span>
              <span class="min-w-0 truncate font-medium"
                :class="JUDGE_STATUS_CLASS[item.status] || 'text-gray-500'">
                {{ JUDGE_STATUS_TEXT[item.status] || `状态 ${item.status}` }}
              </span>
              <span class="ml-auto shrink-0 font-mono text-xs tabular-nums text-gray-500">
                {{ formatTime(item.time) }} · {{ formatMemory(item.memory) }}
              </span>
            </article>
          </div>
          <div v-else class="rounded border border-dashed border-gray-200 bg-gray-50 px-4 py-5 text-center text-sm text-gray-400">
            {{ emptyCaseText }}
          </div>
        </section>

        <section class="mt-5" aria-label="提交代码">
          <div class="mb-2 flex items-center justify-between">
            <h3 class="text-[15px] font-semibold text-gray-700">源代码</h3>
            <span class="text-xs text-gray-400">只读</span>
          </div>
          <div v-if="submission.code" class="max-h-[55vh] overflow-auto rounded">
            <CodeViewer :code="submission.code" :language="submission.language" />
          </div>
          <el-empty v-else description="暂无可查看的代码" :image-size="70" />
        </section>
      </template>
    </div>

    <template #footer>
      <el-button @click="visible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>
