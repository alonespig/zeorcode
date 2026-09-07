import request from '@/utils/request'

export const getLanguages = () => {
  return request({
    url: '/languages',
    method: 'get',
  })
}
