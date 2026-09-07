import { useUserStore } from '@/stores/user'

// v-permission="['admin']" / v-permission="['admin','super_admin']"
// role=1 → 'admin'，其它 → 'user'
export default {
  mounted(el, binding) {
    const userStore = useUserStore()
    const role = userStore.user?.role === 1 ? 'admin' : 'user'
    const allowed = binding.value
    if (!Array.isArray(allowed) || !allowed.includes(role)) {
      el.remove()
    }
  },
}
