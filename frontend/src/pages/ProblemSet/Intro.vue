<script setup>
import { shallowRef } from "vue";
import { ElMessage } from "element-plus";
import { Lock } from "@element-plus/icons-vue";
import { unlockProblemSet } from "@/api/problemset";

const props = defineProps({
  detail: { type: Object, required: true },
});
const emit = defineEmits(["unlocked"]);

const inviteCode = shallowRef("");
const unlocking = shallowRef(false);

const submitUnlock = async () => {
  const code = inviteCode.value.trim();
  if (!code) {
    ElMessage.warning("请输入邀请码");
    return;
  }
  unlocking.value = true;
  try {
    await unlockProblemSet(props.detail.id, code);
    inviteCode.value = "";
    ElMessage.success("解锁成功");
    emit("unlocked");
  } catch (err) {
    console.error(err);
  } finally {
    unlocking.value = false;
  }
};
</script>

<template>
  <section class="f-panel">
    <div class="border-b border-gray-100 px-5 py-3.5">
      <h2 class="text-[15px] font-semibold text-gray-700">简介</h2>
    </div>

    <div v-if="detail.description" class="desc-panel px-5 py-4">
      <MdEditor :model-value="detail.description" preview-only editor-id="problemset-desc" />
    </div>
    <el-empty v-else description="暂无简介" :image-size="80" />

    <div v-if="detail.locked" class="border-t border-gray-100 px-5 py-10 text-center">
      <el-icon class="text-gray-300" :size="32">
        <Lock />
      </el-icon>
      <h3 class="mt-2.5 mb-1.5 text-base font-medium text-gray-700">该题单需要邀请码</h3>
      <p class="mb-4 text-[13px] text-gray-400">
        输入邀请码后即可查看全部 {{ detail.problemCount }} 道题目和排行榜
      </p>
      <div class="flex justify-center gap-2.5">
        <el-input
          v-model="inviteCode"
          placeholder="请输入邀请码"
          class="w-52!"
          @keyup.enter="submitUnlock"
        />
        <el-button type="primary" :loading="unlocking" @click="submitUnlock">解锁</el-button>
      </div>
    </div>
  </section>
</template>

<style scoped>
.desc-panel :deep(.md-editor-preview > :first-child) {
  margin-top: 0;
}
</style>
