<template>
  <el-button :disabled="cooldown > 0 || sending" :loading="sending" @click="handleSend">
    {{ cooldown > 0 ? `${cooldown}s 后重发` : "发送验证码" }}
  </el-button>
</template>

<script setup>
import { ref, onBeforeUnmount } from "vue";
import { ElMessage } from "element-plus";
import { sendVerifyCode } from "@/api/user";

const props = defineProps({
  email: { type: String, default: "" },
  scene: { type: String, required: true }, // register | reset | bind
});

const cooldown = ref(0);
const sending = ref(false);
let timer = null;

const emailValid = (e) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e);

const handleSend = async () => {
  if (cooldown.value > 0 || sending.value) return;
  if (!emailValid(props.email)) {
    ElMessage.warning("请先填写正确的邮箱");
    return;
  }
  sending.value = true;
  try {
    await sendVerifyCode({ email: props.email, scene: props.scene });
    ElMessage.success("验证码已发送，请查收邮箱");
    startCooldown();
  } catch {
    // 错误消息由 request 拦截器统一弹出
  } finally {
    sending.value = false;
  }
};

const startCooldown = () => {
  cooldown.value = 60;
  timer = setInterval(() => {
    cooldown.value--;
    if (cooldown.value <= 0) {
      clearInterval(timer);
      timer = null;
    }
  }, 1000);
};

onBeforeUnmount(() => timer && clearInterval(timer));
</script>
