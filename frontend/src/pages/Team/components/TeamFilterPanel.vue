<script setup>
import { Plus, Search, UserFilled } from "@element-plus/icons-vue";

defineProps({
  mine: { type: Boolean, default: false },
  visibility: { type: Number, default: null },
  total: { type: Number, default: 0 },
  isLogin: { type: Boolean, default: false },
});

const keyword = defineModel("keyword", { type: String, default: "" });
const emit = defineEmits(["search", "toggleMine", "changeVisibility", "create"]);
</script>

<template>
  <section class="team-filter-panel f-panel">
    <div class="filter-row filter-row--search">
      <strong class="filter-label">搜索团队</strong>
      <el-input
        v-model="keyword"
        class="team-search"
        placeholder="输入团队名称"
        clearable
        :prefix-icon="Search"
        @keyup.enter="emit('search')"
        @clear="emit('search')"
      />
      <el-button type="primary" :icon="Search" @click="emit('search')">搜索</el-button>
      <button
        v-if="isLogin"
        class="mine-button"
        :class="{ 'mine-button--active': mine }"
        type="button"
        :aria-pressed="mine"
        @click="emit('toggleMine')"
      >
        <el-icon><UserFilled /></el-icon>
        {{ mine ? "查看全部团队" : "我的团队" }}
      </button>
      <el-button v-permission="['admin']" type="primary" :icon="Plus" @click="emit('create')">创建团队</el-button>
      <span class="filter-total">共 {{ total }} 个团队</span>
    </div>

    <div class="filter-row filter-row--visibility">
      <strong class="filter-label">团队权限</strong>
      <div class="visibility-options" role="group" aria-label="团队权限筛选">
        <button
          class="visibility-option visibility-option--all"
          :class="{ active: visibility === null }"
          type="button"
          :aria-pressed="visibility === null"
          @click="emit('changeVisibility', null)"
        >
          全部
        </button>
        <button
          class="visibility-option visibility-option--public"
          :class="{ active: visibility === 0 }"
          type="button"
          :aria-pressed="visibility === 0"
          @click="emit('changeVisibility', 0)"
        >
          公开团队
        </button>
        <button
          class="visibility-option visibility-option--private"
          :class="{ active: visibility === 1 }"
          type="button"
          :aria-pressed="visibility === 1"
          @click="emit('changeVisibility', 1)"
        >
          私有团队
        </button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.team-filter-panel { margin-bottom: 16px; padding: 19px 22px; }
.filter-row { min-height: 36px; display: flex; align-items: center; gap: 10px; }
.filter-row + .filter-row { margin-top: 14px; }
.filter-label { flex: 0 0 76px; color: #263243; font-size: 16px; font-weight: 650; }
.team-search { width: 240px; }
.mine-button { min-height: 32px; padding: 0 13px; border: 1px solid #f0cf9e; border-radius: 4px; display: inline-flex; align-items: center; gap: 6px; color: #9b671d; background: #fff8ea; font-size: 14px; cursor: pointer; transition: border-color 150ms ease, color 150ms ease, background-color 150ms ease; }
.mine-button:hover, .mine-button--active { border-color: #e8b965; color: #87540e; background: #fde6bb; }
.mine-button:focus-visible, .visibility-option:focus-visible { outline: 3px solid rgba(23, 105, 224, 0.17); outline-offset: 2px; }
.filter-total { margin-left: auto; color: #7c8797; font-size: 14px; }
.visibility-options { display: flex; align-items: center; gap: 10px; }
.visibility-option { min-height: 30px; padding: 0 12px; border: 1px solid #d8e0ea; border-radius: 4px; color: #637083; background: #fff; font-size: 14px; cursor: pointer; transition: border-color 150ms ease, color 150ms ease, background-color 150ms ease; }
.visibility-option:hover { border-color: #aebdce; color: #344256; }
.visibility-option--all.active { border-color: #1769e0; color: #fff; background: #1769e0; }
.visibility-option--public { border-color: #b9dfc6; color: #2d9154; }
.visibility-option--public:hover, .visibility-option--public.active { border-color: #56b878; color: #207542; background: #edf9f1; }
.visibility-option--private { border-color: #efc7c7; color: #c35757; }
.visibility-option--private:hover, .visibility-option--private.active { border-color: #dc8585; color: #aa3e3e; background: #fff2f2; }

@media (max-width: 880px) {
  .filter-row--search { flex-wrap: wrap; }
  .filter-total { flex-basis: 100%; margin-left: 86px; }
}

@media (max-width: 560px) {
  .team-filter-panel { padding: 16px; }
  .filter-row { align-items: flex-start; flex-wrap: wrap; }
  .filter-label { flex-basis: 100%; }
  .team-search { width: 100%; }
  .filter-total { margin-left: 0; }
  .visibility-options { flex-wrap: wrap; }
}
</style>
