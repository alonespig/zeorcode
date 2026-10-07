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
      component: () => import("@/pages/ProblemSet/Detail.vue"),
      children: [
        {
          path: "",
          name: "ProblemSetIntro",
          meta: { title: "题单简介" },
          component: () => import("@/pages/ProblemSet/Intro.vue"),
        },
        {
          path: "problems",
          name: "ProblemSetProblems",
          meta: { title: "题目列表" },
          component: () => import("@/pages/ProblemSet/Problems.vue"),
        },
        {
          path: "rank",
          name: "ProblemSetRank",
          meta: { title: "题单排行榜" },
          component: () => import("@/pages/ProblemSet/Rank.vue"),
        },
      ],
    },
  ],
}
