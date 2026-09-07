<template>
  <div class="acm-header f-shadow">
    <div class="acm-container">
      <button
        class="mobile-nav-toggle"
        type="button"
        :aria-expanded="isNavOpen"
        aria-label="切换导航菜单"
        @click="toggleNav"
      >
        <el-icon :size="18">
          <Close v-if="isNavOpen" />
          <Menu v-else />
        </el-icon>
        <span>菜单</span>
      </button>

      <ul class="acm-nav" :class="{ 'is-open': isNavOpen }">
        <li v-for="item in navItems" :key="item.to">
          <RouterLink :to="item.to" active-class="nav-link-active" @click="closeNav">
            <span :class="['iconfont', item.icon]"></span>
            <span>{{ item.label }}</span>
          </RouterLink>
        </li>
      </ul>
      <div class="nav-right">
        <template v-if="userStore.isLogin">
          <el-badge :value="unreadTotal" :hidden="!unreadTotal" :max="99" class="bell">
            <span @click="router.push('/message')" class="iconfont icon-xinfengtianchong text-xl!"></span>
          </el-badge>
          <div class="nav-item user-avatar">
            <el-dropdown>
              <span class="el-dropdown-link">
                <el-avatar :size="36" :src="userStore.user?.avatar" />
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="$router.push({ name: 'User', params: { id: userStore.user?.id } })">我的资料</el-dropdown-item>
                  <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
        <template v-else>
          <button class="btn-auth ghost" @click="router.push('/login')">登录</button>
          <button class="btn-auth solid" @click="router.push({ path: '/login', query: { tab: 'register' } })">
            注册
          </button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, shallowRef, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getUnreadCount } from '@/api/notification'
import { Close, Menu } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const isNavOpen = shallowRef(false)
const toggleNav = () => { isNavOpen.value = !isNavOpen.value }
const closeNav = () => { isNavOpen.value = false }

const handleLogout = async () => {
  try {
    await userStore.logout()
    router.push('/home')
  } catch {
    // 请求层已经提示错误；保留本地登录态，避免与仍有效的 Cookie 不一致。
  }
}

// 未读通知铃铛：登录后轮询
const unreadTotal = ref(0)
let unreadTimer = null
const loadUnread = async () => {
  if (!userStore.isLogin) {
    unreadTotal.value = 0
    return
  }
  const res = await getUnreadCount().catch(() => null)
  unreadTotal.value = res?.data?.total || 0
}
onMounted(() => {
  loadUnread()
  unreadTimer = setInterval(loadUnread, 45000)
})
onUnmounted(() => unreadTimer && clearInterval(unreadTimer))
watch(() => userStore.isLogin, loadUnread)
watch(() => route.fullPath, closeNav)

const navItems = computed(() => {
  const items = [
    { to: '/home', label: '首页', icon: 'icon-home3' },
    { to: '/contest', label: '比赛', icon: 'icon-jiangbei-' },
    { to: '/problem', label: '题库', icon: 'icon-xinxiliebiao' },
    { to: '/problemset', label: '题单', icon: 'icon-shuji' },
    { to: '/team', label: '团队', icon: 'icon-renqun' },
    { to: '/blog', label: '社区', icon: 'icon-jurassic_bbs' },
    { to: '/submission', label: '评测', icon: 'icon-shalou1' },
    { to: '/rank', label: '排名', icon: 'icon-paixingbang' },
  ]
  if (userStore.isAdmin) {
    items.push({ to: '/admin', label: '后台', icon: 'icon-fuwuqi' })
  }
  return items
})
</script>


<style scoped lang="scss">
.acm-header {
  height: 58px;
  font-size: 16px;
  background: #fff;
  border-bottom: 1px solid #eee;
}

.acm-container {
  max-width: 1260px;
  padding: 0 20px;
  margin: 0 auto;
  height: 100%;
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mobile-nav-toggle {
  display: none;
}

.acm-nav {
  height: 100%;
  display: flex;
  gap: 20px;
  align-items: center;

  li {
    height: 100%;
  }

  a {
    display: flex;
    align-items: center;
    height: 100%;
    gap: 3px;
    box-sizing: border-box;
    padding: 0 6px;
    color: #333;
    transition: background-color 0.3s, color 0.3s, border-color 0.3s;
    border-bottom: 2px solid transparent;

    &:hover {
      background-color: #f5f7fa;
      border-bottom: 2px solid #1f2020;
      color: #007bff;
    }

    &.nav-link-active {
      color: #007bff;
      background: #f2f3f4;
      border-bottom: 2px solid #007bff;
    }
  }
}

.user-avatar {
  &:hover {
    cursor: pointer;
  }
}

.bell {
  cursor: pointer;
  color: #4a5058;
  display: flex;
  align-items: center;

  &:hover {
    color: #007bff;
  }
}

.nav-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.btn-auth {
  height: 32px;
  padding: 0 16px;
  font-size: 14px;
  border-radius: 4px;
  border: 1px solid transparent;
  cursor: pointer;
  transition: background-color 0.2s, border-color 0.2s, color 0.2s;

  &.ghost {
    background: #fff;
    border-color: #dcdfe6;
    color: #4a5058;

    &:hover {
      border-color: #007bff;
      color: #007bff;
    }
  }

  &.solid {
    background: #007bff;
    border-color: #007bff;
    color: #fff;

    &:hover {
      background: #1a86ff;
      border-color: #1a86ff;
    }
  }
}

@media (max-width: 767px) {
  .acm-header {
    height: 54px;
  }

  .acm-container {
    padding: 0 12px;
  }

  .mobile-nav-toggle {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 36px;
    padding: 0 10px;
    border-radius: var(--radius-base);
    color: #333;
    font-size: 15px;

    &:hover {
      background-color: #f5f7fa;
      color: #007bff;
    }

    &:focus-visible {
      outline: 2px solid rgba(0, 123, 255, 0.45);
      outline-offset: 2px;
    }
  }

  .acm-nav {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    z-index: 110;
    display: none;
    height: auto;
    gap: 0;
    padding: 8px 12px 12px;
    background: #fff;
    border-top: 1px solid #eef0f3;
    border-bottom: 1px solid #e5e7eb;
    box-shadow: var(--shadow-card);

    &.is-open {
      display: flex;
      flex-direction: column;
    }

    li {
      width: 100%;
      height: 42px;
    }

    a {
      width: 100%;
      padding: 0 12px;
      border-bottom: 0;
      border-radius: var(--radius-base);

      &:hover,
      &.nav-link-active {
        border-bottom: 0;
      }
    }
  }

  .nav-right {
    gap: 10px;
  }

  .btn-auth {
    height: 32px;
    padding: 0 12px;
  }
}
</style>
