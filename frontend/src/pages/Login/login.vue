<template>
  <div class="login-page" :class="{ 'login-page--register': active === 'register' }">
    <div class="login-box">
      <!-- 猫头鹰（输密码时捂眼，沿用原动画） -->
      <div class="owl" :class="{ password: isPasswordFocus }">
        <div class="hand"></div>
        <div class="hand hand-r"></div>
        <div class="arms">
          <div class="arm"></div>
          <div class="arm arm-r"></div>
        </div>
      </div>

      <!-- 毛玻璃卡片 -->
      <div class="glass-card">
        <div class="card-body">
          <div class="brand">
            <h1>Zeor<span>Code</span><small>Online Judge</small></h1>
          </div>

          <div class="tabs">
            <button v-for="t in tabs" :key="t.key" class="tab" :class="{ active: active === t.key }"
              @click="switchTab(t.key)">{{ t.label }}</button>
          </div>

          <!-- 登录 -->
          <el-form v-if="active === 'login'" ref="loginRef" :model="loginForm" :rules="loginRules" label-position="left"
            label-width="82px" @keyup.enter="handleLogin">
            <el-form-item label="账号" prop="username">
              <el-input v-model.trim="loginForm.username" placeholder="用户名或邮箱" :prefix-icon="User" />
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input v-model.trim="loginForm.password" type="password" placeholder="请输入密码"
                :prefix-icon="Lock" show-password @focus="isPasswordFocus = true" @blur="isPasswordFocus = false" />
            </el-form-item>
            <el-form-item label="验证码" prop="captchaCode">
              <LoginCaptcha
                v-model="loginForm.captchaCode"
                v-model:captcha-id="loginForm.captchaId"
                v-model:loading="captchaLoading"
                :refresh-key="captchaRefreshKey"
              />
            </el-form-item>
            <el-button type="primary" class="submit" :loading="loading" :disabled="captchaLoading"
              @click="handleLogin">登录</el-button>
            <p class="alt"><a @click="switchTab('reset')">忘记密码？</a></p>
          </el-form>

          <!-- 注册 -->
          <el-form v-else-if="active === 'register'" ref="regRef" :model="regForm" :rules="regRules"
            label-position="left" label-width="82px" @keyup.enter="handleRegister">
            <el-form-item class="avatar-form-item" label="头像">
              <SquareImageCropper
                v-model="regForm.avatar"
                :upload-request="uploadRegisterAvatar"
                :preview-size="84"
                round
                title="裁剪头像"
                empty-text="选择头像"
                change-text="重新裁剪"
                help-text="可选，PNG/JPG，最大 5MB"
                remove-text="移除头像"
                success-message="头像已上传"
                file-prefix="avatar"
                preview-alt="注册头像预览"
              />
            </el-form-item>
            <el-form-item label="用户名" prop="username">
              <el-input v-model.trim="regForm.username" placeholder="登录账号和公开昵称，2-20 位" :prefix-icon="User" />
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model.trim="regForm.email" placeholder="请输入邮箱" :prefix-icon="Message" />
            </el-form-item>
            <el-form-item label="验证码" prop="code">
              <div class="code-row">
                <el-input v-model.trim="regForm.code" placeholder="邮箱验证码" :prefix-icon="Key" />
                <SendCodeButton :email="regForm.email" scene="register" />
              </div>
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input v-model.trim="regForm.password" type="password" placeholder="至少 6 位"
                :prefix-icon="Lock" show-password @focus="isPasswordFocus = true" @blur="isPasswordFocus = false" />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm">
              <el-input v-model.trim="regForm.confirm" type="password" placeholder="再次输入密码"
                :prefix-icon="Lock" show-password @focus="isPasswordFocus = true" @blur="isPasswordFocus = false" />
            </el-form-item>
            <el-button type="primary" class="submit" :loading="loading" @click="handleRegister">注册</el-button>
          </el-form>

          <!-- 找回密码 -->
          <el-form v-else ref="resetRef" :model="resetForm" :rules="resetRules" label-position="left"
            label-width="82px" @keyup.enter="handleReset">
            <el-form-item label="邮箱" prop="email">
              <el-input v-model.trim="resetForm.email" placeholder="注册时填写的邮箱" :prefix-icon="Message" />
            </el-form-item>
            <el-form-item label="验证码" prop="code">
              <div class="code-row">
                <el-input v-model.trim="resetForm.code" placeholder="邮箱验证码" :prefix-icon="Key" />
                <SendCodeButton :email="resetForm.email" scene="reset" />
              </div>
            </el-form-item>
            <el-form-item label="新密码" prop="password">
              <el-input v-model.trim="resetForm.password" type="password" placeholder="至少 6 位"
                :prefix-icon="Lock" show-password @focus="isPasswordFocus = true" @blur="isPasswordFocus = false" />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm">
              <el-input v-model.trim="resetForm.confirm" type="password" placeholder="再次输入新密码"
                :prefix-icon="Lock" show-password @focus="isPasswordFocus = true" @blur="isPasswordFocus = false" />
            </el-form-item>
            <el-button type="primary" class="submit" :loading="loading" @click="handleReset">重置密码</el-button>
            <p class="alt"><a @click="switchTab('login')">← 返回登录</a></p>
          </el-form>
        </div>
      </div>

      <p class="back"><router-link to="/home">返回首页</router-link></p>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, shallowRef } from "vue";
