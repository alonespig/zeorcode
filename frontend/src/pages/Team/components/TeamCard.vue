<script setup>
import { computed } from "vue";
import { Calendar, Lock, User, UserFilled } from "@element-plus/icons-vue";

const props = defineProps({
  team: { type: Object, required: true },
  isLogin: { type: Boolean, default: false },
});

const emit = defineEmits(["enter", "join", "requestJoin"]);

const shortDate = computed(() => props.team.createdAt?.slice(0, 10) || "—");
const visualTone = computed(() => `team-entry__visual--tone-${(Number(props.team.id) || 0) % 3}`);
</script>

<template>
  <article class="team-entry" :class="{ 'team-entry--private': team.visibility === 1 }">
    <div class="team-entry__main">
      <div
        class="team-entry__visual"
        :class="[visualTone, { 'team-entry__visual--has-cover': team.coverUrl }]"
        aria-label="团队封面"
      >
        <img v-if="team.coverUrl" :src="team.coverUrl" :alt="`${team.name}封面`" />
        <template v-else>
          <span class="team-entry__visual-shape" aria-hidden="true"></span>
          <el-icon :size="44"><UserFilled /></el-icon>
        </template>
      </div>

      <div class="team-entry__content">
        <div class="team-entry__title-row">
          <button class="team-entry__name" type="button" @click="emit('enter', team)">
            {{ team.name }}
          </button>
          <span
            class="team-entry__access"
            :class="{ 'team-entry__access--private': team.visibility === 1 }"
            :aria-label="team.visibility === 1 ? '需邀请码' : '公开团队'"
            :title="team.visibility === 1 ? '需邀请码' : undefined"
          >
            <el-icon v-if="team.visibility === 1"><Lock /></el-icon>
            <template v-else>公开</template>
          </span>
        </div>

        <div class="team-entry__meta">
          <span class="team-entry__members">
            <el-icon><User /></el-icon>
            × {{ team.memberCount || 0 }}
          </span>
          <i class="team-entry__divider" aria-hidden="true"></i>
          <div class="team-entry__date-block">
            <span>
              <el-icon><Calendar /></el-icon>
              {{ shortDate }}
            </span>
          </div>
        </div>

        <div class="team-entry__owner-actions">
          <div class="team-entry__owner">
            <el-avatar class="team-entry__owner-avatar" :size="24" :src="team.ownerAvatar || undefined">
              <el-icon :size="14"><UserFilled /></el-icon>
            </el-avatar>
            <strong>{{ team.owner || "—" }}</strong>
          </div>
          <button
            v-if="isLogin && team.myRole != null"
            class="team-entry__action team-entry__action--quiet"
            type="button"
            @click="emit('enter', team)"
          >
            进入团队
          </button>
        </div>
      </div>
    </div>

    <footer v-if="isLogin && team.myRole == null" class="team-entry__footer">
      <span class="team-entry__membership">未加入</span>
      <button
        v-if="team.visibility === 1"
        class="team-entry__action team-entry__action--outline"
        type="button"
        @click="emit('requestJoin', team)"
      >
        输入邀请码
      </button>
      <button v-else class="team-entry__action" type="button" @click="emit('join', team)">
        加入团队
      </button>
    </footer>
  </article>
</template>

<style scoped>
.team-entry {
  min-width: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  border: 1px solid #dfe4eb;
  border-radius: 4px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(24, 34, 48, 0.05);
  transition: border-color 160ms ease, box-shadow 160ms ease;
}

