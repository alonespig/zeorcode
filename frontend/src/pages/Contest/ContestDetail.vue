<template>
  <div class="mx-auto max-w-310 px-6 pb-5">

    <!-- 头部卡片 -->
    <div class="f-panel">
      <!-- 头部 -->
      <div class="px-7 py-5">
        <div>
          <!-- 第一行：名称居中，报名按钮右列右对齐（不影响居中） -->
          <!-- 第一行：名称居中，报名按钮右对齐 -->
          <h1 class="truncate text-center text-2xl font-semibold text-gray-700">{{ contestData?.name }}</h1>

          <!-- 第二行：左 公开/非公开赛，右 赛制（参考 HOJ 两端对齐） -->
          <div class="mt-3 flex items-center justify-between">
            <span class="inline-flex items-center gap-1 rounded border px-2.5 py-1 text-xs" :class="contestData?.needInviteCode
              ? 'border-amber-200 bg-amber-50 text-amber-600'
              : 'border-green-200 bg-green-50 text-green-600'">
              <el-icon v-if="contestData?.needInviteCode">
                <Lock />
              </el-icon>{{ contestData?.needInviteCode ? '非公开赛' : '公开赛' }}
            </span>
            <el-tag v-if="contestData?.rule" :type="ruleTagType" effect="plain" size="large">
              <span class="inline-flex items-center gap-1">
                <el-icon>
                  <Trophy />
                </el-icon>{{ contestData.rule }}
              </span>
            </el-tag>
          </div>

          <!-- 第二行：进度条（el-slider 只读） -->
          <div class="mt-5 flex items-center gap-4">
            <span class="shrink-0 text-[13px] tabular-nums text-gray-700">{{ startText }}</span>
            <el-slider class="flex-1" v-model="progressValue" :format-tooltip="fmtTooltip" :style="sliderStyle" />
            <span class="shrink-0 text-[13px] tabular-nums text-gray-700">{{ endText }}</span>
          </div>
          <div class="mt-3.5 text-center">
            <el-tag :type="cdTagType" effect="dark" size="large">
              <span class="inline-flex items-center gap-1.5">
                <span class="h-1.5 w-1.5 rounded-full bg-white/90"></span>
                <template v-if="status !== CONTEST_STATUS.ENDED">
                  <span>{{ cdLabel }}</span><span class="tabular-nums">{{ statusTime }}</span>
                </template>
                <template v-else>比赛已结束</template>
              </span>
            </el-tag>
          </div>
        </div>
      </div>
    </div>

    <!-- tab（单独一张卡片） -->
    <div class="f-panel mt-4">
      <div class="flex items-center gap-6 px-6">
        <template v-for="t in tabs" :key="t.to">
          <router-link v-if="!t.locked" :to="t.to"
            class="flex items-center gap-1.5 border-b-2 border-transparent py-3 text-sm text-gray-600 hover:text-blue-500"
            active-class="border-blue-500! text-blue-500! font-medium">
            <span class="iconfont" :class="t.icon"></span>{{ t.name }}
          </router-link>
          <span v-else class="flex cursor-not-allowed items-center gap-1.5 py-3 text-sm text-gray-300"
            title="报名后可查看">
            <span class="iconfont" :class="t.icon"></span>{{ t.name }}
          </span>
        </template>
        <el-button v-if="userStore.isAdmin" size="small" class="ml-auto" @click="goEdit">编辑</el-button>
      </div>
    </div>

    <!-- 未报名提示（报名的家：进入详情随时可报名） -->
    <div v-if="showCta" class="mt-4 flex items-center gap-3 rounded border border-sky-200 bg-sky-50 px-5 py-4">
      <el-icon class="text-blue-500" :size="18">
        <Lock />
      </el-icon>
      <span class="text-sm text-sky-800"><b class="font-semibold">你还未报名本场比赛</b> —— 报名后即可查看题目、提交代码并参与排名。</span>
      <el-button type="success" class="ml-auto" @click="handleJoin">报名参赛</el-button>
    </div>

    <!-- 各 tab 内容 -->
    <div class="mt-4">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import { Lock, Trophy } from "@element-plus/icons-vue";
import dayjs from "dayjs";
import { getContestDetail, joinContest } from "@/api/contest";
import { CONTEST_STATUS } from "@/constants/index";
import { useUserStore } from "@/stores/user";
import { useContestStatus } from "@/hooks/useContestStatus";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

