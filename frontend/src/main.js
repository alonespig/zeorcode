import { createApp } from "vue";
import { createPinia } from "pinia";

import App from "./App.vue";
import router from "./router";
import "@/styles/common.css";
import "@/styles/assets.css"
import "@/styles/table.css"
// Element Plus 组件由 vite.config.js 的 ElementPlusResolver 按需引入，这里只引全量样式；
// 中文语言包在 App.vue 的 el-config-provider 上设置。
import 'element-plus/dist/index.css'
import '@/styles/element-theme.css'
import '@/font/iconfont.css'
import permission from '@/directives/permission'
import { useUserStore } from "@/stores/user"


const app = createApp(App);

// 全局注册指令
app.directive('permission', permission)

app.use(createPinia());

const userStore = useUserStore()
userStore.loadFromStorage()

// 路由页面发起业务请求前，先用服务端 Cookie 恢复真实登录态。
const bootstrap = async () => {
  await userStore.restoreSession()
  app.use(router)
  app.mount("#app")
}

bootstrap()