.team-entry:hover { border-color: #c6d2e2; box-shadow: 0 5px 16px rgba(24, 34, 48, 0.07); }
.team-entry__main { min-height: 148px; flex: 1; display: grid; grid-template-columns: 136px minmax(0, 1fr); }
.team-entry__visual { position: relative; width: 124px; height: 124px; margin: 12px 0 12px 12px; overflow: hidden; display: grid; place-items: center; color: #fff; background: #1769e0; }
.team-entry__visual::before { content: ""; position: absolute; width: 112px; height: 112px; border: 19px solid rgba(255, 255, 255, 0.12); transform: rotate(28deg); }
.team-entry__visual img { position: relative; z-index: 2; width: 100%; height: 100%; display: block; object-fit: cover; }
.team-entry__visual-shape { position: absolute; right: -18px; bottom: -28px; width: 82px; height: 82px; border-radius: 50%; background: rgba(255, 255, 255, 0.1); }
.team-entry__visual :deep(.el-icon) { position: relative; z-index: 1; }
.team-entry__visual--tone-1 { background: #1987a7; }
.team-entry__visual--tone-2 { background: #5271c4; }
.team-entry--private .team-entry__visual { background: #c88224; }
.team-entry__content { min-width: 0; padding: 20px 16px 9px; display: flex; flex-direction: column; }
.team-entry__title-row { min-width: 0; display: flex; align-items: flex-start; gap: 8px; }
.team-entry__name { min-width: 0; overflow: hidden; padding: 0; border: 0; color: #1b2636; background: transparent; font-size: 18px; font-weight: 650; line-height: 1.45; text-align: left; text-overflow: ellipsis; white-space: nowrap; cursor: pointer; }
.team-entry__name:hover { color: #1769e0; }
.team-entry__name:focus-visible, .team-entry__action:focus-visible { outline: 3px solid rgba(23, 105, 224, 0.18); outline-offset: 2px; }
.team-entry__access { flex: 0 0 auto; min-height: 24px; margin-left: auto; display: inline-flex; align-items: center; gap: 4px; color: #249150; font-size: 13px; white-space: nowrap; }
.team-entry__access--private { color: #bd7610; }
.team-entry__owner-actions { margin-top: 12px; display: flex; align-items: center; gap: 10px; }
.team-entry__owner { min-width: 0; display: flex; align-items: center; gap: 7px; color: #667386; font-size: 14px; }
.team-entry__owner-avatar { flex: 0 0 auto; border: 1px solid #dce6f2; color: #1769e0; background: #edf4fd; }
.team-entry__owner strong { min-width: 0; overflow: hidden; color: #4b5769; font-weight: 550; text-overflow: ellipsis; white-space: nowrap; }
.team-entry__meta { margin-top: 12px; display: flex; align-items: flex-start; gap: 8px; color: #5d6a7c; font-size: 14px; }
.team-entry__meta span { display: inline-flex; align-items: center; gap: 4px; white-space: nowrap; }
.team-entry__members { padding-top: 2px; }
.team-entry__divider { width: 1px; height: 16px; margin-top: 2px; background: #dfe4eb; }
.team-entry__date-block { min-width: 0; margin-left: auto; display: flex; align-items: flex-end; flex-direction: column; gap: 7px; }
.team-entry__footer { min-height: 46px; padding: 6px 10px 6px 14px; border-top: 1px solid #e8edf3; display: flex; align-items: center; gap: 10px; background: #fbfcfe; }
.team-entry__membership { color: #2d3747; font-size: 14px; font-weight: 520; }
.team-entry__action { min-height: 32px; margin-left: auto; padding: 0 12px; border: 1px solid #1769e0; border-radius: 4px; color: #fff; background: #1769e0; font-size: 14px; font-weight: 520; cursor: pointer; transition: border-color 150ms ease, color 150ms ease, background-color 150ms ease; }
.team-entry__action:hover { border-color: #0f58c4; background: #0f58c4; }
.team-entry__action--outline { border-color: #9dbce7; color: #1769e0; background: #fff; }
.team-entry__action--outline:hover { border-color: #75a1dd; color: #0f58c4; background: #f4f8fe; }
.team-entry__action--quiet { min-height: 30px; padding-inline: 10px; border-color: #c9d9ef; color: #1769e0; background: #fff; }
.team-entry__action--quiet:hover { border-color: #9fbce4; color: #0f58c4; background: #f2f7fd; }

@media (max-width: 560px) {
  .team-entry__main { grid-template-columns: 116px minmax(0, 1fr); }
  .team-entry__visual { width: 104px; height: 104px; }
  .team-entry__content { padding-inline: 13px; }
  .team-entry__visual :deep(.el-icon) { font-size: 38px !important; }
  .team-entry__access { font-size: 12px; }
}

@media (prefers-reduced-motion: reduce) {
  .team-entry, .team-entry__action { transition: none; }
}
</style>
