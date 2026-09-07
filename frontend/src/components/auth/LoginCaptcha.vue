<script setup>
import { shallowRef, watch } from "vue";
import { RefreshRight } from "@element-plus/icons-vue";
import { getLoginCaptcha } from "@/api/user";

const code = defineModel({ type: String, default: "" });
const captchaId = defineModel("captchaId", { type: String, default: "" });
const loading = defineModel("loading", { type: Boolean, default: false });
const props = defineProps({
  refreshKey: { type: Number, default: 0 },
});

const image = shallowRef("");

async function refreshCaptcha() {
  if (loading.value) return;
  loading.value = true;
  try {
    const response = await getLoginCaptcha();
    captchaId.value = response.data.id;
    image.value = response.data.image;
    code.value = "";
  } catch (error) {
    captchaId.value = "";
    image.value = "";
    console.error(error);
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
      :prefix-icon="RefreshRight"
      maxlength="5"
      inputmode="numeric"
      autocomplete="off"
      placeholder="请输入图中数字"
    />
    <button
      type="button"
      class="captcha-image"
      :disabled="loading"
      title="点击更换验证码"
      aria-label="点击更换验证码"
      @click="refreshCaptcha"
    >
      <img v-if="image" :src="image" alt="登录图形验证码" />
      <el-icon v-else class="captcha-loading" :class="{ 'is-loading': loading }"><RefreshRight /></el-icon>
    </button>
  </div>
</template>

<style scoped>
.login-captcha {
  display: flex;
  width: 100%;
  gap: 8px;
}

.login-captcha :deep(.el-input) {
  min-width: 0;
  flex: 1;
}

.captcha-image {
  width: 108px;
  height: 36px;
  padding: 0;
  overflow: hidden;
  flex: 0 0 auto;
  border: 1px solid rgba(255, 255, 255, 0.75);
  border-radius: 3px;
  color: #63736f;
  background: rgba(255, 255, 255, 0.48);
  cursor: pointer;
  transition: border-color 0.2s, opacity 0.2s;
}

.captcha-image:hover:not(:disabled) {
  border-color: #2f7fe0;
}

.captcha-image:disabled {
  cursor: wait;
  opacity: 0.72;
}

.captcha-image img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: fill;
}

.captcha-loading {
  font-size: 18px;
}

.captcha-loading.is-loading {
  animation: captcha-spin 0.8s linear infinite;
}

@keyframes captcha-spin {
  to { transform: rotate(360deg); }
}

@media (prefers-reduced-motion: reduce) {
  .captcha-loading.is-loading { animation: none; }
}
</style>
