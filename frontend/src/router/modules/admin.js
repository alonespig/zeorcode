export const adminRoutes = {
  path: "admin",
  meta: { requiresAdmin: true },
  redirect: "/admin/users",
  component: () => import("@/pages/Admin/Layout.vue"),
  children: [
    {
      path: "users",
      name: "AdminUsers",
      meta: { title: "用户管理", permission: "user.read" },
      component: () => import("@/pages/Admin/Users.vue"),
    },
    {
      path: "post-review",
      name: "AdminPostReview",
      meta: { title: "帖子审核", permission: "post.review" },
      component: () => import("@/pages/Admin/PostReview.vue"),
    },
    {
      path: "problemset",
      name: "AdminProblemSet",
      meta: { title: "题单管理", permission: "problemset.read" },
      component: () => import("@/pages/Admin/ProblemSet.vue"),
    },
    {
      path: "tags",
      name: "AdminTags",
      meta: { title: "算法标签", permission: "tag.read" },
      component: () => import("@/pages/Admin/Tags.vue"),
    },
    {
      path: "languages",
      name: "AdminLanguages",
      meta: { title: "编程语言", permission: "judge.language.read" },
      component: () => import("@/pages/Admin/Languages.vue"),
    },
    {
      path: "judge",
      name: "AdminJudge",
      meta: { title: "评测机状态", permission: "judge.node.read" },
      component: () => import("@/pages/Admin/JudgeStatus.vue"),
    },
    {
      path: "remote",
      name: "AdminRemote",
      meta: { title: "远程账号", permission: "judge.remote.read" },
      component: () => import("@/pages/Admin/RemoteAccount.vue"),
    },
    {
      path: "notice",
      name: "AdminNotice",
      meta: { title: "系统通知", permission: "notification.send" },
      component: () => import("@/pages/Admin/Notice.vue"),
    },
  ],
}
