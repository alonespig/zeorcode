import { computed, shallowRef } from "vue";
import { defineStore } from "pinia";
import { store } from "../utils";
import {
  getSession,
  login as loginApi,
  logout as logoutApi,
  type AdminUser,
  type LoginParams
} from "@/api/auth";
import { router, resetRouter, routerArrays } from "../utils";
import { useMultiTagsStoreHook } from "./multiTags";

export const useUserStore = defineStore("admin-user", () => {
  const user = shallowRef<AdminUser | null>(null);
  const initialized = shallowRef(false);

  const authenticated = computed(() => user.value !== null);
  const isAdmin = computed(() => user.value?.role === 1);
  const avatar = computed(() => user.value?.avatar || "");
  const username = computed(() => user.value?.username || "");
  const nickname = computed(() => user.value?.username || "");

  function setUser(value: AdminUser | null, avatarOverride?: string) {
    user.value = value
      ? { ...value, avatar: avatarOverride || value.avatar || "" }
      : null;
  }

  function clearSession() {
    setUser(null);
    initialized.value = true;
  }

  async function restoreSession(force = false) {
    if (initialized.value && !force) return authenticated.value;
    try {
      const response = await getSession();
      const session = response.data;
      if (!session?.authenticated || !session.user) {
        clearSession();
        return false;
      }
      setUser(session.user, session.avatar);
      initialized.value = true;
      return true;
    } catch {
      clearSession();
      return false;
    }
  }

  async function login(params: LoginParams) {
    const response = await loginApi(params);
    if (!response.data?.user) throw new Error("登录响应缺少用户信息");

    setUser(response.data.user, response.data.avatar);
    initialized.value = true;
    if (!isAdmin.value) {
      await logoutApi().catch(() => undefined);
      clearSession();
      throw new Error("当前账号没有后台管理权限");
    }
  }

  async function logOut() {
    try {
      await logoutApi();
    } finally {
      clearSession();
      useMultiTagsStoreHook().handleTags("equal", [...routerArrays]);
      resetRouter();
      await router.push("/login");
    }
  }

  return {
    user,
    initialized,
    authenticated,
    isAdmin,
    avatar,
    username,
    nickname,
    setUser,
    clearSession,
    restoreSession,
    login,
    logOut
  };
});

export function useUserStoreHook() {
  return useUserStore(store);
}
