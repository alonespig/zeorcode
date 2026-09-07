import { createRouter, createWebHistory } from "vue-router";
import NProgress from "nprogress";
import "nprogress/nprogress.css";
import { useUserStore } from "@/stores/user"

// 顶部路由进度条：切页面时显示，无转圈小球
NProgress.configure({ showSpinner: false, speed: 400, trickleSpeed: 120 })
import { contestRoutes } from "./modules/contest"
import { problemRoutes } from "./modules/problem"
import { problemSetRoutes } from "./modules/problemset"
import { teamRoutes } from "./modules/team"
import { submissionRoutes } from "./modules/submission"
import { adminRoutes } from "./modules/admin"
import { postRoutes } from "./modules/post"

const routes = [
  {
    path: "/",
    name: "Layout",
    redirect: "/home",
    component: () => import("@/pages/Layout/index.vue"),
    children: [
      {
        path: "home",
        name: "Home",
        meta: { title: "首页" },
        component: () => import("@/pages/Home/index.vue"),
      },
      contestRoutes,
      problemRoutes,
      problemSetRoutes,
      teamRoutes,
      postRoutes,
      {
        path: "rank",
        name: "Rank",
        meta: { title: "排名" },
        component: () => import("@/pages/Rank/index.vue"),
      },
      {
        path: "message",
        name: "Message",
        meta: { title: "消息", requiresAuth: true },
        component: () => import("@/pages/Message/index.vue"),
      },
      {
        path: "training",
        name: "Training",
        meta: { title: "训练" },
        component: () => import("@/pages/Training/index.vue"),
      },
      submissionRoutes,
      {
        // 个人中心：布局(头像资料头 + tab 栏) + 子路由 tab。公开可看，任何人可访问 /user/:id
        path: 'user/:id',
        component: () => import("@/pages/User/index.vue"),
        children: [
          {
            path: '',
            name: 'User',
            meta: { title: "个人中心" },
            component: () => import("@/pages/User/tabs/Overview.vue"),
          },
          {
            path: 'problems',
            name: 'UserProblems',
            meta: { title: "题目" },
            component: () => import("@/pages/User/tabs/Problems.vue"),
          },
          {
            path: 'posts',
            name: 'UserPosts',
            meta: { title: "帖子" },
            component: () => import("@/pages/User/tabs/Posts.vue"),
          },
          {
            path: 'submissions',
            name: 'UserSubmissions',
            meta: { title: "提交记录" },
            component: () => import("@/pages/User/tabs/Submissions.vue"),
          },
        ],
      },
      adminRoutes,
    ],
  },
  {
    path: "/login",
    name: "Login",
    meta: { title: "登录" },
    component: () => import("@/pages/Login/login.vue"),
  },
  {
    path: "/404",
    name: "404",
    meta: { title: "404" },
    component: () => import("@/pages/404.vue"),
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/404",
  },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

const SITE_TITLE = 'ZeorCode'
const whiteList = ['/login', '/404']

router.beforeEach((to, from, next) => {
  NProgress.start()
  const userStore = useUserStore()

  if (whiteList.includes(to.path)) {
    if (to.path === '/login' && userStore.isLogin) {
      next('/')
      return
    }
    next()
    return
  }
  if (!userStore.isLogin && to.meta.requiresAuth) {
    next({ path: '/login', query: { redirect: to.fullPath } })
    return
  }
  if (to.meta.requiresAdmin && !userStore.isAdmin) {
    next('/')
    return
  }
  next()
})

router.afterEach((to) => {
  NProgress.done()
  document.title = to.meta.title ? `${to.meta.title} - ${SITE_TITLE}` : SITE_TITLE
})

// 导航被中断/出错时也收尾，避免进度条卡住
router.onError(() => {
  NProgress.done()
})

export default router;
