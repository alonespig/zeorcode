import { cloneDeep, isUrl, openLink } from "@pureadmin/utils";
import {
  createRouter,
  type RouteComponent,
  type RouteRecordRaw
} from "vue-router";
import { getConfig } from "@/config";
import NProgress from "@/utils/progress";
import { buildHierarchyTree } from "@/utils/tree";
import remainingRouter from "./modules/remaining";
import { usePermissionStoreHook } from "@/store/modules/permission";
import { useUserStoreHook } from "@/store/modules/user";
import {
  ascending,
  getHistoryMode,
  formatTwoStageRoutes,
  formatFlatteningRoutes,
  handleAliveRoute
} from "./utils";

const modules: Record<string, any> = import.meta.glob(
  ["./modules/**/*.ts", "!./modules/**/remaining.ts"],
  { eager: true }
);

const routes: RouteRecordRaw[] = [];
Object.keys(modules).forEach(key => routes.push(modules[key].default));

export const constantRoutes: RouteRecordRaw[] = formatTwoStageRoutes(
  formatFlatteningRoutes(buildHierarchyTree(ascending(routes.flat(Infinity))))
);

const initConstantRoutes = cloneDeep(constantRoutes);

export const constantMenus: RouteComponent[] = ascending(
  routes.flat(Infinity)
).concat(...remainingRouter);

export const remainingPaths = Object.keys(remainingRouter).map(
  key => remainingRouter[key].path
);

export const router = createRouter({
  history: getHistoryMode(import.meta.env.VITE_ROUTER_HISTORY),
  routes: constantRoutes.concat(...(remainingRouter as any)),
  strict: true,
  scrollBehavior(_to, _from, savedPosition) {
    return savedPosition || { left: 0, top: 0 };
  }
});

const loadedPaths = new Set<string>();

export function resetLoadedPaths() {
  loadedPaths.clear();
}

export function resetRouter() {
  router.clearRoutes();
  for (const route of initConstantRoutes.concat(...(remainingRouter as any))) {
    router.addRoute(route);
  }
  usePermissionStoreHook().clearAllCachePage();
  usePermissionStoreHook().handleWholeMenus([]);
  resetLoadedPaths();
}

router.beforeEach(async (to, from) => {
  to.meta.loaded = loadedPaths.has(to.path);
  if (!to.meta.loaded) NProgress.start();

  if (to.meta?.keepAlive) {
    handleAliveRoute(to, "add");
    if (from.name === undefined || from.name === "Redirect") {
      handleAliveRoute(to);
    }
  }

  if (!isUrl(to.name as string)) {
    const title = to.meta?.title;
    const platformTitle = getConfig().Title;
    if (title) {
      document.title = platformTitle ? `${title} | ${platformTitle}` : title;
    }
  } else {
    openLink(to.name as string);
    NProgress.done();
    return false;
  }

  const permissionStore = usePermissionStoreHook();
  if (permissionStore.wholeMenus.length === 0) {
    permissionStore.handleWholeMenus([]);
  }

  const userStore = useUserStoreHook();
  if (!userStore.initialized) await userStore.restoreSession();

  if (to.path === "/login") {
    return userStore.isAdmin ? "/dashboard" : true;
  }

  if (!userStore.authenticated) {
    return {
      path: "/login",
      query:
        to.fullPath === "/dashboard" ? undefined : { redirect: to.fullPath }
    };
  }

  if (!userStore.isAdmin) {
    return to.path === "/access-denied" ? true : "/access-denied";
  }

  return true;
});

router.afterEach(to => {
  loadedPaths.add(to.path);
  NProgress.done();
});

export default router;
