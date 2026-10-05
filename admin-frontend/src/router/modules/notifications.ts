export default {
  path: "/notifications",
  name: "AdminNotifications",
  component: () => import("@/views/notifications/index.vue"),
  meta: {
    icon: "ep/bell",
    title: "系统通知",
    rank: 4
  }
} satisfies RouteConfigsTable;
