import { createApp } from "vue";
import { createPinia } from "pinia";

import App from "./App.vue";
import router from "./router";
import "@/styles/common.css";
import "@/styles/assets.css"
import "@/styles/table.css"
// 模板中的 Element Plus 组件及样式由 ElementPlusResolver 按需引入。
// ElMessage / ElMessageBox 由业务代码直接调用，需要在入口补充其样式。
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/message-box/style/css'
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
