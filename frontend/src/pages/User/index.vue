<template>
  <div class="mx-auto max-w-[1200px] pt-17 pb-6">

    <!-- 资料头（头像居中） -->
    <div class="relative flex flex-col items-center gap-2 rounded border border-gray-200 bg-white px-6 pt-20 pb-7 text-center">
      <el-button v-if="isSelf" class="!absolute !top-4 !right-4" size="small" @click="openEdit">编辑资料</el-button>

      <el-avatar :size="125" :src="user?.avatar" class="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 shadow-[0_0_0_3px_#fff,0_0_0_4px_#e6e9ee]">
        {{ user?.name?.charAt(0)?.toUpperCase() }}
      </el-avatar>

      <div class="text-[22px] font-semibold leading-tight" :style="{ color: nameColor }">{{ user?.name }}</div>
      <div class="text-[12.5px] font-medium" :style="{ color: ratingColor }">
        {{ tierName }}<span v-if="rating > 0"> · {{ rating }}</span>
      </div>
      <div v-if="user?.signature" class="text-[13px] text-gray-500">{{ user.signature }}</div>

      <div class="mt-0.5 flex flex-wrap justify-center gap-5 text-[12.5px] text-gray-400">
        <span v-if="user?.school" class="inline-flex items-center gap-1">🎓 {{ user.school }}</span>
        <span class="inline-flex items-center gap-1">
          <span class="iconfont icon-weibiaoti-1-13"></span>加入于 {{ user?.createdAt }}
        </span>
        <span v-if="isSelf" class="inline-flex items-center gap-1">
          <span class="iconfont icon-xinfengtianchong"></span>{{ user?.email }}
        </span>
      </div>

      <div class="mt-2 flex divide-x divide-gray-100 [&>div]:px-7">
        <div>
          <div class="text-[23px] font-semibold tabular-nums text-gray-800">{{ solved }}</div>
          <div class="mt-0.5 text-[12.5px] text-gray-500">已解决</div>
        </div>
        <div>
          <div class="text-[23px] font-semibold tabular-nums text-amber-600">{{ attempted }}</div>
          <div class="mt-0.5 text-[12.5px] text-gray-500">尝试中</div>
        </div>
        <div>
          <div class="text-[23px] font-semibold tabular-nums text-gray-800">{{ passRate }}</div>
          <div class="mt-0.5 text-[12.5px] text-gray-500">通过率</div>
        </div>
        <div>
          <div class="text-[23px] font-semibold tabular-nums" :style="{ color: ratingColor }">{{ rating > 0 ? rating : '—'
          }}</div>
          <div class="mt-0.5 text-[12.5px] text-gray-500">Rating</div>
        </div>
      </div>
    </div>

    <!-- tab 栏 + 内容（子路由） -->
    <div class="mt-4 pt-2 overflow-hidden rounded border border-gray-200 bg-white">
      <div class="flex gap-2.5 border-b border-gray-200 px-2.5">
        <span v-for="t in tabs" :key="t.name" @click="go(t.name)"
          class="-mb-px cursor-pointer border-b-2 px-2 py-1.5 text-[15px] transition-colors"
          :class="route.name === t.name
            ? 'border-blue-500 font-medium text-blue-500'
            : 'border-transparent text-gray-600 hover:text-blue-500'">
          {{ t.label }}
        </span>
      </div>
      <router-view />
    </div>

    <!-- 编辑资料（仅本人） -->
    <el-dialog v-model="editVisible" title="编辑资料" width="460px">
      <el-form label-width="72px">
        <el-form-item label="头像">
          <SquareImageCropper
            v-model="editForm.avatar"
            :preview-size="80"
            round
            title="裁剪头像"
            empty-text="选择头像"
            help-text="PNG/JPG，最大 5MB"
            remove-text="移除头像"
            success-message="头像已上传"
            file-prefix="avatar"
            preview-alt="用户头像预览"
          />
        </el-form-item>
        <el-form-item label="性别">
          <el-radio-group v-model="editForm.gender">
            <el-radio :value="1">男</el-radio>
            <el-radio :value="2">女</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="邮箱">
          <div class="flex w-full items-center gap-2">
            <span class="text-sm text-gray-600">{{ editForm.email }}</span>
            <el-button size="small" @click="openBind">换绑</el-button>
          </div>
        </el-form-item>
        <el-form-item label="账号安全">
          <el-button size="small" @click="openPassword">修改密码</el-button>
        </el-form-item>
        <el-form-item label="学校">
          <el-input v-model="editForm.school" maxlength="50" show-word-limit placeholder="就读/所在学校" clearable />
        </el-form-item>
        <el-form-item label="个性签名">
          <el-input v-model="editForm.signature" maxlength="25" show-word-limit placeholder="一句话介绍自己" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveProfile">保存</el-button>
      </template>
    </el-dialog>

    <!-- 修改密码：必须校验当前密码，成功后退出登录 -->
    <el-dialog v-model="passwordVisible" title="修改密码" width="420px" append-to-body @closed="resetPasswordForm">
      <el-form ref="passwordRef" :model="passwordForm" :rules="passwordRules" label-width="88px">
        <el-form-item label="当前密码" prop="currentPassword">
          <el-input v-model="passwordForm.currentPassword" type="password" show-password autocomplete="current-password" />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input v-model="passwordForm.newPassword" type="password" show-password autocomplete="new-password" placeholder="至少 6 位" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="passwordForm.confirmPassword" type="password" show-password autocomplete="new-password" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordVisible = false">取消</el-button>
        <el-button type="primary" :loading="changingPassword" @click="savePassword">确认修改</el-button>
      </template>
    </el-dialog>

    <!-- 换绑邮箱（需邮箱验证码） -->
    <el-dialog v-model="bindVisible" title="换绑邮箱" width="420px" append-to-body>
      <el-form ref="bindRef" :model="bindForm" :rules="bindRules" label-width="72px">
        <el-form-item label="新邮箱" prop="email">
          <el-input v-model.trim="bindForm.email" placeholder="请输入新邮箱" />
        </el-form-item>
        <el-form-item label="验证码" prop="code">
          <div class="flex w-full gap-2">
            <el-input v-model.trim="bindForm.code" placeholder="邮箱验证码" class="flex-1" />
            <SendCodeButton :email="bindForm.email" scene="bind" />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="bindVisible = false">取消</el-button>
        <el-button type="primary" :loading="binding" @click="saveBind">确定换绑</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, provide, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getUserProfile, getUserRating, updateUserInfo, bindEmail, userInfo, changePassword } from '@/api/user'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { ratingTier } from '@/constants/index'
