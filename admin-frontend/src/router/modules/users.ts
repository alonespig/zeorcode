export default {
  path: "/users",
  name: "AdminUsers",
  component: () => import("@/views/users/index.vue"),
  meta: {
    icon: "ep/user",
    title: "用户管理",
    rank: 1
  }
} satisfies RouteConfigsTable;
