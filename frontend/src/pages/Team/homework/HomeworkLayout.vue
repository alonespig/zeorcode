<template>
  <div class="page-container py-4">
    <!-- 作业头部 + tab 合并成一张卡片，同团队页 -->
    <div class="f-panel">
      <div class="px-6 py-4">
        <div class="flex flex-wrap items-center gap-3">
          <router-link class="text-sm text-gray-400 hover:text-blue-500" :to="`/team/${hw.teamId}`">
            {{ hw.teamName }}
          </router-link>
          <span class="text-gray-300">/</span>
          <h1 class="text-xl font-semibold text-gray-800">{{ hw.title }}</h1>
          <el-tag size="small" :type="statusMeta(hw.status).type" effect="light">{{ statusMeta(hw.status).text }}</el-tag>
        </div>
      </div>

      <div class="flex items-center gap-6 border-t border-gray-100 px-6">
        <router-link v-for="tab in tabs" :key="tab.name" :to="tab.to"
          class="flex items-center gap-1.5 border-b-2 border-transparent py-3 text-[15px] font-medium text-gray-700 hover:text-blue-500"
          exact-active-class="border-blue-500! text-blue-500! font-medium">
          {{ tab.label }}
        </router-link>
      </div>
    </div>

    <!-- 内容区不包 f-panel，各 tab 自己渲染独立卡片，间距 mt-4 -->
    <div class="mt-4">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, provide, watch } from "vue";
import { useRoute } from "vue-router";
import { getHomeworkDetail } from "@/api/team";

const route = useRoute();
const hw = ref({});

// 供四个 tab 注入使用，避免重复请求
provide("homework", hw);

const tabs = computed(() => {
  const base = `/team/${route.params.id}/homework/${route.params.hid}`;
  return [
    { name: "HomeworkIntro", label: "简介", to: base },
    { name: "HomeworkProblems", label: "题目列表", to: `${base}/problems` },
    { name: "HomeworkRank", label: "排行榜", to: `${base}/rank` },
    { name: "HomeworkSubmissions", label: "提交", to: `${base}/submissions` },
  ];
});

const statusMeta = (s) => ({
  0: { text: "未开始", type: "warning" },
  1: { text: "进行中", type: "success" },
  2: { text: "已截止", type: "info" },
}[s] || { text: "-", type: "info" });

const fetchHomework = async () => {
  try {
    const res = await getHomeworkDetail(route.params.hid);
    hw.value = res.data;
  } catch (err) {
    console.error(err);
  }
};

watch(() => route.params.hid, fetchHomework, { immediate: true });
</script>
