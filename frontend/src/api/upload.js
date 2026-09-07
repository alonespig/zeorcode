import request from '@/utils/request'

export const uploadImage = (data) => {
  return request({
    url: '/upload/image',
    method: 'post',
    data
  })
}

// 注册页尚未登录，使用独立限流的头像上传接口。
export const uploadRegisterAvatar = (data) => {
  return request({
    url: '/upload/register-avatar',
    method: 'post',
    data,
  })
}