const contestData = ref(null);
const { status, statusTime, progress } = useContestStatus(contestData);

const cdLabel = computed(() =>
  status.value === CONTEST_STATUS.NOT_STARTED ? "距开始" : "剩余"
);

// 倒计时 el-tag 颜色：未开始橙 / 进行中绿 / 已结束灰
const cdTagType = computed(() => ({
  [CONTEST_STATUS.NOT_STARTED]: "warning",
  [CONTEST_STATUS.RUNNING]: "success",
  [CONTEST_STATUS.ENDED]: "info",
}[status.value] || "info"));

// 赛制标签配色（参考 HOJ：ACM 蓝 / OI 绿 / IOI 橙）
const ruleTagType = computed(() => ({
  ACM: "primary",
  OI: "success",
  IOI: "warning",
  CF: "danger",
}[contestData.value?.rule] || "primary"));

const startText = computed(() =>
  contestData.value?.startTime ? dayjs(contestData.value.startTime).format("MM-DD HH:mm") : ""
);
const endText = computed(() =>
  contestData.value?.endTime ? dayjs(contestData.value.endTime).format("MM-DD HH:mm") : ""
);

// 进度条（只读）：值=已进行百分比；空 setter 让滑块不可拖动（仿 HOJ 原版）
const progressValue = computed({
  get: () => progress.value,
  set: () => {},
});
const sliderStyle = computed(() => ({
  "--el-slider-main-bg-color": "var(--judge-accepted)", // 进行中/已结束都保持绿色
  "--el-slider-height": "8px",
}));
const fmtTooltip = (val) => {
  const s = contestData.value?.startTime, e = contestData.value?.endTime;
  if (!s || !e) return "";
  const t = dayjs(s).valueOf() + (val / 100) * (dayjs(e).valueOf() - dayjs(s).valueOf());
  return dayjs(t).format("MM-DD HH:mm");
};

// 报名/已结束/管理员可看全部 tab；否则只放开「简介」
const allowAll = computed(() =>
  userStore.isAdmin ||
  status.value === CONTEST_STATUS.ENDED ||
  (status.value === CONTEST_STATUS.RUNNING && contestData.value?.isRegistered)
);

// 未报名且比赛未结束（非管理员）→ 简介页顶部给报名提示
const showCta = computed(() =>
  !userStore.isAdmin &&
  !contestData.value?.isRegistered &&
  status.value !== CONTEST_STATUS.ENDED
);

const tabs = computed(() => {
  const base = `/contest/${route.params.id}`;
  const items = [
    { name: "简介", to: `${base}/description`, icon: "icon-home1", always: true },
    { name: "题目", to: `${base}/problem`, icon: "icon-tijiaojilu" },
    // 管理员看全场提交，普通用户只看自己
    { name: userStore.isAdmin ? "提交" : "我的提交", to: `${base}/submission`, icon: "icon-shalou" },
    { name: "排名", to: `${base}/rank`, icon: "icon-paihangbang" },
  ];
  // 重判：仅管理员可见
  if (userStore.isAdmin) {
    items.push({ name: "重判", to: `${base}/rejudge`, icon: "icon-fuwuqi", always: true });
  }
  return items.map((it) => ({ ...it, locked: !it.always && !allowAll.value }));
});

const goEdit = () => router.push({ name: "EditContest", params: { id: route.params.id } });

const fetchContest = async () => {
  const res = await getContestDetail(route.params.id);
  contestData.value = res.data || {};
};

// 报名：公开比赛直接报，邀请码比赛先弹 prompt；成功后刷新报名状态
const handleJoin = async () => {
  let inviteCode = "";
  if (contestData.value?.needInviteCode) {
    try {
      const { value } = await ElMessageBox.prompt("请输入邀请码", "邀请码报名", {
        confirmButtonText: "报名",
        cancelButtonText: "取消",
        inputPattern: /\S+/,
        inputErrorMessage: "邀请码不能为空",
      });
      inviteCode = value.trim();
    } catch {
      return; // 用户取消
    }
  }
  try {
    await joinContest({ id: Number(route.params.id), inviteCode });
    ElMessage.success("报名成功");
    await fetchContest();
  } catch (err) {
    console.error(err);
  }
};

onMounted(fetchContest);
</script>
