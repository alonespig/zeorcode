import { defineStore } from "pinia";
import {
  type cacheType,
  store,
  ascending,
  getKeyList,
  filterTree,
  constantMenus,
  formatFlatteningRoutes
} from "../utils";
import { useMultiTagsStoreHook } from "./multiTags";
import { useUserStoreHook } from "./user";

/** 按当前用户权限过滤菜单：叶子无权限则隐藏，父级子项全空则整组隐藏。 */
function filterByPermission(menus: any[]): any[] {
  const userStore = useUserStoreHook();
  return menus.filter(menu => {
    if (menu.meta?.permission && !userStore.can(menu.meta.permission)) {
      return false;
    }
    if (Array.isArray(menu.children) && menu.children.length > 0) {
      menu.children = filterByPermission(menu.children);
      return menu.children.length > 0;
    }
    return true;
  });
}

export const usePermissionStore = defineStore("pure-permission", {
  state: () => ({
    // 静态路由生成的菜单
    constantMenus,
    // 整体路由生成的菜单（静态、动态）
    wholeMenus: [],
    // 整体路由（一维数组格式）
    flatteningRoutes: [],
    // 缓存页面keepAlive
    cachePageList: []
  }),
  actions: {
    /** 根据本地静态路由生成菜单。 */
    handleWholeMenus(routes: any[] = []) {
      this.wholeMenus = filterByPermission(
        filterTree(ascending(this.constantMenus.concat(routes)))
      );
      this.flatteningRoutes = formatFlatteningRoutes(
        this.constantMenus.concat(routes) as any
      );
    },
    /** 监听缓存页面是否存在于标签页，不存在则删除 */
    clearCache() {
      let cacheLength = this.cachePageList.length;
      const nameList = getKeyList(useMultiTagsStoreHook().multiTags, "name");
      while (cacheLength > 0) {
        nameList.findIndex(v => v === this.cachePageList[cacheLength - 1]) ===
          -1 &&
          this.cachePageList.splice(
            this.cachePageList.indexOf(this.cachePageList[cacheLength - 1]),
            1
          );
        cacheLength--;
      }
    },
    cacheOperate({ mode, name }: cacheType) {
      const delIndex = this.cachePageList.findIndex(v => v === name);
      switch (mode) {
        case "refresh":
          this.cachePageList = this.cachePageList.filter(v => v !== name);
          this.clearCache();
          break;
        case "add":
          this.cachePageList.push(name);
          break;
        case "delete":
          delIndex !== -1 && this.cachePageList.splice(delIndex, 1);
          this.clearCache();
          break;
      }
    },
    /** 清空缓存页面 */
    clearAllCachePage() {
      this.wholeMenus = [];
      this.cachePageList = [];
    }
  }
});

export function usePermissionStoreHook() {
  return usePermissionStore(store);
}
