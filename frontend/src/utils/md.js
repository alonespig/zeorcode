// md-editor-v3 的 MdPreview 与 MdCatalog 默认 mdHeadingId 不一致
// （preview 默认 () => ""，catalog 默认 () => undefined），
// 导致目录点击时 getElementById 找不到标题、无法跳转。
// 这里提供一个统一的生成函数，正文和目录都用它，id 才能对上。
// preview 与 catalog 传入的 index 都是 1 开始，可直接用。
export const mdHeadingId = ({ index }) => `heading-${index}`
