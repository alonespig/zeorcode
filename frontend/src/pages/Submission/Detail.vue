<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import CodeViewer from "@/components/CodeViewer.vue";
import SubmissionCaseResults from "@/components/submission/SubmissionCaseResults.vue";
import SubmissionCompileOutput from "@/components/submission/SubmissionCompileOutput.vue";
import SubmissionSummaryCard from "@/components/submission/SubmissionSummaryCard.vue";
import { useSubmissionDetail } from "@/composables/useSubmissionDetail";
import { AcceptedCode, CompileErrorCode } from "@/constants/index";

const route = useRoute();
const router = useRouter();
const submissionId = computed(() => route.params.id);
const { detail, loading } = useSubmissionDetail(submissionId);

const passedCases = computed(() =>
  detail.value.caseResults.filter(({ status }) => status === AcceptedCode).length,
);
const isCompilationError = computed(() => detail.value.submission.status === CompileErrorCode);
const canViewCode = computed(() => detail.value.submission.canViewCode === true);
const sourceRestricted = computed(() => !canViewCode.value);

function openProblem(problemId) {
  if (problemId) router.push(`/problem/${problemId}`);
}

function openUser(userId) {
  if (userId) router.push({ name: "User", params: { id: userId } });
}
</script>

<template>
  <main v-loading="loading" class="submission-page">
    <section class="detail-card" aria-label="评测详情">
      <SubmissionSummaryCard
        embedded
        :user="detail.user"
        :problem="detail.problem"
        :submission="detail.submission"
        :passed-cases="passedCases"
        :total-cases="detail.caseResults.length"
        @problem-click="openProblem"
        @user-click="openUser"
      />

      <SubmissionCompileOutput
        v-if="isCompilationError"
        embedded
        :output="detail.submission.compileOutput"
        :restricted="sourceRestricted"
      />
      <SubmissionCaseResults
        v-else
        embedded
        :case-results="detail.caseResults"
        :status="detail.submission.status"
        :remote-oj="detail.problem.oj"
      />

      <CodeViewer
        v-if="canViewCode"
        embedded
        title="本次提交"
        show-line-count
        :code="detail.submission.code"
        :language="detail.submission.language"
      />
      <section v-else class="restricted-panel" aria-label="本次提交">
        <h2>本次提交</h2>
        <p>仅提交者和管理员可以查看源代码。</p>
      </section>
    </section>
  </main>
</template>

<style scoped>
.submission-page { width: min(1320px, calc(100% - 48px)); min-height: 420px; margin: 0 auto; padding: 0 0 56px; }
.detail-card { display: grid; gap: 12px; }
.restricted-panel { overflow: hidden; border: 0; border-radius: 0; background: #fff; box-shadow: none; }
.restricted-panel h2 { min-height: 58px; margin: 0; padding: 0 22px; border-bottom: 1px solid #e5e9f0; display: flex; align-items: center; color: #182230; font-size: 16px; font-weight: 640; }
.restricted-panel p { margin: 0; padding: 48px 22px; color: #7b8697; font-size: 13px; text-align: center; }
@media (max-width: 720px) {
  .submission-page { width: min(100% - 28px, 1320px); }
  .detail-card { gap: 8px; }
}
</style>
