<template>
  <div class="layout">
    <div class="nav">
      <LayoutNav />
    </div>
    <div class="content">
      <router-view />
    </div>
    <footer class="footer">
      <div class="footer-links">
        <a href="https://github.com" target="_blank" rel="noopener">GitHub</a>
        <span class="dot">·</span>
        <RouterLink to="/blog">社区</RouterLink>
        <span class="dot">·</span>
        <RouterLink to="/rank">排名</RouterLink>
      </div>
      <div class="footer-copy">&copy; 2026 ZeorCode · Go + Vue 构建的在线评测系统</div>
    </footer>
  </div>
</template>

<script setup>
import LayoutNav from "@/pages/Layout/components/LayoutNav.vue";
</script>

<style scoped lang="scss">
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;

  .nav {
    width: 100%;
    margin-bottom: 20px;
    position: sticky;
    top: 0;
    z-index: 100;
  }

  .content {
    flex: 1;
  }

  .footer {
    text-align: center;
    padding: 22px 0;
    border-top: 1px solid #eef0f3;

    .footer-links {
      display: flex;
      justify-content: center;
      align-items: center;
      gap: 10px;
      font-size: 13px;

      a {
        color: #606266;
        transition: color 0.2s;

        &:hover {
          color: #007bff;
        }
      }

      .dot {
        color: #c0c4cc;
      }
    }

    .footer-copy {
      margin-top: 8px;
      font-size: 12px;
      color: #a8abb2;
    }
  }
}

@media (max-width: 767px) {
  .layout {
    .nav {
      margin-bottom: 12px;
    }

    .footer {
      padding: 18px 12px;
    }
  }
}
</style>
