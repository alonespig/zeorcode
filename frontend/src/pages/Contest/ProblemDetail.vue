<template>
  <div class="mx-auto max-w-[1340px] px-6 pb-5 max-md:px-3">
    <div class="flex items-start gap-5 max-lg:flex-col max-lg:items-stretch max-lg:gap-4">
      <!-- 左：题面 + 编辑器 -->
      <div class="flex min-w-0 flex-1 flex-col gap-5 max-lg:order-3 max-lg:w-full">
        <ContestProblem :problem="problem" />

        <div class="f-panel px-6 py-4">
          <div class="mb-3 flex items-center">
            <span class="text-sm text-gray-600">语言：</span>
            <el-select v-model="language" style="width: 120px">
              <el-option v-for="item in languageList" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
          </div>
          <!-- 高度随视口收缩，避免小屏上编辑器独占一整屏 -->
          <Codemirror v-model="code" :extensions="extensions" :style="{ height: 'clamp(240px, 40vh, 360px)' }" />
          <div class="mt-4 flex">
            <el-button type="primary" class="ml-auto" :loading="submitting" @click="handleSubmit">
              <span class="iconfont icon-7 mr-1.5"></span>提交代码
            </el-button>
          </div>
        </div>
      </div>

      <!-- 右：比赛信息 + 题目切换。窄屏下 aside 自身 display:contents「溶解」，
           内部各块直接成为外层单列 flex 的子项：比赛信息与切题器上移到题面之上
           （赛中必须随时可达），最近提交下沉到题面之下。 -->
      <aside class="flex w-72 shrink-0 flex-col gap-4 max-lg:contents">
        <div class="f-panel overflow-hidden max-lg:order-1">
          <!-- 比赛名称同时作为返回题目列表的入口 -->
          <button
            type="button"
            class="block w-full truncate px-4 py-3 text-center text-[16px] font-semibold text-gray-800 transition-colors hover:text-blue-500 hover:underline"
            title="返回比赛题目列表"
            @click="goList"
          >
            {{ contestData?.name || '比赛' }}
          </button>
          <div
            class="flex items-center justify-center gap-2 border-t border-gray-100 bg-gray-50 px-4 py-3 text-[13px] text-gray-500"
          >
            <span class="h-2 w-2 shrink-0 rounded-full" :class="contestStatusDotClass"></span>
            <template v-if="cdLabel">
              <span>{{ cdLabel }}</span>
              <strong class="font-mono font-semibold tabular-nums text-gray-700">{{ statusTime }}</strong>
            </template>
            <span v-else>{{ statusText }}</span>
          </div>
          <div class="grid grid-cols-2 gap-2 border-t border-gray-100 p-3">
            <el-button plain class="!ml-0 !h-9 w-full" @click="goRank">排行榜</el-button>
            <el-button plain class="!ml-0 !h-9 w-full" @click="goContestHome">比赛首页</el-button>
          </div>
        </div>

        <div class="f-panel max-lg:order-2">
          <div class="border-b border-gray-100 px-4 py-2.5 text-sm text-gray-700">题目列表</div>
          <div class="flex flex-wrap gap-2 p-3.5">
            <span v-for="p in problemList" :key="p.label" @click="goProblem(p.label)"
              class="flex h-9 w-9 cursor-pointer items-center justify-center rounded border font-mono text-sm transition-colors"
              :class="switchClass(p)">
              {{ p.label }}
            </span>
            <span v-if="!problemList.length" class="py-1 text-sm text-gray-400">暂无题目</span>
          </div>
        </div>

        <RecentSubmissions
          class="max-lg:order-4"
          :contest-id="route.params.id"
          :problem-id="curInList?.problemID"
          @click-id="goSubmissionDetail"
          @view-all="goSubmission"
        />
      </aside>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { Codemirror } from "vue-codemirror";
import { EditorView } from "@codemirror/view";
import { getContestProblem, getContestProblemList, getContestDesc, submit } from "@/api/contest";
import { useContestStatus } from "@/hooks/useContestStatus";
import { useLanguages } from "@/hooks/useLanguages";
import { getEditorLanguageExtension } from "@/utils/editorLanguage";
import { CONTEST_STATUS, AcceptedCode, WrongAnswerCode } from "@/constants/index";
import { useUserStore } from "@/stores/user";
import ContestProblem from "@/components/problem/ContestProblem.vue";
import RecentSubmissions from "@/components/contest/RecentSubmissions.vue";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