import SendCodeButton from '@/components/SendCodeButton.vue'
import SquareImageCropper from '@/components/upload/SquareImageCropper.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 整份资料（含 user / solveItem / unsolveItem）向子 tab 注入，避免重复请求
const profile = ref()
provide('userProfile', profile)

// rating（当前分 + 历史）向主页 tab 注入，画折线图
const ratingData = ref(null)
provide('userRating', ratingData)
const rating = computed(() => ratingData.value?.rating || 0)
const tierObj = computed(() => ratingTier(rating.value))
const tierName = computed(() => tierObj.value.name)
const ratingColor = computed(() => tierObj.value.color)            // 段位标签 + Rating 数字
const nameColor = computed(() => (rating.value > 0 ? tierObj.value.color : '#334155')) // 未定级名字用深灰保证可读

const user = computed(() => profile.value?.user)
const isSelf = computed(
  () => userStore.isLogin && String(userStore.user?.id) === String(route.params.id)
)

const tabs = [
  { name: 'User', label: '主页' },
  { name: 'UserProblems', label: '题目' },
  { name: 'UserPosts', label: '帖子' },
  { name: 'UserSubmissions', label: '提交记录' },
]

const go = (name) => {
  if (route.name !== name) router.push({ name, params: { id: route.params.id } })
}

const solved = computed(() => profile.value?.solveItem?.length || 0)
const attempted = computed(() => profile.value?.unsolveItem?.length || 0)
const passRate = computed(() => {
  const total = solved.value + attempted.value
  return total ? Math.round((solved.value / total) * 100) + '%' : '0%'
})

