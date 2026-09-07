export const problemSetRoutes = {
  path: "problemset",
  name: "ProblemSet",
  children: [
    {
      path: "",
      name: "ProblemSetList",
      meta: { title: "题单" },
      component: () => import("@/pages/ProblemSet/index.vue"),
    },
    {
      // 未登录也能看，只是不展示做题进度，所以不加 requiresAuth
      path: ":id",
      name: "ProblemSetDetail",
      meta: { title: "题单详情" },
      component: () => import("@/pages/ProblemSet/Detail.vue"),
    },
  ],
}