const contestData = ref(null);
const problem = ref({});
const problemList = ref([]);
const code = ref("");
const language = ref();
const submitting = ref(false);

const { languageList, loadLanguages } = useLanguages();
const { status, statusText, statusTime } = useContestStatus(contestData);

const cdLabel = computed(() => {
  if (status.value === CONTEST_STATUS.NOT_STARTED) return "距开始";
  if (status.value === CONTEST_STATUS.RUNNING) return "剩余";
  return "";
});

const contestStatusDotClass = computed(() => {
  if (status.value === CONTEST_STATUS.NOT_STARTED) return "bg-amber-400";
  if (status.value === CONTEST_STATUS.RUNNING) return "bg-emerald-500";
  return "bg-gray-300";
});

const curLabel = computed(() => route.params.problemId);
const curInList = computed(() =>
  problemList.value.find((p) => String(p.label) === String(curLabel.value))
);

// 题目切换器：当前蓝底、已过绿、挂红、其余灰
const switchClass = (p) => {
  if (String(p.label) === String(curLabel.value)) return "border-blue-500 bg-blue-500 text-white";
  if (p.status === AcceptedCode) return "border-emerald-200 bg-emerald-50 text-emerald-600 hover:border-emerald-400";
  if (p.status === WrongAnswerCode) return "border-rose-200 bg-rose-50 text-rose-500 hover:border-rose-400";
  return "border-gray-200 text-gray-500 hover:border-blue-400 hover:text-blue-500";
};

const customTheme = EditorView.theme({
  "&": { fontSize: "14px", border: "1px solid #d1d5db", borderRadius: "2px", overflow: "hidden" },
  "&.cm-focused": { borderColor: "#d1d5db", outline: "none" },
  ".cm-content": { fontFamily: '"JetBrains Mono", "Fira Code", Consolas, monospace', fontVariantLigatures: "contextual" },
  ".cm-gutters": { fontFamily: '"JetBrains Mono", "Fira Code", Consolas, monospace' },
});
const selectedLanguageName = computed(() =>
  languageList.value.find((item) => item.id === language.value)?.name || ""
);
const extensions = computed(() => {
  const languageExtension = getEditorLanguageExtension(selectedLanguageName.value);
  return languageExtension ? [languageExtension, customTheme] : [customTheme];
});

const goList = () => router.push(`/contest/${route.params.id}/problem`);
const goRank = () => router.push(`/contest/${route.params.id}/rank`);
const goSubmission = () => router.push(`/contest/${route.params.id}/submission`);
const goContestHome = () => router.push({ name: "ContestDescription", params: { id: route.params.id } });
const goSubmissionDetail = (id) => router.push({ name: "SubmissionDetail", params: { id } });
const goProblem = (label) => {
  if (String(label) !== String(curLabel.value)) {
    router.push(`/contest/${route.params.id}/problem/${label}`);
  }
};

const fetchProblem = async () => {
  const res = await getContestProblem({ contestID: route.params.id, problemId: route.params.problemId });
  problem.value = res.data || {};
};

const fetchProblemList = async () => {
  const res = await getContestProblemList(route.params.id);
  problemList.value = res.data.problemList || [];
};

const fetchContest = async () => {
  const res = await getContestDesc(route.params.id);
  contestData.value = res.data || {};
};

const handleSubmit = async () => {
  if (!userStore.isLogin) {
    userStore.promptLogin();
    return;
  }
  if (!language.value) {
    ElMessage.warning("请选择语言");
    return;
  }
  if (!code.value.trim()) {
    ElMessage.warning("请输入代码");
    return;
  }
  submitting.value = true;
  try {
    await submit({
      contestId: Number(route.params.id),
      label: route.params.problemId,
      language: language.value,
      code: code.value,
    });
    ElMessage.success("提交成功");
    router.push(`/contest/${route.params.id}/submission`);
  } catch (err) {
    console.error(err);
  } finally {
    submitting.value = false;
  }
};

onMounted(async () => {
  await loadLanguages();
  if (languageList.value?.length && !language.value) {
    language.value = languageList.value[0].id;
  }
  fetchProblem();
  fetchProblemList();
  fetchContest();
});

// 侧栏切题：同一路由换 label，重新拉题面并清空代码
watch(
  () => route.params.problemId,
  () => {
    code.value = "";
    fetchProblem();
  }
);
</script>
