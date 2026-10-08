const Layout = () => import("@/layout/index.vue");

export default {
  path: "/",
  name: "Home",
  component: Layout,
  redirect: "/dashboard",
  meta: {
    icon: "ep/odometer",
    title: "工作台",
    rank: 0
  },
  children: [
    {
      path: "/dashboard",
      name: "AdminDashboard",
      component: () => import("@/views/dashboard/index.vue"),
      meta: {
        title: "工作台"
      }
    }
  ]
} satisfies RouteConfigsTable;
