import request from '@/utils/request'

export const login = (data) => {
  return request({
    url: '/login',
    method: 'post',
    data
  })
}

export const getLoginCaptcha = () => {
  return request({
    url: '/captcha',
    method: 'get',
  })
}

// 注册：后端 POST /api/user（用户名 + 密码 + 邮箱 + 验证码）
export const register = (data) => {
  return request({
    url: '/user',
    method: 'post',
    data
  })
}

// 发送邮箱验证码：scene = register | reset | bind
export const sendVerifyCode = (data) => {
  return request({
    url: '/verify-code',
    method: 'post',
    data
  })
}

// 通过邮箱验证码重置密码
export const resetPassword = (data) => {
  return request({
    url: '/reset-password',
    method: 'post',
    data
  })
}

// 登录用户校验当前密码后修改密码；成功后服务端会退出当前登录态。
export const changePassword = (data) => {
  return request({
    url: '/user/password',
    method: 'put',
    data,
  })
}

// 换绑/绑定邮箱（需登录）
export const bindEmail = (data) => {
  return request({
    url: '/user/email',
    method: 'put',
    data
  })
}

export const submit = (data) => {
  return request({
    url: '/submission',
    method: 'post',
    data
  })
}

export const getSubmit = (params) => {
  return request({
    url: '/submission',
    method: 'get',
    params
  })
}

export const getSubmitDetail = (id) => {
  return request({
    url: `/submission/${id}`,
    method: 'get',
  })
}

// 单条重判（超管）
export const rejudgeSubmission = (id) => {
  return request({
    url: `/admin/submission/${id}/rejudge`,
    method: 'post',
  })
}

export const userRank = (params) => {
  return request({
    url: '/user/rank',
    method: 'get',
    params
  })
}

// 按 rating 排名
export const ratingRank = (params) => {
  return request({
    url: '/user/rating-rank',
    method: 'get',
    params
  })
}

export const userInfo = () => {
  return request({
    url: '/user/info',
    method: 'get',
  })
}

// 更新当前登录用户的资料（签名/性别/邮箱/头像）
export const updateUserInfo = (data) => {
  return request({
    url: '/user/info',
    method: 'put',
    data,
  })
}

export const getUserRecent7DaysAc = (id) => {
  return request({
    url: `/user/${id}/ac-stats`,
    method: 'get',
  })
}


// 某用户 rating + 历次变化（个人页折线图）
export const getUserRating = (id) => {
  return request({
    url: `/user/${id}/rating`,
    method: 'get',
  })
}

// 某用户参赛记录（个人页「比赛记录」：每场 rating 变化/计算中/不计分）
export const getUserContestHistory = (id) => {
  return request({
    url: `/user/${id}/contest-history`,
    method: 'get',
  })
}

export const getUserProfile = (id) => {
  return request({
    url: `/user/${id}/profile`,
    method: 'get',
  })
}
