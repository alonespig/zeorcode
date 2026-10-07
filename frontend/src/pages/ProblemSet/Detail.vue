<script setup>
import { computed, shallowRef, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Lock } from "@element-plus/icons-vue";
import { getProblemSetDetail } from "@/api/problemset";
import { tagStyle } from "@/utils/tag";

const route = useRoute();
const router = useRouter();
const detail = shallowRef(null);
const loading = shallowRef(false);
let detailSeq = 0;

const tabs = computed(() => {
  const base = `/problemset/${route.params.id}`;
  const items = [
    { label: "简介", to: base, name: "ProblemSetIntro" },
    { label: "题目列表", to: `${base}/problems`, name: "ProblemSetProblems" },
    { label: "排行榜", to: `${base}/rank`, name: "ProblemSetRank" },
  ];
  return detail.value?.locked ? items.slice(0, 1) : items;
});

const loadDetail = async () => {
  const seq = ++detailSeq;
  loading.value = true;
  try {
    const res = await getProblemSetDetail(route.params.id);
    if (seq !== detailSeq) return;
    detail.value = res.data;
  } catch (err) {
    if (seq !== detailSeq) return;
    console.error(err);
    detail.value = null;
  } finally {
    if (seq === detailSeq) loading.value = false;
  }
};

watch(
  () => route.params.id,
  loadDetail,
  { immediate: true },
);

watch(
  [detail, () => route.name],
  () => {
    if (!detail.value?.locked || route.name === "ProblemSetIntro") return;
    router.replace({ name: "ProblemSetIntro", params: { id: route.params.id } });
  },
);
</script>

<template>
  <div class="problemset-detail page-container" v-loading="loading">
    <template v-if="detail">
      <div class="f-panel">
        <div class="px-5 py-4">
          <h1 class="flex items-center gap-2 text-xl font-semibold text-gray-800">
            {{ detail.title }}
            <el-tooltip v-if="detail.visibility === 1" content="需要邀请码" placement="top">
              <el-icon class="text-gray-400" :size="16">
                <Lock />
              </el-icon>
            </el-tooltip>
          </h1>
          <div class="mt-2.5 flex flex-wrap items-center gap-2 text-[12.5px] text-gray-400">
            <el-tag v-for="tag in detail.tags" :key="tag.id" effect="light" :style="tagStyle(tag)">
              {{ tag.name }}
            </el-tag>
            <span v-if="detail.tags?.length" class="text-gray-300">·</span>
            <span>共 <b class="text-gray-700">{{ detail.problemCount }}</b> 题</span>
            <span class="text-gray-300">·</span>
            <span>{{ detail.author || "-" }}</span>
            <span class="text-gray-300">·</span>
            <span>更新于 {{ detail.updatedAt }}</span>
          </div>
        </div>

        <nav class="flex items-center gap-6 border-t border-gray-100 px-5" aria-label="题单导航">
          <router-link
            v-for="tab in tabs"
            :key="tab.name"
            :to="tab.to"
            class="border-b-2 border-transparent py-3 text-[15px] text-gray-600 hover:text-blue-500"
            exact-active-class="border-blue-500! text-blue-500! font-medium"
          >
            {{ tab.label }}
          </router-link>
        </nav>
      </div>

      <div class="mt-3.5">
        <router-view v-slot="{ Component }">
          <component
            :is="Component"
            v-if="!detail.locked || route.name === 'ProblemSetIntro'"
            :detail="detail"
            @unlocked="loadDetail"
          />
        </router-view>
      </div>
    </template>
  </div>
</template>

<style scoped>
.problemset-detail {
  min-height: 360px;
  padding-top: 8px;
  padding-bottom: 24px;
}
</style>