const fetchProfile = async () => {
  const res = await getUserProfile(route.params.id)
  profile.value = res.data
  if (userStore.isLogin && String(userStore.user?.id) === String(route.params.id)) {
    const selfRes = await userInfo().catch(() => null)
    if (selfRes?.data && profile.value?.user) {
      Object.assign(profile.value.user, {
        email: selfRes.data.email || '',
        gender: selfRes.data.gender || profile.value.user.gender,
        signature: selfRes.data.signature ?? profile.value.user.signature,
        school: selfRes.data.school ?? profile.value.user.school,
      })
    }
  }
  const r = await getUserRating(route.params.id).catch(() => null)
  ratingData.value = r?.data || null
}

// ---------- 编辑资料 ----------
const editVisible = ref(false)
const saving = ref(false)
const editForm = ref({ avatar: '', gender: 1, email: '', signature: '', school: '' })

const openEdit = () => {
  const u = profile.value?.user || {}
  editForm.value = {
    avatar: u.avatar || '',
    gender: u.gender || 1,
    email: u.email || '',
    signature: u.signature || '',
    school: u.school || '',
  }
  editVisible.value = true
}

const saveProfile = async () => {
  saving.value = true
  try {
    await updateUserInfo({
      avatar: editForm.value.avatar,
      gender: editForm.value.gender,
      signature: editForm.value.signature,
      school: editForm.value.school,
    })
    ElMessage.success('保存成功')
    editVisible.value = false
    await fetchProfile()
    if (userStore.user) userStore.user.avatar = editForm.value.avatar
  } catch (err) {
    console.error(err)
  } finally {
    saving.value = false
  }
}

// ---------- 修改密码 ----------
const passwordVisible = ref(false)
const changingPassword = ref(false)
const passwordRef = ref(null)
const passwordForm = ref({ currentPassword: '', newPassword: '', confirmPassword: '' })

const validatePasswordBytes = (_rule, value, callback) => {
  if (new TextEncoder().encode(value || '').length > 72) {
    callback(new Error('密码不能超过 72 个字节'))
    return
  }
  callback()
}

const passwordRules = {
  currentPassword: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' },
    { validator: validatePasswordBytes, trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => (
        value === passwordForm.value.newPassword
          ? callback()
          : callback(new Error('两次密码不一致'))
      ),
      trigger: 'blur',
    },
  ],
}

const resetPasswordForm = () => {
  passwordForm.value = { currentPassword: '', newPassword: '', confirmPassword: '' }
  passwordRef.value?.clearValidate()
}

const openPassword = () => {
  resetPasswordForm()
  passwordVisible.value = true
}

const savePassword = async () => {
  try {
    await passwordRef.value.validate()
  } catch {
    return
  }
  changingPassword.value = true
  try {
    await changePassword({
      currentPassword: passwordForm.value.currentPassword,
      newPassword: passwordForm.value.newPassword,
    })
    ElMessage.success('密码修改成功，请重新登录')
    passwordVisible.value = false
    editVisible.value = false
    userStore.localLogout()
    await router.replace({ path: '/login', query: { redirect: route.fullPath } })
  } catch (err) {
    console.error(err)
  } finally {
    changingPassword.value = false
  }
}

// ---------- 换绑邮箱（需验证码） ----------
const bindVisible = ref(false)
const binding = ref(false)
const bindRef = ref(null)
const bindForm = ref({ email: '', code: '' })
const bindRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  code: [{ required: true, message: '请输入验证码', trigger: 'blur' }],
}

const openBind = () => {
  bindForm.value = { email: '', code: '' }
  bindVisible.value = true
}

const saveBind = async () => {
  try {
    await bindRef.value.validate()
  } catch {
    return
  }
  binding.value = true
  try {
    await bindEmail({ email: bindForm.value.email, code: bindForm.value.code })
    ElMessage.success('邮箱换绑成功')
    editForm.value.email = bindForm.value.email
    bindVisible.value = false
    await fetchProfile()
  } catch (err) {
    console.error(err)
  } finally {
    binding.value = false
  }
}

let oldTitle = ''

onMounted(async () => {
  await fetchProfile()
  if (user.value?.name) {
    oldTitle = document.title
    document.title = user.value.name
  }
})

// 从一个用户切到另一个用户时重新拉取
watch(() => route.params.id, fetchProfile)

onBeforeUnmount(() => {
  if (oldTitle) document.title = oldTitle
})
</script>
