<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router";
import {
  Bell,
  Cloudy,
  Collection,
  DocumentChecked,
  Cpu,
  Monitor,
  PriceTag,
  Setting,
  User,
} from "@element-plus/icons-vue";
import { useUserStore } from "@/stores/user";

const route = useRoute();
const userStore = useUserStore();

const menuItems = computed(() =>
  [
    { path: "/admin/users", label: "用户管理", icon: User, permission: "user.read" },
    { path: "/admin/post-review", label: "帖子审核", icon: DocumentChecked, permission: "post.review" },
    { path: "/admin/problemset", label: "题单管理", icon: Collection, permission: "problemset.read" },
    { path: "/admin/tags", label: "算法标签", icon: PriceTag, permission: "tag.read" },
    { path: "/admin/languages", label: "编程语言", icon: Cpu, permission: "judge.language.read" },
    { path: "/admin/judge", label: "评测机状态", icon: Monitor, permission: "judge.node.read" },
    { path: "/admin/remote", label: "远程账号", icon: Cloudy, permission: "judge.remote.read" },
    { path: "/admin/notice", label: "系统通知", icon: Bell, permission: "notification.send" },
  ].filter(item => userStore.can(item.permission))
);
</script>

<template>
  <div class="admin-page">
    <!-- 左侧导航（浅色，仿 HOJ） -->
    <aside class="admin-side">
      <div class="side-title">
        <el-icon>
          <Setting />
        </el-icon>
        <span>管理后台</span>
      </div>
      <el-menu :default-active="route.path" router class="side-menu">
        <el-menu-item
          v-for="item in menuItems"
          :key="item.path"
          :index="item.path"
        >
          <el-icon>
            <component :is="item.icon" />
          </el-icon>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>
    </aside>

    <!-- 右侧内容 -->
    <main class="admin-content">
      <router-view />
    </main>
  </div>
</template>

<style scoped>
.admin-page {
  /* 后台使用整行可用宽度，表格页可以完整展示管理字段。 */
  box-sizing: border-box;
  width: 100%;
  padding: 0 24px 32px;
  display: flex;
  justify-content: flex-start;
  gap: 16px;
  align-items: flex-start;
}

.admin-side {
  width: 220px;
  flex: 0 0 auto;
  background: #fff;
  border: 1px solid #e6e9ee;
  border-radius: 4px;
  overflow: hidden;
  /* 左侧导航吸顶常驻：top = 顶部导航高度 58px + 其下边距 20px，
     保证滚动时导航栏与侧栏始终可见、中间留 20px 空隙不重叠 */
  position: sticky;
  top: 78px;
  align-self: flex-start;
  max-height: calc(100vh - 78px - 16px);
  overflow-y: auto;
}

.side-title {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px;
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  border-bottom: 1px solid #f0f2f5;
}

.side-menu {
  border-right: none;
}

.admin-content {
  flex: 1;
  min-width: 0;
  width: 100%;
}

/* 后台表格密度：比前台大一档，管理页要长时间盯着看，13/14px 偏小。
   这里直接覆写 common.css 里的表格 token —— element-theme.css(Element Plus 表格)
   和 table.css(原生 .f-table/.data-table) 都消费这几个变量，且自定义属性会跨组件继承，
   所以 5 个后台页面的表格一起生效，不必每页单独调。 */
.admin-page {
  --table-font-size: 15px;
  --table-header-font-size: 14px;
  --table-header-height: 52px;
  --table-cell-padding-y: 14px;
}

/* 表格内的标签、按钮、下拉不吃上面那几个 token，按同一档单独对齐，
   否则会比正文明显小一圈。用 :deep() 是因为它们渲染在子路由组件里。 */
.admin-page :deep(.el-table .el-tag) {
  font-size: 13px;
  height: 24px;
  line-height: 22px;
  padding: 0 9px;
}

.admin-page :deep(.el-table .el-button--small) {
  font-size: 14px;
}

.admin-page :deep(.el-table .el-select) {
  font-size: 14px;
}

.admin-page :deep(.el-table .el-select--small .el-select__wrapper) {
  min-height: 30px;
  font-size: 14px;
}

/* 窄屏：侧栏取消吸顶，改为纵向堆叠在内容之上 */
@media (max-width: 767px) {
  .admin-page {
    flex-direction: column;
    padding: 12px;
  }

  .admin-side {
    position: static;
    width: 100%;
    max-height: none;
  }

  .admin-content {
    /* .admin-page 的 align-items:flex-start 是给桌面横排时纵向对齐用的，
       但纵向堆叠后它会控制横向：子元素默认按内容宽度收缩(shrink-to-fit)而不是撑满，
       内容里的表格一收缩就是 900~1000px，比手机屏幕宽得多，整个页面被横向撑开。
       显式撑满宽度，覆盖掉这个收缩行为 */
    width: 100%;
  }
}
</style>
