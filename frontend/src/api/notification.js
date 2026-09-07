import request from '@/utils/request'

// 我的通知列表；type 可为空(全部)或逗号分隔(如 comment,reply)
export const getNotifications = (params) =>
  request({ url: '/notifications', method: 'get', params })

// 各类型未读数 {comment,reply,like,system,rating,total}
export const getUnreadCount = () =>
  request({ url: '/notifications/unread-count', method: 'get' })

// 标已读：{id} 单条 / {type} 该类型 / 空 全部
export const markNotificationRead = (data) =>
  request({ url: '/notifications/read', method: 'post', data })

// 管理员广播系统通知
export const broadcastNotification = (data) =>
  request({ url: '/admin/notifications', method: 'post', data })
