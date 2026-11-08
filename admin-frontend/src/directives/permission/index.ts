import type { Directive } from "vue";
import { useUserStoreHook } from "@/store/modules/user";

/**
 * v-permission="'problem.write'" 或 v-permission="['problem.read', 'problem.write']"
 * 数组表示拥有其中任意一个即可；超级管理员始终通过。
 */
export const permission: Directive<HTMLElement, string | string[]> = {
  mounted(el, binding) {
    const value = binding.value;
    const userStore = useUserStoreHook();
    const allowed = Array.isArray(value)
      ? value.some(item => userStore.can(item))
      : userStore.can(value);
    if (!allowed) {
      el.parentNode?.removeChild(el);
    }
  }
};
