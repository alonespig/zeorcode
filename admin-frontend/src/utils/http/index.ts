import Axios, {
  type AxiosInstance,
  type AxiosRequestConfig,
  type CustomParamsSerializer,
  type InternalAxiosRequestConfig
} from "axios";
import { stringify } from "qs";
import { message } from "@/utils/message";
import { notifyAuthExpired } from "@/utils/session-events";
import type {
  PureHttpError,
  RequestMethods,
  PureHttpResponse,
  PureHttpRequestConfig
} from "./types.d";

const SUCCESS_CODE = 200;
const AUTH_FAILURE_CODES = new Set([20001, 20002, 20003]);

/** blob 响应统一处理：Content-Type 为 application/json 时按业务错误解析，正常文件原样返回 */
async function handleBlobResponse(
  response: PureHttpResponse
): Promise<PureHttpResponse> {
  const data = response.data as Blob;
  if (!data || !data.type || !data.type.includes("application/json")) {
    return response;
  }

  let result: { code?: number; msg?: string } = {};
  try {
    result = JSON.parse(await data.text()) as { code?: number; msg?: string };
  } catch {
    // 无法解析为 JSON 时按通用错误处理
  }

  if (result.code !== undefined && AUTH_FAILURE_CODES.has(result.code)) {
    notifyAuthExpired();
  }
  message(result.msg || "请求失败", { type: "error" });
  return Promise.reject(result);
}

const defaultConfig: AxiosRequestConfig = {
  baseURL: import.meta.env.VITE_API_BASE_URL || "/api",
  timeout: 10000,
  withCredentials: true,
  headers: {
    Accept: "application/json, text/plain, */*",
    "X-Requested-With": "XMLHttpRequest"
  },
  paramsSerializer: {
    serialize: stringify as unknown as CustomParamsSerializer
  }
};

class PureHttp {
  private static initConfig: PureHttpRequestConfig = {};
  private static axiosInstance: AxiosInstance = Axios.create(defaultConfig);

  constructor() {
    this.installRequestInterceptor();
    this.installResponseInterceptor();
  }

  private installRequestInterceptor(): void {
    PureHttp.axiosInstance.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const requestConfig = config as InternalAxiosRequestConfig &
          PureHttpRequestConfig;
        if (typeof requestConfig.beforeRequestCallback === "function") {
          requestConfig.beforeRequestCallback(requestConfig);
          return config;
        }
        PureHttp.initConfig.beforeRequestCallback?.(requestConfig);
        return config;
      },
      error => Promise.reject(error)
    );
  }

  private installResponseInterceptor(): void {
    PureHttp.axiosInstance.interceptors.response.use(
      (response: PureHttpResponse) => {
        const config = response.config;
        if (typeof config.beforeResponseCallback === "function") {
          config.beforeResponseCallback(response);
          return response.data;
        }
        if (PureHttp.initConfig.beforeResponseCallback) {
          PureHttp.initConfig.beforeResponseCallback(response);
          return response.data;
        }
        if (config.responseType === "blob") {
          return handleBlobResponse(response);
        }

        const result = response.data as {
          code?: number;
          msg?: string;
        };
        if (result?.code === SUCCESS_CODE) return result;

        if (result && AUTH_FAILURE_CODES.has(result.code)) {
          notifyAuthExpired();
        }
        message(result?.msg || "请求失败", { type: "error" });
        return Promise.reject(result);
      },
      (error: PureHttpError) => {
        error.isCancelRequest = Axios.isCancel(error);
        if (!error.isCancelRequest) {
          const responseData = error.response?.data as
            | { msg?: string }
            | undefined;
          message(responseData?.msg || error.message || "网络异常", {
            type: "error"
          });
        }
        return Promise.reject(error);
      }
    );
  }

  public request<T>(
    method: RequestMethods,
    url: string,
    param?: AxiosRequestConfig,
    axiosConfig?: PureHttpRequestConfig
  ): Promise<T> {
    return PureHttp.axiosInstance.request({
      method,
      url,
      ...param,
      ...axiosConfig
    }) as Promise<T>;
  }

  public post<T, P>(
    url: string,
    params?: AxiosRequestConfig<P>,
    config?: PureHttpRequestConfig
  ): Promise<T> {
    return this.request<T>("post", url, params, config);
  }

  public get<T, P>(
    url: string,
    params?: AxiosRequestConfig<P>,
    config?: PureHttpRequestConfig
  ): Promise<T> {
    return this.request<T>("get", url, params, config);
  }
}

export const http = new PureHttp();
