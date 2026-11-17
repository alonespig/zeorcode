import { defineStore } from 'pinia'
import { shallowRef, watch } from 'vue'
import { getUnreadCount, markNotificationRead } from '@/api/notification'
import { useUserStore } from '@/stores/user'

const emptyUnread = () => ({ comment: 0, reply: 0, like: 0, system: 0, rating: 0, total: 0 })

export const useNotificationStore = defineStore('notification', () => {
  const userStore = useUserStore()
  const unread = shallowRef(emptyUnread())
  const markingRead = shallowRef(false)
  let refreshRequest = null
  let sessionVersion = 0

  watch(() => userStore.user?.id, () => {
    sessionVersion++
    refreshRequest = null
    unread.value = emptyUnread()
    markingRead.value = false
  }, { flush: 'sync' })

  const refreshUnread = () => {
    if (!userStore.isLogin || markingRead.value) return
    if (refreshRequest) return refreshRequest

    const request = getUnreadCount().then(({ data }) => {
      if (refreshRequest === request && data) unread.value = { ...emptyUnread(), ...data }
    }).catch(() => {
      // 暂时取不到计数时保留上一次的结果。
    }).finally(() => {
      if (refreshRequest === request) refreshRequest = null
    })
    refreshRequest = request
    return request
  }

  // 传入通知标记单条，不传则标记全部。
  const markRead = async (notification) => {
    if (!userStore.isLogin || markingRead.value) return false
    const version = sessionVersion
    markingRead.value = true
    // 作废正在进行的轮询，避免旧计数覆盖本次已读结果。
    refreshRequest = null
    try {
      await markNotificationRead(notification ? { id: notification.id } : {})
      if (version !== sessionVersion) return false

      if (notification) {
        const { type } = notification
        const counts = { ...unread.value }
        counts.total = Math.max(0, counts.total - 1)
        if (type !== 'total' && Object.hasOwn(counts, type)) {
          counts[type] = Math.max(0, counts[type] - 1)
        }
        unread.value = counts
      } else {
        unread.value = emptyUnread()
      }
      return true
    } catch {
      return false
    } finally {
      if (version === sessionVersion) {
        markingRead.value = false
        refreshUnread()
      }
    }
  }

  return { unread, markingRead, refreshUnread, markRead }
})
