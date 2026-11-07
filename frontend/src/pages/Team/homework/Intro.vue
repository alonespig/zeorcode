<template>
  <div class="f-panel intro-panel">
    <div v-if="hw.description" class="intro-content">
      <MdEditor :model-value="hw.description" preview-only editor-id="homework-desc" />
    </div>
    <div v-else class="intro-empty">
      <el-empty description="暂无简介" :image-size="88" />
    </div>
  </div>
</template>

<script setup>
import { inject } from "vue";

const hw = inject("homework");
</script>

<style scoped>
.intro-panel {
  display: flex;
  flex-direction: column;
}

.intro-content {
  padding: 20px 24px;
}

.intro-empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 280px;
}
</style>
