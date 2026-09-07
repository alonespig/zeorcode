export const postRoutes ={
  path: 'blog',
  name: 'Blog',
  children: [
    {
      path: '',
      name: `BlogList`,
      component: () => import('@/pages/Post/List.vue'),
    },
    {
      path: 'create',
      name: `BlogCreate`,
      meta: { title: `发布博客`, requiresAuth: true },
      component: () => import('@/pages/Post/Editor.vue'),
    },
    {
      path: ':id',
      name: `BlogDetail`,
      meta: { title: `博客详情`  },
      component: () => import('@/pages/Post/Detail.vue'),
    },
    {
      path: ':id/edit',
      name: `BlogEdit`,
      meta: { title: `编辑博客`,  requiresAuth: true },
      component: () => import('@/pages/Post/Editor.vue'),
    },
  ],
}

