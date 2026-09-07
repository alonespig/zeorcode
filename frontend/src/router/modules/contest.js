export const contestRoutes = {
  path: "contest",
  name: "Contest",
  children: [
    {
      path: "",
      name: "ContestList",
      meta: { title: "比赛" },
      component: () => import("@/pages/Contest/index.vue"),
    },
    {
      path: "create",
      name: "CreateContest",
      meta: { title: "创建比赛", requiresAuth: true },
      component: () => import("@/pages/Contest/CreateContest.vue"),
    },
    {
      path: ":id",
      name: "ContestDetail",
      children: [
        {
          path: "",
          name: "ContestDetailView",
          meta: { title: "比赛详情" },
          component: () => import("@/pages/Contest/ContestDetail.vue"),
          redirect: {
            name: 'ContestDescription',
          },
          children: [
            {
              path: "description",
              name: "ContestDescription",
              meta: { title: "比赛详情" },
              component: () => import("@/pages/Contest/ContestDescription.vue"),
            },
            {
              path: "problem",
              name: "ContestProblem",
              meta: { title: "比赛题目" },
              component: () => import("@/pages/Contest/ContestProblem.vue"),
            },
            {
              path: "submit",
              name: "ContestSubmit",
              meta: { title: "比赛提交" },
              component: () => import("@/pages/Contest/ContestSubmit.vue"),
            },
            {
              path: "submission",
              name: "ContestSubmission",
              meta: { title: "比赛评测" },
              component: () => import("@/pages/Contest/ContestSubmission.vue"),
            },
            {
              path: "rank",
              name: "ContestRank",
              meta: { title: "比赛排名" },
              component: () => import("@/pages/Contest/ContestRank.vue"),
            },
            {
              path: "rejudge",
              name: "ContestRejudge",
              meta: { title: "比赛重判", requiresAuth: true },
              component: () => import("@/pages/Contest/ContestRejudge.vue"),
            },
          ],
        },
      ],
    },
    {
      path: ":id/edit",
      name: "EditContest",
      meta: { title: "编辑比赛", requiresAuth: true },
      component: () => import("@/pages/Contest/EditContest.vue"),
    },
    {
      path: ":id/problem/:problemId",
      name: "ContestProblemDetail",
      meta: { title: "比赛题目详情" },
      component: () => import("@/pages/Contest/ProblemDetail.vue"),
    },
  ],
}
