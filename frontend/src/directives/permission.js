import { useUserStore } from '@/stores/user'

// v-permission="['admin']"：role>=1 即可
// v-permission="'problem.write'"：按权限点判断
// v-permission="['problem.read', 'problem.write']"：满足任意一个即可
export default {
  mounted(el, binding) {
    const userStore = useUserStore()
    const value = binding.value
    const list = Array.isArray(value) ? value : [value]
    const allowed = list.some(item =>
      item === 'admin' ? userStore.isAdmin : userStore.can(item)
    )
    if (!allowed) el.remove()
  },
}
