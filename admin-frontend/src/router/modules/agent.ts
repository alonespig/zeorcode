export default {
  path: "/agent",
  name: "AdminAgent",
  component: () => import("@/views/agent/index.vue"),
  meta: {
    icon: "ri/robot-2-line",
    title: "AI 助手",
    rank: 5
  }
} satisfies RouteConfigsTable;
