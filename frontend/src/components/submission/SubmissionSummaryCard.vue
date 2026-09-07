<script setup>
import { computed } from "vue";
import { AcceptedCode, JUDGE_STATUS, JUDGE_STATUS_TEXT } from "@/constants/index";
import { formatMemory, formatTime } from "@/utils/format";

const props = defineProps({
  user: { type: Object, default: () => ({}) },
  problem: { type: Object, default: () => ({}) },
  submission: { type: Object, default: () => ({}) },
  passedCases: { type: Number, default: 0 },
  totalCases: { type: Number, default: 0 },
  embedded: { type: Boolean, default: false },
});

const emit = defineEmits(["problemClick", "userClick"]);
const isAccepted = computed(() => props.submission.status === AcceptedCode);
const isProcessing = computed(() =>
  props.submission.status === JUDGE_STATUS.PENDING || props.submission.status === JUDGE_STATUS.JUDGING,
);
const statusText = computed(() => JUDGE_STATUS_TEXT[props.submission.status] || `Status ${props.submission.status ?? "-"}`);
const statusTone = computed(() => isAccepted.value ? "success" : isProcessing.value ? "processing" : "error");
const timeText = computed(() => isProcessing.value || props.submission.status === JUDGE_STATUS.COMPILE_ERROR
  ? "—" : formatTime(props.submission.time));
const memoryText = computed(() => isProcessing.value || props.submission.status === JUDGE_STATUS.COMPILE_ERROR
  ? "—" : formatMemory(props.submission.memory));
const passText = computed(() => props.totalCases > 0 ? `${props.passedCases} / ${props.totalCases}` : "—");
const userInitial = computed(() => (props.user.name || "?").slice(0, 1).toUpperCase());
</script>

<template>
  <section
    class="summary-card"
    :class="[`summary-card--${statusTone}`, { 'summary-card--embedded': embedded }]"
    aria-labelledby="submission-status"
  >
    <div class="status-block">
      <span class="status-icon" aria-hidden="true">
        <svg v-if="isAccepted" viewBox="0 0 24 24" fill="none"><path d="m5 12.5 4.2 4.2L19 7" /></svg>
        <svg v-else viewBox="0 0 24 24" fill="none"><path d="m7 7 10 10M17 7 7 17" /></svg>
      </span>
      <div class="status-copy">
        <h1 id="submission-status" class="status-title">{{ statusText }}</h1>
        <span>评测编号 <span class="mono">#{{ submission.id || "—" }}</span></span>
        <span>提交于 {{ submission.createdAt || "—" }}</span>
      </div>
    </div>

    <dl class="facts">
      <div class="fact">
        <dt>提交者</dt>
        <dd>
          <button
            class="user-link"
            type="button"
            :disabled="!user.id"
            @click="emit('userClick', user.id)"
          >
            <el-avatar class="user-avatar" :size="28" :src="user.avatar || undefined">{{ userInitial }}</el-avatar>
            <span>{{ user.name || "—" }}</span>
          </button>
        </dd>
      </div>
      <div class="fact">
        <dt>题目</dt>
        <dd><button class="problem-link" type="button" @click="emit('problemClick', problem.id)">{{ problem.name || "—" }}</button></dd>
      </div>
      <div class="fact"><dt>语言</dt><dd class="mono">{{ submission.language || "—" }}</dd></div>
      <div class="fact"><dt>用时</dt><dd class="mono">{{ timeText }}</dd></div>
      <div class="fact"><dt>内存</dt><dd class="mono">{{ memoryText }}</dd></div>
      <div class="fact"><dt>通过</dt><dd class="mono">{{ passText }}</dd></div>
    </dl>
  </section>
</template>

<style scoped>
.summary-card { --status-color: #d63c3c; --status-soft: #fff0ef; position: relative; overflow: hidden; display: grid; grid-template-columns: 310px minmax(0, 1fr); min-height: 154px; border: 1px solid #d7dde6; border-radius: 10px; background: #fff; box-shadow: 0 8px 28px rgba(24, 34, 48, 0.06); }
.summary-card::before { content: ""; position: absolute; inset: 0 auto 0 0; width: 4px; background: var(--status-color); }
.summary-card--success { --status-color: #208b4e; --status-soft: #eaf7ef; }
.summary-card--processing { --status-color: #1769e0; --status-soft: #edf5ff; }
.summary-card--embedded { border: 0; border-radius: 0; box-shadow: none; }
.status-block { display: flex; align-items: center; gap: 17px; padding: 28px 30px; border-right: 1px solid #e5e9f0; }
.status-icon { flex: 0 0 auto; width: 54px; height: 54px; border-radius: 50%; display: grid; place-items: center; color: var(--status-color); background: var(--status-soft); }
.status-icon svg { width: 27px; height: 27px; }
.status-icon path { stroke: currentColor; stroke-width: 2.2; stroke-linecap: round; stroke-linejoin: round; }
.status-copy { display: flex; min-width: 0; flex-direction: column; gap: 4px; color: #465366; font-size: 13px; font-weight: 500; }
.status-title { margin: 0 0 4px; overflow-wrap: anywhere; color: var(--status-color); font-size: 24px; font-weight: 650; line-height: 1.2; }
.mono { font-family: "Cascadia Code", "JetBrains Mono", Consolas, monospace; font-variant-numeric: tabular-nums; }
.facts { display: grid; grid-template-columns: repeat(6, minmax(72px, 1fr)); align-items: center; margin: 0; padding: 24px 12px; }
.fact { min-width: 0; margin: 0; padding: 3px 12px; border-right: 1px solid #e5e9f0; }
.fact:last-child { border-right: 0; }
.fact dt { margin-bottom: 9px; color: #465366; font-size: 14px; font-weight: 520; }
.fact dd { overflow: hidden; margin: 0; color: #172033; font-size: 15px; font-weight: 520; line-height: 1.4; text-overflow: ellipsis; white-space: nowrap; }
.user-link { max-width: 100%; overflow: hidden; padding: 0; border: 0; display: inline-flex; align-items: center; gap: 8px; color: #172033; background: transparent; font: inherit; cursor: pointer; }
.user-link span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.user-link:hover { color: #1769e0; }
.user-link:disabled { color: #172033; cursor: default; }
.user-link:focus-visible { outline: 3px solid rgba(23, 105, 224, 0.2); outline-offset: 3px; }
.user-avatar { flex: 0 0 auto; background: #e8eef7; color: #465366; font-size: 12px; }
.problem-link { max-width: 100%; overflow: hidden; padding: 0; border: 0; color: #1769e0; background: transparent; cursor: pointer; text-overflow: ellipsis; white-space: nowrap; }
.problem-link:hover { color: #0f55bb; }
.problem-link:focus-visible { outline: 3px solid rgba(23, 105, 224, 0.2); outline-offset: 3px; }
@media (max-width: 1080px) {
  .summary-card { grid-template-columns: 235px minmax(0, 1fr); }
  .status-block { padding: 24px 20px; }
  .status-icon { width: 48px; height: 48px; }
  .fact { padding-inline: 8px; }
}
@media (max-width: 720px) {
  .summary-card { grid-template-columns: 1fr; }
  .status-block { border-right: 0; border-bottom: 1px solid #e5e9f0; }
  .facts { grid-template-columns: repeat(2, minmax(0, 1fr)); padding: 18px 8px; row-gap: 18px; }
  .fact:nth-child(even) { border-right: 0; }
}
</style>
