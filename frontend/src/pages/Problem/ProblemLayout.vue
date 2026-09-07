<template>
  <div class="problem-page mx-auto max-w-[1300px] px-10 py-4 max-md:px-3">
    <div class="flex items-start gap-5 max-lg:flex-col max-lg:items-stretch max-lg:gap-4">
      <!-- 左：主内容区，随子路由切换 -->
      <div class="min-w-0 flex-1 max-lg:order-2 max-lg:w-full">
        <router-view />
      </div>

      <!-- 右：常驻侧栏。窄屏下 aside 自身 display:contents「溶解」，
           内部各块直接成为外层单列 flex 的子项，用 order 重排：
           导航上移到题面之上，信息类面板下沉到题面之下。 -->
      <aside class="flex w-66 shrink-0 flex-col gap-6 max-lg:contents">

        <!-- 窄屏：纵向导航轨 → 可横滚的横向 tab 条，高亮边框由左侧改到底部 -->
        <ul class="f-panel flex flex-col  text-gray-600  text-base py-0.5
        [&>li:hover]:border-blue-500
        [&>li]:px-4 [&>li]:py-2 [&>li]:border-l-3 [&>li]:border-transparent
        [&>li]:cursor-pointer [&_.iconfont]:mr-2
        max-lg:order-1 max-lg:flex-row max-lg:overflow-x-auto max-lg:py-0
        max-lg:[&>li]:shrink-0 max-lg:[&>li]:whitespace-nowrap
        max-lg:[&>li]:border-l-0 max-lg:[&>li]:border-b-3
        ">
          <li :class="activeClass('ProblemDetail')" @click="go('ProblemDetail')">
            <span class="iconfont icon-xinxiliebiao"></span>
            题目
          </li>
          <li :class="activeClass('ProblemSubmissions')" @click="go('ProblemSubmissions')">
            <span class="iconfont icon-shalou1"></span>
            提交记录
          </li>
          <li :class="activeClass('ProblemDiscuss')" @click="go('ProblemDiscuss')">
            <span class="iconfont icon-jurassic_bbs"></span>
            讨论
          </li>
          <li :class="activeClass('ProblemSolutions')" @click="go('ProblemSolutions')">
            <span class="iconfont icon-nantijiejue1"></span>
            题解
          </li>
          <li v-permission="['admin']" @click="goFile(problem?.id)">
            <span class="iconfont icon-24gf-folderOpen"></span>
            文件
          </li>
          <li v-permission="['admin']" @click="goEdit(problem?.id)">
            <span class="iconfont icon-edit"></span>
            编辑
          </li>
        </ul>

        <!-- 信息类面板。桌面端 wrapper 自身 display:contents，两个面板仍是 aside 的直接
             flex 子项，gap-6 原样生效；窄屏才变成并排两列（<768 再降回单列）。 -->
        <div class="contents max-lg:order-3 max-lg:grid max-lg:grid-cols-2 max-lg:gap-4 max-md:grid-cols-1">
          <!-- 题目信息 -->
          <div class="rounded border-2 border-gray-200 bg-white">
            <div class="border-b border-gray-100 px-4 py-3 text-sm text-gray-700">题目信息</div>
            <div class="divide-y divide-gray-100 text-[13.5px]">
              <div class="flex justify-between px-4 py-2">
                <span class="text-gray-500 font-medium">编号</span>
                <span class="font-mono text-gray-700">{{ problem?.id }}</span>
              </div>
              <div class="flex justify-between px-4 py-2">
                <span class="text-gray-500">时间限制</span>
                <span class="font-mono text-gray-700">{{ problem?.timeLimitMs }} MS</span>
              </div>
              <div class="flex justify-between px-4 py-2">
                <span class="text-gray-500">内存限制</span>
                <span class="font-mono text-gray-700">{{ problem?.memoryLimitMb }} MB</span>
              </div>
              <div class="flex justify-between px-4 py-2">
                <span class="text-gray-500">难度</span>
                <span :class="diffClass(problem?.difficulty)">{{ diffText(problem?.difficulty) }}</span>
              </div>
              <div class="flex justify-between px-4 py-2">
                <span class="text-gray-500">来源</span>
                <span v-if="problem?.oj" class="text-sky-600">远程 · {{ problem.oj }}</span>
                <span v-else class="text-gray-700">本地</span>
              </div>
              <div class="flex justify-between px-4 py-2">
                <span class="text-gray-500">通过 / 提交</span>
                <span class="font-mono text-gray-700">{{ problem?.acceptedCount ?? 0 }} / {{ problem?.submitCount ?? 0
                }}</span>
              </div>
              <div v-if="problem?.tags?.length" class="flex justify-between px-4 py-2">
                <span class="text-gray-400">知识点</span>
                <span class="cursor-pointer text-blue-500" @click="showTags = !showTags">
                  {{ showTags ? "隐藏" : "显示" }}
                </span>
              </div>
              <div v-if="problem?.tags?.length && showTags" class="flex flex-wrap gap-1.5 px-4 py-2.5">
                <el-tag v-for="tag in problem.tags" :key="tag">
                  {{ tag.name }}
                </el-tag>
              </div>
            </div>
          </div>

          <!-- 提交统计 -->
          <div class="f-panel">
            <Statistics :data="statistics" title="提交统计" />
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, provide, onMounted, onBeforeUnmount, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getProblem } from "@/api/problems";
import Statistics from "@/components/problem/Statistics.vue";

const route = useRoute();
const router = useRouter();

const problem = ref({});
const showTags = ref(false);

// 难度：数字 → 文字 + 颜色
const DIFF_TEXT = { 1: "简单", 2: "中等", 3: "困难" };
const DIFF_CLASS = { 1: "text-emerald-600", 2: "text-amber-600", 3: "text-red-500" };
const diffText = (d) => DIFF_TEXT[d] || "—";
const diffClass = (d) => DIFF_CLASS[d] || "text-gray-700";

// 供各 tab 子页(题面/题解/讨论/提交)注入使用，避免重复请求
provide("problem", problem);

const statistics = computed(() => [
  { name: "通过", value: problem.value.acceptedCount ?? 0 },
  { name: "未通过", value: (problem.value.submitCount ?? 0) - (problem.value.acceptedCount ?? 0) },
]);


const activeClass = (module) => {
  return route.meta.module === module ? 'bg-blue-400 text-white' : '';
}

const go = (name) => {
  router.push({ name, params: { id: route.params.id } });
};

const goEdit = (id) => {
  router.push({ name: "EditProblem", params: { id } });
};

const goFile = (id) => {
  router.push({ name: "ProblemFile", params: { id } });
};

const fetchProblem = async () => {
  const res = await getProblem(route.params.id);
  problem.value = res.data;
};

let oldTitle = "";

onMounted(async () => {
  await fetchProblem();
  if (problem.value.name) {
    oldTitle = document.title;
    document.title = `${problem.value.id}. ${problem.value.name}`;
  }
});

// 从一道题切到另一道题时重新拉取
watch(
  () => route.params.id,
  () => fetchProblem()
);

onBeforeUnmount(() => {
  if (oldTitle) document.title = oldTitle;
});
</script>

<style scoped lang="scss">
/* 宽度与内边距已移到模板的 Tailwind 类上（原先这里的 padding: 0 40px 特异性更高，
   把模板上的 py-4 覆盖掉了，且 40px 在窄屏放不下）。此处只保留字体。 */
.problem-page {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC",
    "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
}
</style>
