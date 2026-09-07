export const teamRoutes = {
  path: "team",
  name: "Team",
  children: [
    {
      path: "",
      name: "TeamList",
      meta: { title: "团队" },
      component: () => import("@/pages/Team/index.vue"),
    },
    {
      // 团队布局：横幅 + tab（概览/作业/成员），子路由注入 team 数据
      path: ":id",
      component: () => import("@/pages/Team/TeamLayout.vue"),
      children: [
        {
          path: "",
          name: "TeamOverview",
          meta: { title: "团队概览", module: "TeamOverview" },
          component: () => import("@/pages/Team/tabs/Overview.vue"),
        },
        {
          path: "homework",
          name: "TeamHomework",
          meta: { title: "团队作业", module: "TeamHomework" },
          component: () => import("@/pages/Team/tabs/HomeworkList.vue"),
        },
        {
          path: "member",
          name: "TeamMembers",
          meta: { title: "团队成员", module: "TeamMembers" },
          component: () => import("@/pages/Team/tabs/Members.vue"),
        },
      ],
    },
    {
      // 作业详情：独立路由（不套团队横幅），内部四个 tab 用子路由
      path: ":id/homework/:hid",
      component: () => import("@/pages/Team/homework/HomeworkLayout.vue"),
      children: [
        {
          path: "",
          name: "HomeworkIntro",
          meta: { title: "作业简介", module: "HomeworkIntro" },
          component: () => import("@/pages/Team/homework/Intro.vue"),
        },
        {
          path: "problems",
          name: "HomeworkProblems",
          meta: { title: "题目列表", module: "HomeworkProblems" },
          component: () => import("@/pages/Team/homework/Problems.vue"),
        },
        {
          path: "rank",
          name: "HomeworkRank",
          meta: { title: "排行榜", module: "HomeworkRank" },
          component: () => import("@/pages/Team/homework/Rank.vue"),
        },
        {
          path: "submissions",
          name: "HomeworkSubmissions",
          meta: { title: "提交列表", module: "HomeworkSubmissions" },
          component: () => import("@/pages/Team/homework/Submissions.vue"),
        },
      ],
    },
    {
      // 作业做题页：保持作业上下文，不跳到公共题库详情
      path: ":id/homework/:hid/problem/:problemId",
      name: "HomeworkProblemDetail",
      meta: { title: "作业题目详情", requiresAuth: true },
      component: () => import("@/pages/Team/homework/ProblemDetail.vue"),
    },
  ],
}
