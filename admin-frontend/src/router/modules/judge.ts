const Layout = () => import("@/layout/index.vue");

export default {
  path: "/judge",
  name: "AdminJudge",
  component: Layout,
  redirect: "/judge/submissions",
  meta: {
    icon: "ep/cpu",
    title: "评测管理",
    rank: 3
  },
  children: [
    {
      path: "/judge/submissions",
      name: "AdminSubmissions",
      component: () => import("@/views/judge/submissions/index.vue"),
      meta: { title: "提交记录", permission: "submission.read" }
    },
    {
      path: "/judge/nodes",
      name: "AdminJudgeNodes",
      component: () => import("@/views/judge/nodes/index.vue"),
      meta: { title: "评测机状态", permission: "judge.node.read" }
    },
    {
      path: "/judge/languages",
      name: "AdminLanguages",
      component: () => import("@/views/judge/languages/index.vue"),
      meta: { title: "编程语言", permission: "judge.language.read" }
    },
    {
      path: "/judge/remote-accounts",
      name: "AdminRemoteAccounts",
      component: () => import("@/views/judge/remote-accounts/index.vue"),
      meta: { title: "远程账号", permission: "judge.remote.read" }
    }
  ]
} satisfies RouteConfigsTable;
