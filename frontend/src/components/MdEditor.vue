<template>
  <MdPreview
    v-if="previewOnly"
    :editor-id="editorId"
    :model-value="content"
    :preview-theme="previewTheme"
    :code-theme="codeTheme"
    :show-code-row-number="true"
    :md-heading-id="mdHeadingId"
  />
  <MdEditor
    v-else
    v-model="content"
    :editor-id="editorId"
    :height="height"
    :toolbars="activeToolbars"
    :preview-theme="previewTheme"
    :code-theme="codeTheme"
    :show-code-row-number="true"
    :md-heading-id="mdHeadingId"
    :on-upload-img="handleUploadImg"
  />
</template>

<script setup>
import { computed } from 'vue'
import { MdEditor, MdPreview } from 'md-editor-v3'
import { uploadImage } from '@/api/upload'
import { mdHeadingId } from '@/utils/md'
import 'md-editor-v3/lib/style.css'
import 'md-editor-v3/lib/preview.css'

const fullToolbars = [
  'bold',
  'underline',
  'italic',
  'strikeThrough',
  'title',
  'sub',
  'sup',
  'quote',
  'unorderedList',
  'orderedList',
  'task',
  'codeRow',
  'code',
  'link',
  'image',
  'table',
  'mermaid',
  'katex',
  'revoke',
  'next',
  'save',
  'preview',
  'htmlPreview',
  'catalog',
  'fullscreen',
]

const props = defineProps({
  modelValue: {
    type: String,
    default: '',
  },
  height: {
    type: String,
    default: '500px',
  },
  previewOnly: {
    type: Boolean,
    default: false,
  },
  toolbars: {
    type: Array,
    default: null,
  },
  previewTheme: {
    type: String,
    default: 'default',
  },
  codeTheme: {
    type: String,
    default: 'github',
  },
  editorId: {
    type: String,
    default: 'zeorcode-md-editor',
  },
})

const emit = defineEmits(['update:modelValue'])

const content = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const activeToolbars = computed(() => props.toolbars || fullToolbars)

const handleUploadImg = async (files, callback) => {
  const urls = []
  for (const file of files) {
    const formData = new FormData()
    formData.append('file', file)
    const res = await uploadImage(formData)
    urls.push(res.data.url)
  }
  callback(urls)
}
</script>
