export const submissionRoutes = {
  path: "submission",
  children: [
    {
      path: "",
      name: "SubmissionList",
      meta: { title: "评测记录" },
      component: () => import("@/pages/Submission/index.vue"),
    },
    {
      path: ":id",
      name: "SubmissionDetail",
      meta: { title: "评测详情" },
      component: () => import("@/pages/Submission/Detail.vue"),
    },
  ],
}