import { useRouter, useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import { User, Lock, Message, Key } from "@element-plus/icons-vue";
import { login, register, resetPassword } from "@/api/user";
import { uploadRegisterAvatar } from "@/api/upload";
import { useUserStore } from "@/stores/user";
import SendCodeButton from "@/components/SendCodeButton.vue";
import SquareImageCropper from "@/components/upload/SquareImageCropper.vue";
import LoginCaptcha from "@/components/auth/LoginCaptcha.vue";

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();

const tabs = [
  { key: "login", label: "登录" },
  { key: "register", label: "注册" },
];
const active = ref(route.query.tab === "register" ? "register" : "login");
const loading = ref(false);
const isPasswordFocus = ref(false);
const captchaLoading = shallowRef(false);
const captchaRefreshKey = shallowRef(0);

const loginRef = ref(null);
const regRef = ref(null);
const resetRef = ref(null);

const loginForm = reactive({ username: "", password: "", captchaId: "", captchaCode: "" });
const regForm = reactive({ avatar: "", username: "", email: "", code: "", password: "", confirm: "" });
const resetForm = reactive({ email: "", code: "", password: "", confirm: "" });

const loginRules = {
  username: [{ required: true, message: "请输入账号", trigger: "blur" }],
  password: [{ required: true, message: "请输入密码", trigger: "blur" }],
  captchaCode: [
    { required: true, message: "请输入图形验证码", trigger: "blur" },
    { pattern: /^\d{5}$/, message: "请输入 5 位数字验证码", trigger: "blur" },
  ],
};
const regRules = {
  username: [
    { required: true, message: "请输入账号", trigger: "blur" },
    { min: 2, max: 20, message: "账号 2-20 位", trigger: "blur" },
    { pattern: /^[A-Za-z0-9_]+$/, message: "只能包含字母、数字和下划线", trigger: "blur" },
  ],
  email: [
    { required: true, message: "请输入邮箱", trigger: "blur" },
    { type: "email", message: "邮箱格式不正确", trigger: "blur" },
  ],
  code: [{ required: true, message: "请输入验证码", trigger: "blur" }],
  password: [
    { required: true, message: "请输入密码", trigger: "blur" },
    { min: 6, message: "密码至少 6 位", trigger: "blur" },
  ],
  confirm: [
    { required: true, message: "请再次输入密码", trigger: "blur" },
    {
      validator: (_r, val, cb) =>
        val === regForm.password ? cb() : cb(new Error("两次密码不一致")),
      trigger: "blur",
    },
  ],
};
const resetRules = {
  email: [
    { required: true, message: "请输入邮箱", trigger: "blur" },
    { type: "email", message: "邮箱格式不正确", trigger: "blur" },
  ],
  code: [{ required: true, message: "请输入验证码", trigger: "blur" }],
  password: [
    { required: true, message: "请输入新密码", trigger: "blur" },
    { min: 6, message: "密码至少 6 位", trigger: "blur" },
  ],
  confirm: [
    { required: true, message: "请再次输入新密码", trigger: "blur" },
    {
      validator: (_r, val, cb) =>
        val === resetForm.password ? cb() : cb(new Error("两次密码不一致")),
      trigger: "blur",
    },
  ],
};

const switchTab = (key) => {
  active.value = key;
  loading.value = false;
};

const handleLogin = async () => {
  try {
    await loginRef.value.validate();
  } catch {
    return;
  }
  loading.value = true;
  try {
    const resp = await login(loginForm);
    userStore.setLogin(resp.data);
    ElMessage.success("登录成功");
    router.replace(route.query.redirect || "/");
  } catch (err) {
    console.error(err);
    captchaRefreshKey.value += 1;
  } finally {
    loading.value = false;
  }
};

const handleRegister = async () => {
  try {
    await regRef.value.validate();
  } catch {
    return;
  }
  loading.value = true;
  try {
    await register({
      username: regForm.username,
      password: regForm.password,
      email: regForm.email,
      code: regForm.code,
      avatar: regForm.avatar,
    });
    ElMessage.success("注册成功，请登录");
    loginForm.username = regForm.username;
    loginForm.password = "";
    regForm.password = "";
    regForm.confirm = "";
    regForm.code = "";
    regForm.avatar = "";
    active.value = "login";
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};

const handleReset = async () => {
  try {
    await resetRef.value.validate();
  } catch {
    return;
  }
  loading.value = true;
  try {
    await resetPassword({
      email: resetForm.email,
      code: resetForm.code,
      password: resetForm.password,
    });
    ElMessage.success("密码已重置，请用新密码登录");
    loginForm.username = resetForm.email;
    loginForm.password = "";
    Object.assign(resetForm, { email: "", code: "", password: "", confirm: "" });
    active.value = "login";
  } catch (err) {
    console.error(err);
  } finally {
    loading.value = false;
  }
};
</script>

<style scoped lang="scss">
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  background: linear-gradient(160deg, #bfe7df, #e4f5ed);
}

.login-page--register {
  align-items: flex-start;
  padding-top: 96px;
  padding-bottom: 32px;
}

.login-box {
  position: relative;
  width: 440px;
  max-width: 100%;
}

/* 毛玻璃卡片 */
.glass-card {
  position: relative;
  padding: 36px 26px 22px;
  border: 1px solid rgba(255, 255, 255, 0.5);
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.28);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  box-shadow: 0 8px 24px rgba(20, 40, 80, 0.14);
}

.brand {
  text-align: center;
  margin-bottom: 16px;

  h1 {
    margin: 0;
    font-size: 20px;
    font-weight: 700;
    letter-spacing: -0.01em;
    color: #2b3a3a;

    span {
      color: #2f7fe0;
    }

    small {
      margin-left: 7px;
      font-size: 12px;
      font-weight: 400;
      letter-spacing: 0.04em;
      color: #8a9995;
    }
  }
}

.tabs {
  display: flex;
  justify-content: center;
  gap: 28px;
  margin-bottom: 16px;

  .tab {
    padding: 2px 2px 6px;
    font-size: 15px;
    color: #5f6b68;
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    cursor: pointer;
    transition: color 0.2s, border-color 0.2s;

    &.active {
      color: #2f7fe0;
      border-bottom-color: #2f7fe0;
      font-weight: 500;
    }
  }
}

.submit {
  width: 100%;
  margin-top: 2px;
}

/* 验证码输入框 + 发送按钮同行 */
.code-row {
  display: flex;
  gap: 8px;
  width: 100%;

  :deep(.el-input) {
    flex: 1;
  }

  :deep(.el-button) {
    flex-shrink: 0;
    padding-left: 10px;
    padding-right: 10px;
  }
}

.alt {
  margin-top: 10px;
  text-align: right;
  font-size: 12px;

  a {
    color: #5f7570;
    cursor: pointer;

    &:hover {
      color: #2f7fe0;
    }
  }
}

.back {
  margin-top: 16px;
  text-align: center;
  font-size: 12px;

  a {
    color: #5f7570;

    &:hover {
      color: #2f7fe0;
    }
  }
}

/* el-input 半透明，贴合玻璃卡片 */
:deep(.el-form-item) {
  margin-bottom: 22px;
}

:deep(.el-form-item__label) {
  display: flex;
  align-items: center;
  font-size: 13px;
  color: #4a5654;
}

.avatar-form-item {
  :deep(.el-form-item__label) {
    height: 84px;
  }

  :deep(.el-form-item__content) {
    line-height: normal;
  }

  :deep(.image-field) {
    align-items: center;
    width: 100%;
  }

  :deep(.image-help) {
    justify-content: center;
  }
}

:deep(.el-input) {
  --el-input-height: 36px;
}

:deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.4);
  box-shadow: 0 0 0 1px rgba(255, 255, 255, 0.7) inset;
  border-radius: 3px;
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #2f7fe0 inset;
}

