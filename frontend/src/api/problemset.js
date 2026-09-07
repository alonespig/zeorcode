import request from '@/utils/request'

// ===== 前台 =====

// 题单列表（只含已发布）。params: { page, pageSize, q, tags, visibility }
export const getProblemSetList = (params) => {
  return request({ url: '/problemset', method: 'get', params })
}

// 题单详情。邀请码题单未解锁时 data.locked 为 true 且 problems 为空
export const getProblemSetDetail = (id) => {
  return request({ url: `/problemset/${id}`, method: 'get' })
}

// 提交邀请码解锁，成功后长期有效
export const unlockProblemSet = (id, inviteCode) => {
  return request({ url: `/problemset/${id}/unlock`, method: 'post', data: { inviteCode } })
}

// ===== 后台（管理员） =====

// 后台题单列表，含草稿
export const getAdminProblemSetList = (params) => {
  return request({ url: '/admin/problemset', method: 'get', params })
}

export const createProblemSet = (data) => {
  return request({ url: '/admin/problemset', method: 'post', data })
}

export const updateProblemSet = (id, data) => {
  return request({ url: `/admin/problemset/${id}`, method: 'put', data })
}

export const deleteProblemSet = (id) => {
  return request({ url: `/admin/problemset/${id}`, method: 'delete' })
}
