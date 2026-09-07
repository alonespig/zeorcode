import request from '@/utils/request'

export const getAdminUserList = (params) => {
  return request({
    url: '/admin/users',
    method: 'get',
    params,
  })
}

export const importAdminUsers = (file) => {
  const data = new FormData()
  data.append('file', file)
  return request({
    url: '/admin/users/import',
    method: 'post',
    data,
    timeout: 120000,
  })
}

// 封禁/解封用户（status: 0 正常 / 1 封禁）
export const setUserStatus = (id, status) => {
  return request({
    url: `/admin/users/${id}/status`,
    method: 'put',
    data: { status },
  })
}

// 修改用户角色（role: 0 普通用户 / 1 管理员）；后端会撤销该用户会话，需其重新登录
export const setUserRole = (id, role) => {
  return request({
    url: `/admin/users/${id}/role`,
    method: 'put',
    data: { role },
  })
}

// ===== 算法标签 =====
export const getAdminTags = () => {
  return request({ url: '/admin/tags', method: 'get' })
}

export const createAdminTag = (data) => {
  return request({ url: '/admin/tags', method: 'post', data })
}

export const updateAdminTag = (id, data) => {
  return request({ url: `/admin/tags/${id}`, method: 'put', data })
}

export const deleteAdminTag = (id) => {
  return request({ url: `/admin/tags/${id}`, method: 'delete' })
}

// ===== 编程语言 =====
export const getAdminLanguages = () => {
  return request({ url: '/admin/languages', method: 'get' })
}

export const createAdminLanguage = (data) => {
  return request({ url: '/admin/languages', method: 'post', data })
}

export const updateAdminLanguage = (id, data) => {
  return request({ url: `/admin/languages/${id}`, method: 'put', data })
}

export const deleteAdminLanguage = (id) => {
  return request({ url: `/admin/languages/${id}`, method: 'delete' })
}

// 评测机健康状态（现场探活各实例）
export const getJudgeStatus = () => {
  return request({ url: '/admin/judge/status', method: 'get' })
}

// ===== 远程账号(各 OJ 提交账号/Cookie/Token) =====
export const getRemoteAccounts = () => {
  return request({ url: '/admin/remote-account', method: 'get' })
}

export const createRemoteAccount = (data) => {
  return request({ url: '/admin/remote-account', method: 'post', data })
}

export const updateRemoteAccount = (id, data) => {
  return request({ url: `/admin/remote-account/${id}`, method: 'put', data })
}

export const deleteRemoteAccount = (id) => {
  return request({ url: `/admin/remote-account/${id}`, method: 'delete' })
}
