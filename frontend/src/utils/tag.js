// 算法标签的统一配色与字号。题库、题单等多处都要渲染标签，
// 集中在这里，避免各页面各写一份导致视觉不一致。
//
// 后端 TagItem 只有 id/name、没有颜色字段，所以按 id 循环分配柔和色板（浅底深字）。
// 字号不能直接写 font-size：el-tag 用 font-size: var(--el-tag-font-size)，
// 且该规则在 CSS 里排在 Tailwind 工具类之后，会盖掉 text-[Npx]，只能覆盖这个变量。
export const TAG_FONT_SIZE = '14px'

export const TAG_PALETTE = [
  { bg: '#ecf5ff', border: '#d9ecff', text: '#409EFF' }, // 蓝
  { bg: '#f0f9eb', border: '#e1f3d8', text: '#67C23A' }, // 绿
  { bg: '#fdf6ec', border: '#faecd8', text: '#E6A23C' }, // 橙
]

// tagStyle 给 el-tag 生成内联 CSS 变量，配合 effect="light" 使用。
export const tagStyle = (tag) => {
  const c = TAG_PALETTE[(Number(tag.id) || 0) % TAG_PALETTE.length]
  return {
    '--el-tag-font-size': TAG_FONT_SIZE,
    '--el-tag-bg-color': c.bg,
    '--el-tag-border-color': c.border,
    '--el-tag-text-color': c.text,
  }
}
