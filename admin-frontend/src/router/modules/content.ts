const Layout = () => import("@/layout/index.vue");

export default {
  path: "/content",
  name: "AdminContent",
  component: Layout,
  redirect: "/content/problems",
  meta: {
    icon: "ri/file-list-3-line",
    title: "内容管理",
    rank: 2
  },
  children: [
    {
      path: "/content/problems",
      name: "AdminProblems",
      component: () => import("@/views/content/problems/index.vue"),
      meta: { title: "题目管理" }
    },
    {
      path: "/content/problemsets",
      name: "AdminProblemSets",
      component: () => import("@/views/content/problemsets/index.vue"),
      meta: { title: "题单管理" }
    },
    {
      path: "/content/problemsets/create",
      name: "AdminProblemSetsCreate",
      component: () => import("@/views/content/problemsets/create.vue"),
      meta: {
        title: "新建题单",
        showLink: false,
        activePath: "/content/problemsets"
      }
    },
    {
      path: "/content/problemsets/:id/edit",
      name: "AdminProblemSetsEdit",
      component: () => import("@/views/content/problemsets/edit.vue"),
      meta: {
        title: "编辑题单",
        showLink: false,
        activePath: "/content/problemsets"
      }
    },
    {
      path: "/content/contests",
      name: "AdminContests",
      component: () => import("@/views/content/contests/index.vue"),
      meta: { title: "比赛管理" }
    },
    {
      path: "/content/post-review",
      name: "AdminPostReview",
      component: () => import("@/views/content/post-review/index.vue"),
      meta: { title: "帖子审核" }
    },
    {
      path: "/content/tags",
      name: "AdminTags",
      component: () => import("@/views/content/tags/index.vue"),
      meta: { title: "算法标签" }
    }
  ]
} satisfies RouteConfigsTable;
