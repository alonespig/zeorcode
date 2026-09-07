// utils/request.js
import axios from "axios";
import { ElMessage } from "element-plus";
import { useUserStore } from "@/stores/user";

const service = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 5000,
  withCredentials: true,
});

// 后端约定：未登录 / token 过期 / token 非法
const AUTH_FAIL_CODES = [20001, 20002, 20003];
// 业务成功码
const SUCCESS_CODE = 200;

service.interceptors.response.use(
  (response) => {
    if (response.config.responseType === "blob") {
      return response;
    }
    const res = response.data;

    // 鉴权失败：清登录态、弹登录框
    if (AUTH_FAIL_CODES.includes(res.code)) {
      const userStore = useUserStore();
      ElMessage.error(res.msg || "登录已过期");
      userStore.localLogout();
      userStore.promptLogin();
      return Promise.reject(res);
    }

    // 业务成功
    if (res.code === SUCCESS_CODE) {
      return res;
    }

    // 其他业务错误：统一弹 msg，进 catch
    ElMessage.error(res.msg || "请求失败");
    return Promise.reject(res);
  },
  (error) => {
    // 非 2xx / 网络异常
    if (error.response) {
      const resp = error.response.data;
      ElMessage.error(resp?.msg || "网络异常");
    } else {
      ElMessage.error(error.message || "网络异常");
    }
    return Promise.reject(error);
  },
);

export default service;
