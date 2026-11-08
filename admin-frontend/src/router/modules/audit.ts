export default {
  path: "/audit",
  name: "AdminAuditLogs",
  component: () => import("@/views/audit/index.vue"),
  meta: {
    icon: "ep/document",
    title: "操作日志",
    permission: "audit.read",
    rank: 5
  }
} satisfies RouteConfigsTable;
