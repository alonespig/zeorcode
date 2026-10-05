<script setup lang="ts">
import Motion from "./utils/motion";
import { reactive, shallowRef, toRaw, useTemplateRef } from "vue";
import { useRoute, useRouter } from "vue-router";
import type { FormInstance, FormRules } from "element-plus";
import { Lock, User } from "@element-plus/icons-vue";
import { message } from "@/utils/message";
import { useLayout } from "@/layout/hooks/useLayout";
import { useNav } from "@/layout/hooks/useNav";
import { useUserStoreHook } from "@/store/modules/user";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { useDataThemeChange } from "@/layout/hooks/useDataThemeChange";
import LoginCaptcha from "./components/LoginCaptcha.vue";
import { avatar, bg, illustration } from "./utils/static";
import dayIcon from "@/assets/svg/day.svg?component";
import darkIcon from "@/assets/svg/dark.svg?component";

defineOptions({ name: "Login" });

interface LoginForm {
  username: string;
  password: string;
  captchaId: string;
  captchaCode: string;
}

const router = useRouter();
const route = useRoute();
const formRef = useTemplateRef<FormInstance>("formRef");
const loading = shallowRef(false);
const captchaLoading = shallowRef(false);
const captchaRefreshKey = shallowRef(0);

const form = reactive<LoginForm>({
  username: "",
  password: "",
  captchaId: "",
  captchaCode: ""
});

const rules: FormRules<LoginForm> = {
  username: [{ required: true, message: "请输入账号", trigger: "blur" }],
  password: [{ required: true, message: "请输入密码", trigger: "blur" }],
  captchaCode: [
    { required: true, message: "请输入图形验证码", trigger: "blur" },
    {
      pattern: /^\d{5}$/,
      message: "请输入 5 位数字验证码",
      trigger: "blur"
    }
  ]
};

const { initStorage } = useLayout();
initStorage();
const { dataTheme, overallStyle, dataThemeChange } = useDataThemeChange();
dataThemeChange(overallStyle.value);
const { title } = useNav();

async function submit() {
  if (!formRef.value || loading.value || captchaLoading.value) return;
  try {
    await formRef.value.validate();
  } catch {
    return;
  }

  loading.value = true;
  try {
    await useUserStoreHook().login({ ...form });
    const redirect =
      typeof route.query.redirect === "string"
        ? route.query.redirect
        : "/dashboard";
    await router.replace(redirect);
    message("登录成功", { type: "success" });
  } catch (error) {
    if (error instanceof Error) {
      message(error.message, { type: "error" });
    }
    captchaRefreshKey.value += 1;
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="select-none">
    <img :src="bg" class="wave" alt="" />
    <div class="absolute right-5 top-3 flex-c">
      <el-switch
        v-model="dataTheme"
        inline-prompt
        :active-icon="dayIcon"
        :inactive-icon="darkIcon"
        aria-label="切换明暗主题"
        @change="dataThemeChange"
      />
    </div>

    <div class="login-container">
      <div class="img" aria-hidden="true">
        <component :is="toRaw(illustration)" />
      </div>
      <div class="login-box">
        <div class="login-form">
          <avatar class="avatar" aria-hidden="true" />
          <Motion>
            <h1 class="outline-hidden">{{ title }}</h1>
            <p class="login-subtitle">仅限系统管理员登录</p>
          </Motion>

          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            size="large"
            @keyup.enter="submit"
          >
            <Motion :delay="100">
              <el-form-item prop="username">
                <el-input
                  v-model.trim="form.username"
                  clearable
                  autocomplete="username"
                  placeholder="用户名或邮箱"
                  :prefix-icon="useRenderIcon(User)"
                />
              </el-form-item>
            </Motion>

            <Motion :delay="150">
              <el-form-item prop="password">
                <el-input
                  v-model="form.password"
                  type="password"
                  show-password
                  autocomplete="current-password"
                  placeholder="密码"
                  :prefix-icon="useRenderIcon(Lock)"
                />
              </el-form-item>
            </Motion>

            <Motion :delay="200">
              <el-form-item prop="captchaCode">
                <LoginCaptcha
                  v-model="form.captchaCode"
                  v-model:captcha-id="form.captchaId"
                  v-model:loading="captchaLoading"
                  :refresh-key="captchaRefreshKey"
                />
              </el-form-item>
            </Motion>

            <Motion :delay="250">
              <el-button
                class="mt-4! w-full"
                type="primary"
                :loading="loading"
                :disabled="captchaLoading"
                @click="submit"
              >
                登录后台
              </el-button>
            </Motion>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@import url("@/style/login.css");

.login-subtitle {
  margin: -10px 0 24px;
  color: var(--el-text-color-secondary);
  font-size: 13px;
  text-align: center;
}
</style>
