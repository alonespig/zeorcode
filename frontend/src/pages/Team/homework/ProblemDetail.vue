<script setup>
import { computed, onMounted, shallowRef, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { Codemirror } from "vue-codemirror";
import { EditorView } from "@codemirror/view";
import { getHomeworkDetail, getHomeworkProblem, submitHomework } from "@/api/team";
import { useLanguages } from "@/hooks/useLanguages";
import { getEditorLanguageExtension } from "@/utils/editorLanguage";
import ContestProblem from "@/components/problem/ContestProblem.vue";
import HomeworkRecentSubmissions from "@/components/homework/HomeworkRecentSubmissions.vue";

const route = useRoute();
const router = useRouter();
const homework = shallowRef({});
const problem = shallowRef({});
const code = shallowRef("");
const language = shallowRef();
const submitting = shallowRef(false);
const problemLoading = shallowRef(false);
let homeworkFetchVersion = 0;
let problemFetchVersion = 0;

const { languageList, loadLanguages } = useLanguages();

const problemList = computed(() => homework.value.problems || []);
const currentProblemId = computed(() => String(route.params.problemId || ""));
const currentProblem = computed(() =>
  problemList.value.find((item) => String(item.id) === currentProblemId.value)
);
const teamId = computed(() => homework.value.teamId || route.params.id);
const homeworkBase = computed(() => `/team/${teamId.value}/homework/${route.params.hid}`);
const canSubmit = computed(() => homework.value.status !== 0);
const homeworkStatus = computed(() => {
  if (homework.value.status === 0) {
    return { text: `开始于 ${homework.value.startTime || "-"}`, dotClass: "bg-amber-400" };
  }
  if (homework.value.status === 1) {
    return { text: `截止于 ${homework.value.endTime || "-"}`, dotClass: "bg-emerald-500" };
  }
  return { text: "作业已截止（仍可提交，不计入排行榜）", dotClass: "bg-gray-300" };
});

const customTheme = EditorView.theme({
  "&": { fontSize: "14px", border: "1px solid #d1d5db", borderRadius: "2px", overflow: "hidden" },
  "&.cm-focused": { borderColor: "#d1d5db", outline: "none" },
  ".cm-content": {
    fontFamily: '"JetBrains Mono", "Fira Code", Consolas, monospace',
    fontVariantLigatures: "contextual",
  },
  ".cm-gutters": { fontFamily: '"JetBrains Mono", "Fira Code", Consolas, monospace' },
});
const selectedLanguageName = computed(() =>
  languageList.value.find((item) => item.id === language.value)?.name || ""
);
const extensions = computed(() => {
  const languageExtension = getEditorLanguageExtension(selectedLanguageName.value);
  return languageExtension ? [languageExtension, customTheme] : [customTheme];
});

const switchClass = (item) => {
  if (String(item.id) === currentProblemId.value) return "border-blue-500 bg-blue-500 text-white";
  if (item.status === 1) return "border-emerald-200 bg-emerald-50 text-emerald-600 hover:border-emerald-400";
  if (item.status != null) return "border-rose-200 bg-rose-50 text-rose-500 hover:border-rose-400";
  return "border-gray-200 text-gray-500 hover:border-blue-400 hover:text-blue-500";
};

const goProblemList = () => router.push(`${homeworkBase.value}/problems`);
const goHomeworkHome = () => router.push(homeworkBase.value);
const goRank = () => router.push(`${homeworkBase.value}/rank`);
const goSubmissions = () => router.push(`${homeworkBase.value}/submissions`);
const goSubmissionDetail = (id) => router.push({ name: "SubmissionDetail", params: { id } });
const goProblem = (problemId) => {
  if (String(problemId) === currentProblemId.value) return;
  router.push({
    name: "HomeworkProblemDetail",
    params: { id: teamId.value, hid: route.params.hid, problemId },
  });
};

const fetchHomework = async (homeworkId) => {
  const version = ++homeworkFetchVersion;
  try {
    const res = await getHomeworkDetail(homeworkId);
    if (version === homeworkFetchVersion) homework.value = res.data || {};
  } catch (err) {
    console.error(err);
  }
};

const fetchProblem = async (homeworkId, problemId) => {
  const version = ++problemFetchVersion;
  problemLoading.value = true;
  code.value = "";
  try {
    const res = await getHomeworkProblem(homeworkId, problemId);
    if (version === problemFetchVersion) {
      problem.value = { ...(res.data || {}), label: problemId };
    }
  } catch (err) {
    console.error(err);
  } finally {
    if (version === problemFetchVersion) problemLoading.value = false;
  }
};

const handleSubmit = async () => {
  if (!canSubmit.value) {
    ElMessage.warning("作业开始后才能提交");
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
    const res = await submitHomework(route.params.hid, {
      problemID: currentProblemId.value,
      language: language.value,
      code: code.value,
    });
    ElMessage.success("提交成功");
    router.push({ name: "SubmissionDetail", params: { id: res.data.id } });
  } catch (err) {
    console.error(err);
  } finally {
    submitting.value = false;
  }
};

watch(
  () => route.params.hid,
  (homeworkId) => fetchHomework(homeworkId),
  { immediate: true }
);

watch(
  () => [route.params.hid, route.params.problemId],
  ([homeworkId, problemId]) => fetchProblem(homeworkId, problemId),
  { immediate: true }
);

onMounted(async () => {
  await loadLanguages();
  if (languageList.value.length && !language.value) language.value = languageList.value[0].id;
});
</script>

<template>
  <div class="mx-auto max-w-[1340px] px-6 pb-5 max-md:px-3">
    <div class="flex items-start gap-5 max-lg:flex-col max-lg:items-stretch max-lg:gap-4">
      <main class="flex min-w-0 flex-1 flex-col gap-5 max-lg:order-3 max-lg:w-full">
        <ContestProblem v-loading="problemLoading" :problem="problem" />

        <section class="f-panel px-6 py-4" aria-label="提交代码">
          <div class="mb-3 flex items-center">
            <span class="text-sm text-gray-600">语言：</span>
            <el-select v-model="language" class="w-32!">
              <el-option v-for="item in languageList" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
          </div>
          <Codemirror v-model="code" :extensions="extensions" :style="{ height: 'clamp(240px, 40vh, 360px)' }" />
          <div class="mt-4 flex">
            <el-button type="primary" class="ml-auto max-md:ml-0! max-md:w-full!" :disabled="!canSubmit"
              :loading="submitting" @click="handleSubmit">
              <span class="iconfont icon-7 mr-1.5"></span>
              {{ canSubmit ? "提交代码" : "作业开始后可提交" }}
            </el-button>
          </div>
        </section>
      </main>

      <aside class="flex w-72 shrink-0 flex-col gap-4 max-lg:contents">
        <section class="f-panel overflow-hidden max-lg:order-1" aria-label="作业信息">
          <button type="button"
            class="block w-full truncate px-4 py-3 text-center text-[16px] font-semibold text-gray-800 transition-colors hover:text-blue-500 hover:underline"
            title="返回作业题目列表" @click="goProblemList">
            {{ homework.title || "作业" }}
          </button>
          <div
            class="flex items-center justify-center gap-2 border-t border-gray-100 bg-gray-50 px-4 py-3 text-[13px] text-gray-500">
            <span class="h-2 w-2 shrink-0 rounded-full" :class="homeworkStatus.dotClass"></span>
            <span class="truncate" :title="homeworkStatus.text">{{ homeworkStatus.text }}</span>
          </div>
          <div class="grid grid-cols-2 gap-2 border-t border-gray-100 p-3">
            <el-button plain class="ml-0! h-9! w-full" @click="goRank">排行榜</el-button>
            <el-button plain class="ml-0! h-9! w-full" @click="goHomeworkHome">作业首页</el-button>
          </div>
        </section>

        <section class="f-panel max-lg:order-2" aria-label="作业题目列表">
          <div class="border-b border-gray-100 px-4 py-2.5 text-sm text-gray-700">题目列表</div>
          <div class="flex flex-wrap gap-2 p-3.5">
            <button v-for="item in problemList" :key="item.id" type="button"
              class="flex h-9 min-w-10 items-center justify-center rounded border px-2 font-mono text-sm transition-colors"
              :class="switchClass(item)" :title="item.name" @click="goProblem(item.id)">
              {{ item.id }}
            </button>
            <span v-if="!problemList.length" class="py-1 text-sm text-gray-400">暂无题目</span>
          </div>
        </section>

        <HomeworkRecentSubmissions class="max-lg:order-4" :homework-id="route.params.hid"
          :problem-id="currentProblem?.id || currentProblemId" @click-id="goSubmissionDetail"
          @view-all="goSubmissions" />
      </aside>
    </div>
  </div>
</template>
