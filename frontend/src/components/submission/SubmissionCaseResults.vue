<script setup>
import { computed } from "vue";
import { AcceptedCode, JUDGE_STATUS, JUDGE_STATUS_CLASS, JUDGE_STATUS_TEXT } from "@/constants/index";
import { formatMemory, formatTime } from "@/utils/format";

const props = defineProps({
  caseResults: { type: Array, default: () => [] },
  status: { type: Number, default: JUDGE_STATUS.PENDING },
  remoteOj: { type: String, default: "" },
  embedded: { type: Boolean, default: false },
});

const passedCases = computed(() => props.caseResults.filter(({ status }) => status === AcceptedCode).length);
const totalCases = computed(() => props.caseResults.length);
const passRate = computed(() => totalCases.value > 0 ? Math.round((passedCases.value / totalCases.value) * 100) : 0);
const isProcessing = computed(() => props.status === JUDGE_STATUS.PENDING || props.status === JUDGE_STATUS.JUDGING);
const overviewText = computed(() => {
  if (totalCases.value === 0) return "暂无测试点结果";
  if (passedCases.value === totalCases.value) return "已通过全部测试点";
  return `${passedCases.value} 个测试点通过，${totalCases.value - passedCases.value} 个未通过`;
});
const statusText = (status) => JUDGE_STATUS_TEXT[status] || `Status ${status}`;
const statusClass = (status) => JUDGE_STATUS_CLASS[status] || "unknown";
</script>

<template>
  <section class="result-panel" :class="{ 'result-panel--embedded': embedded }" aria-labelledby="case-result-title">
    <header class="panel-head"><h2 id="case-result-title">测试点详情</h2></header>
    <div v-if="remoteOj" class="empty-state">远程评测（{{ remoteOj }}）由原 OJ 完成，不提供测试点明细。</div>
    <div v-else-if="isProcessing && totalCases === 0" class="empty-state">正在评测，测试点结果将在完成后显示。</div>
    <div v-else-if="totalCases === 0" class="empty-state">暂无测试点信息。</div>
    <template v-else>
      <div class="case-overview">
        <div class="overview-line">
          <strong>{{ overviewText }}</strong>
          <span class="mono">{{ passedCases }} / {{ totalCases }}</span>
        </div>
        <div class="progress" role="progressbar" aria-label="测试点通过率" aria-valuemin="0" aria-valuemax="100" :aria-valuenow="passRate">
          <span class="progress-fill" :style="{ width: `${passRate}%` }"></span>
        </div>
      </div>
      <div class="table-wrap">
        <table class="case-table">
          <thead><tr><th>测试点</th><th>状态</th><th>用时</th><th>内存</th></tr></thead>
          <tbody>
            <tr v-for="item in caseResults" :key="item.id">
              <td class="case-id">#{{ item.id }}</td>
              <td><span class="case-status" :class="statusClass(item.status)">{{ statusText(item.status) }}</span></td>
              <td class="mono">{{ formatTime(item.time) }}</td>
              <td class="mono">{{ formatMemory(item.memory) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="panel-foot">共 {{ totalCases }} 个测试点</footer>
    </template>
  </section>
</template>

<style scoped>
.result-panel { overflow: hidden; border: 1px solid #d7dde6; border-radius: 10px; background: #fff; box-shadow: 0 3px 14px rgba(24, 34, 48, 0.035); }
.result-panel--embedded { border: 0; border-radius: 0; box-shadow: none; }
.panel-head { min-height: 58px; padding: 0 22px; border-bottom: 1px solid #e5e9f0; display: flex; align-items: center; }
.panel-head h2 { margin: 0; color: #182230; font-size: 16px; font-weight: 640; }
.case-overview { padding: 20px 22px 14px; }
.overview-line { display: flex; align-items: baseline; justify-content: space-between; gap: 20px; margin-bottom: 12px; font-size: 13px; }
.overview-line strong { color: #208b4e; font-weight: 600; }
.overview-line span { color: #687386; }
.progress { height: 6px; overflow: hidden; border-radius: 99px; background: #edf0f4; }
.progress-fill { display: block; height: 100%; border-radius: inherit; background: #208b4e; transition: width 240ms ease; }
.table-wrap { overflow-x: auto; padding: 0 22px 18px; }
.case-table { width: 100%; min-width: 540px; border-collapse: collapse; color: #3a4656; font-size: 13px; font-variant-numeric: tabular-nums; }
.case-table th { height: 40px; color: #687386; background: #f8fafc; font-size: 12px; font-weight: 540; text-align: left; }
.case-table th, .case-table td { padding: 0 13px; border-bottom: 1px solid #e5e9f0; }
.case-table td { height: 43px; }
.case-table tbody tr:last-child td { border-bottom: 0; }
.case-table tbody tr:hover td { background: #fafcff; }
.case-id { color: #5f6b7d; font-family: "Cascadia Code", "JetBrains Mono", Consolas, monospace; }
.case-status { font-weight: 560; }
.mono { font-family: "Cascadia Code", "JetBrains Mono", Consolas, monospace; font-variant-numeric: tabular-nums; }
.panel-foot { padding: 12px 22px; border-top: 1px solid #e5e9f0; color: #687386; background: #fcfdfe; font-size: 12px; }
.empty-state { padding: 48px 22px; color: #7b8697; font-size: 13px; text-align: center; }
@media (prefers-reduced-motion: reduce) { .progress-fill { transition: none; } }
</style>