/* 猫头鹰（原图 + 原捂眼动画） */
.owl {
  width: 211px;
  height: 108px;
  background: url("@/image/owl-login.png") no-repeat;
  position: absolute;
  top: -92px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 2;
}

.owl .hand {
  width: 34px;
  height: 34px;
  border-radius: 40px;
  background-color: #472d20;
  position: absolute;
  left: 12px;
  bottom: -8px;
  transform: scaleY(0.6);
  transition: 0.3s ease-out;
}

.owl .hand.hand-r {
  left: 170px;
}

.owl.password .hand {
  transform: translateX(42px) translateY(-15px) scale(0.7);
}

.owl.password .hand.hand-r {
  transform: translateX(-42px) translateY(-15px) scale(0.7);
}

.owl .arms {
  position: absolute;
  top: 58px;
  width: 100%;
  height: 41px;
  overflow: hidden;
}

.owl .arms .arm {
  width: 40px;
  height: 65px;
  position: absolute;
  left: 20px;
  top: 40px;
  background: url("@/image/login-arm.png") no-repeat;
  transform: rotate(-20deg);
  transition: 0.3s ease-out;
}

.owl .arms .arm.arm-r {
  transform: rotate(20deg) scaleX(-1);
  left: 158px;
}

.owl.password .arms .arm {
  transform: translateY(-40px) translateX(40px);
}

.owl.password .arms .arm.arm-r {
  transform: translateY(-40px) translateX(-40px) scaleX(-1);
}
</style>
