<script setup lang="ts">
import { shallowRef, watch } from "vue";
import { RefreshRight } from "@element-plus/icons-vue";
import { getCaptcha } from "@/api/auth";

const code = defineModel<string>({ default: "" });
const captchaId = defineModel<string>("captchaId", { default: "" });
const loading = defineModel<boolean>("loading", { default: false });
const props = withDefaults(defineProps<{ refreshKey?: number }>(), {
  refreshKey: 0
});

const image = shallowRef("");

async function refreshCaptcha() {
  if (loading.value) return;
  loading.value = true;
  try {
    const response = await getCaptcha();
    captchaId.value = response.data?.id || "";
    image.value = response.data?.image || "";
    code.value = "";
  } catch {
    captchaId.value = "";
    image.value = "";
  } finally {
    loading.value = false;
  }
}

watch(() => props.refreshKey, refreshCaptcha, { immediate: true });
</script>

<template>
  <div class="login-captcha">
    <el-input
      v-model.trim="code"
      maxlength="5"
      inputmode="numeric"
      autocomplete="off"
      placeholder="请输入图中数字"
    />
    <button
      type="button"
      class="captcha-image"
      :disabled="loading"
      aria-label="点击更换验证码"
      title="点击更换验证码"
      @click="refreshCaptcha"
    >
      <img v-if="image" :src="image" alt="登录图形验证码" />
      <el-icon v-else :class="{ 'is-loading': loading }">
        <RefreshRight />
      </el-icon>
    </button>
  </div>
</template>

<style scoped>
.login-captcha {
  display: flex;
  width: 100%;
  gap: 10px;
}

.login-captcha :deep(.el-input) {
  min-width: 0;
  flex: 1;
}

.captcha-image {
  width: 112px;
  height: 40px;
  padding: 0;
  overflow: hidden;
  flex: 0 0 auto;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-blank);
  cursor: pointer;
}

.captcha-image:hover:not(:disabled) {
  border-color: var(--el-color-primary);
}

.captcha-image:disabled {
  cursor: wait;
  opacity: 0.7;
}

.captcha-image img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: fill;
}

.is-loading {
  animation: captcha-spin 0.8s linear infinite;
}

@keyframes captcha-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .is-loading {
    animation: none;
  }
}
</style>
