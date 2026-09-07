export const problemRoutes = {
  path: "problem",
  name: "Problem",
  children: [
    {
      path: "",
      name: "ProblemList",
      meta: { title: "题库" },
      component: () => import("@/pages/Problem/index.vue"),
    },
    {
      path: "create",
      name: "CreateProblem",
      meta: { title: "创建题目", requiresAuth: true },
      component: () => import("@/pages/Problem/CreateProblem.vue"),
    },
    {
      // 题目详情：布局(常驻侧栏 + 信息 + 统计) + 子路由 tab
      path: ":id",
      component: () => import("@/pages/Problem/ProblemLayout.vue"),
      children: [
        {
          path: "",
          name: "ProblemDetail",
          meta: { title: "题目详情", module: 'ProblemDetail' },
          component: () => import("@/pages/Problem/ProblemDetail.vue"),
        },
        {
          path: "solutions",
          name: "ProblemSolutions",
          meta: { title: "题解", module: 'ProblemSolutions' },
          component: () => import("@/pages/Problem/tabs/Solutions.vue"),
        },
        {
          path: "discuss",
          name: "ProblemDiscuss",
          meta: { title: "讨论", module: 'ProblemDiscuss'  },
          component: () => import("@/pages/Problem/tabs/Discuss.vue"),
        },
        {
          path: "submissions",
          name: "ProblemSubmissions",
          meta: { title: "提交记录", module: 'ProblemSubmissions'  },
          component: () => import("@/pages/Problem/tabs/Submissions.vue"),
        },
      ],
    },
    {
      path: ":id/edit",
      name: "EditProblem",
      meta: { title: "编辑题目", requiresAuth: true },
      component: () => import("@/pages/Problem/EditProblem.vue"),
    },
    {
      path: ":id/submit",
      name: "ProblemSubmit",
      meta: { title: "提交代码", requiresAuth: true },
      component: () => import("@/pages/Problem/Submit.vue"),
    },
    {
      path: ":id/file",
      name: "ProblemFile",
      meta: { title: "测试数据", requiresAuth: true },
      component: () => import("@/pages/Problem/TestDataPanel.vue"),
    },
  ],
}
