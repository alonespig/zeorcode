import { defineStore } from "pinia";
import { ref, computed } from "vue";
import request from "@/utils/request";

export const useUserStore = defineStore("user",()=>{
  const user = ref(null)
  const isLogin = computed(() => user.value !== null)
  const isAdmin = computed(() => user.value?.role === 1 || false)

  const setLogin = (data) => {
    user.value = data?.user || null
    if (data?.avatar) {
      user.value = { ...user.value, avatar: data.avatar }
    }
    localStorage.setItem('user', JSON.stringify(user.value))
  }

  const loadFromStorage = () => {
    try {
      user.value = JSON.parse(localStorage.getItem('user') || 'null')
    } catch {
      localStorage.removeItem('user')
      user.value = null
    }
  }

  // localLogout 仅清除本地状态（用于鉴权失败时的被动清理）
  const localLogout = () => {
    user.value = null
    localStorage.removeItem('user')
  }

  // restoreSession 以服务端 HttpOnly Cookie 为准同步登录态，localStorage 只作为展示缓存。
  const restoreSession = async () => {
    try {
      const res = await request({ url: '/session', method: 'get' })
      if (!res.data?.authenticated || !res.data?.user) {
        localLogout()
        return false
      }
      setLogin({ user: res.data.user, avatar: res.data.avatar })
      return true
    } catch {
      // 无法确认服务端身份时按未登录处理，避免展示过期的管理员权限。
      localLogout()
      return false
    }
  }

  // logout 主动退出：只有服务端成功清除 Cookie 后才清本地状态。
  const logout = async () => {
    await request({ url: '/logout', method: 'post' })
    localLogout()
  }

  // promptLogin 未登录时跳转登录页（带回跳地址），替代原来的登录弹窗
  const promptLogin = () => {
    import("@/router").then(({ default: router }) => {
      const redirect = router.currentRoute.value?.fullPath
      router.push({
        path: "/login",
        query: redirect && redirect !== "/login" ? { redirect } : {},
      })
    })
  }

  return {
    user,
    isAdmin,
    isLogin,
    setLogin,
    loadFromStorage,
    restoreSession,
    logout,
    localLogout,
    promptLogin,
  }
} );
