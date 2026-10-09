<script setup>
import { computed, provide, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Lock, UserFilled, View } from "@element-plus/icons-vue";
import { getTeamDetail } from "@/api/team";
import { useUserStore } from "@/stores/user";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const team = ref({});

// 供概览/作业/成员 tab 注入使用，避免重复请求
provide("team", team);

const tabs = computed(() => {
  const teamID = route.params.id;
  const items = [
    { name: "TeamOverview", label: "概览", to: `/team/${teamID}` },
    { name: "TeamHomework", label: "作业", to: `/team/${teamID}/homework` },
    { name: "TeamMembers", label: "成员", to: `/team/${teamID}/member` },
  ];
  return items;
});

const isTabActive = (tab) => {
  if (tab.name === "TeamHomework") return route.meta.module === "TeamHomework";
  return route.name === tab.name;
};

const shortDate = (date) => (date ? date.slice(0, 10) : "-");

const openAgent = () => {
  router.push({
    name: "Agent",
    query: {
      from: route.fullPath,
      teamId: String(route.params.id),
    },
  });
};

const fetchTeam = async (teamID) => {
  team.value = {};
  try {
    const res = await getTeamDetail(teamID);
    if (String(route.params.id) !== String(teamID)) return;
    team.value = res.data;
  } catch (err) {
    console.error(err);
  }
};

watch(() => route.params.id, fetchTeam, { immediate: true });
</script>

<template>
  <div class="page-container py-4">
    <div class="f-panel">
      <div class="team-heading">
        <div class="team-heading__cover" aria-label="团队封面">
          <img v-if="team.coverUrl" :src="team.coverUrl" :alt="`${team.name || '团队'}封面`" />
          <el-icon v-else :size="34">
            <UserFilled />
          </el-icon>
        </div>

        <div class="team-heading__identity">
          <div class="team-heading__title-row">
            <h1 class="text-2xl font-semibold text-gray-700">{{ team.name }}</h1>
            <span v-if="team.visibility === 1"
              class="inline-flex items-center gap-1 rounded border border-amber-200 bg-amber-50 px-2.5 py-1 text-xs text-amber-600">
              <el-icon>
                <Lock />
              </el-icon>非公开
            </span>
            <span v-else
              class="inline-flex items-center gap-1 rounded border border-green-200 bg-green-50 px-2.5 py-1 text-xs text-green-600">
              <el-icon>
                <View />
              </el-icon>公开
            </span>
          </div>
        </div>

        <div class="team-heading__stats">
          <div class="team-heading__stat">
            <span>成员</span>
            <b>{{ team.memberCount }}</b>
          </div>
          <div class="team-heading__stat">
            <span>创建时间</span>
            <b>{{ shortDate(team.createdAt) }}</b>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-6 border-t border-gray-100 px-6">
        <router-link v-for="tab in tabs" :key="tab.name" :to="tab.to"
          class="flex items-center gap-1.5 border-b-2 border-transparent py-3 text-[15px] font-medium text-gray-700 hover:text-blue-500"
          :class="{ 'border-blue-500! text-blue-500! font-medium': isTabActive(tab) }">
          {{ tab.label }}
        </router-link>
        <button v-if="userStore.isAdmin" type="button" class="team-agent-trigger" @click="openAgent">
          AI 助手
        </button>
      </div>
    </div>

    <div class="mt-4">
      <router-view :key="route.params.id" />
    </div>

  </div>
</template>

<style scoped>
.team-heading {
  min-height: 104px;
  padding: 16px 28px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.team-heading__cover {
  width: 72px;
  height: 72px;
  flex: 0 0 72px;
  overflow: hidden;
  display: grid;
  place-items: center;
  border: 1px solid #d9e3ef;
  border-radius: 4px;
  color: #1769e0;
  background: #edf4fd;
}

.team-heading__cover img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: cover;
}

.team-heading__identity {
  min-width: 0;
}

.team-heading__title-row {
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.team-heading__title-row h1 {
  min-width: 0;
  overflow-wrap: anywhere;
}

.team-heading__stats {
  margin-left: auto;
  display: flex;
  gap: 24px;
  color: #4b5563;
  font-size: 14px;
}

.team-heading__stat {
  min-width: 70px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.team-heading__stat span {
  font-weight: 500;
}

.team-heading__stat b {
  color: #111827;
  font-size: 16px;
  font-variant-numeric: tabular-nums;
}

.team-agent-trigger {
  align-self: stretch;
  display: inline-flex;
  align-items: center;
  padding: 0;
  border-bottom: 2px solid transparent;
  color: #374151;
  font-size: 15px;
  font-weight: 500;
  transition: color 150ms ease, border-color 150ms ease;
}

.team-agent-trigger:hover {
  border-bottom-color: #3b82f6;
  color: #3b82f6;
}

@media (max-width: 640px) {
  .team-heading {
    padding: 14px 16px;
    align-items: flex-start;
    flex-wrap: wrap;
  }

  .team-heading__cover {
    width: 60px;
    height: 60px;
    flex-basis: 60px;
  }

  .team-heading__identity {
    flex: 1;
  }

  .team-heading__stats {
    width: 100%;
    margin-left: 76px;
    justify-content: flex-start;
  }

}
</style>
