import request from '@/utils/request'

// 支持的远程 OJ 列表
export const getRemoteOJList = () => {
  return request({
    url: '/remote/oj',
    method: 'get',
  })
}

// 拉取远程题目  /remote/problem?oj=HDU&pid=1000
export const getRemoteProblem = (params) => {
  return request({
    url: '/remote/problem',
    method: 'get',
    params,
  })
}
