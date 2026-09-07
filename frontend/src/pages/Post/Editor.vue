<template>
  <div class="page-container">
    <section class="rounded border border-gray-200 bg-white">
      <!-- 顶部操作栏 -->
      <div class="flex items-center justify-between px-5 py-3 border-b border-gray-100">
        <h2 class="m-0 text-base font-semibold text-gray-600">{{ isEdit ? '编辑' : '发布' }}{{ pageTitle }}</h2>
        <div class="flex gap-2.5">
          <el-button @click="handleCancel">取消</el-button>
          <el-button type="primary" :loading="saving" @click="handleSubmit">
            <span class="iconfont icon-baocun mr-1"></span>
            {{ submitText }}
          </el-button>
        </div>
      </div>

      <el-form label-position="top" class="editor-form px-5 py-4">
        <el-form-item label="标题">
          <el-input v-model="form.title" maxlength="120" show-word-limit placeholder="请输入标题" />
        </el-form-item>

        <!-- 从题目页进入：类型在标题里，这里只展示题目名（不可选） -->
        <el-form-item v-if="locked" label="关联题目">
          <span class="text-gray-800">{{ problemName || '—' }}</span>
        </el-form-item>

        <!-- 从全站"发布"进入：类型可选，题解可搜题目 -->
        <el-form-item v-else label="类型">
          <div class="flex items-center gap-4 flex-wrap">
            <el-select v-model="form.category" class="w-40!" placeholder="选择类型">
              <el-option v-for="item in availableCategories" :key="item.id" :value="item.id" :label="item.name" />
            </el-select>

            <template v-if="form.category === 'solution'">
              <el-select v-model="form.problemID" filterable remote clearable :remote-method="remoteSearchProblem"
                :loading="problemLoading" class="w-72!" placeholder="搜索题目名…">
                <el-option v-for="p in problemOptions" :key="p.id" :value="p.id" :label="p.name" />
              </el-select>
              <!-- <span class="text-xs text-gray-400">选「题解」需关联一道题目（按题目名搜，不用记题号）</span> -->
            </template>
          </div>
        </el-form-item>

        <el-form-item label="正文">
          <MdEditor v-model="form.content" height="560px" />
        </el-form-item>
      </el-form>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { categoryList, categoryTitleMap } from '@/constants/index'
import MdEditor from '@/components/MdEditor.vue'
import { createPost, getPost, updatePost } from '@/api/post'
import { getProblemList, getProblem } from '@/api/problems'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 公告只有管理员能发：非管理员的类型选项里去掉"公告"
const availableCategories = computed(() =>
  userStore.isAdmin ? categoryList : categoryList.filter((c) => c.id !== 'announcement')
)

const form = reactive({
  category: route.query.category || route.meta.category || 'blog',
  problemID: route.query.pid || '',
  title: '',
  content: '',
})

const saving = ref(false)

const isEdit = computed(() => Boolean(route.params.id))
const pageTitle = computed(() => categoryTitleMap[form.category] || '文章')
// 新建即发布，编辑才是保存
const submitText = computed(() => (isEdit.value ? '保存' : '发布'))

// 从题目页进入(URL 带 pid)：类型体现在标题(发布题解/发布讨论)，题目只展示名称、不可选
const locked = computed(() => !!route.query.pid)
const problemName = ref('')

// ---- 关联题目：按题目名远程搜索 ----
const problemOptions = ref([])
const problemLoading = ref(false)

const remoteSearchProblem = async (q) => {
  problemLoading.value = true
  try {
    const res = await getProblemList({ page: 1, pageSize: 20, q: q || undefined })
    problemOptions.value = res.data.list || []
  } finally {
    problemLoading.value = false
  }
}

const loadPost = async () => {
  if (!isEdit.value) return
  const res = await getPost(route.params.id)
  const data = res.data
  form.category = data.category
  form.problemID = data.problemID || ''
  form.title = data.title || ''
  form.content = data.content || ''
  // 编辑题解时，回填所选题目的名字（否则下拉只显示 id）
  if (form.category === 'solution' && form.problemID) {
    try {
      const p = await getProblem(form.problemID)
      problemOptions.value = [{ id: p.data.id, name: p.data.name }]
    } catch {
      // 题目可能已删除，忽略
    }
  }
}

const handleCancel = () => {
  router.back()
}

const handleSubmit = async () => {
  if (!form.title.trim()) {
    ElMessage.warning('标题不能为空')
    return
  }
  if (!form.content.trim()) {
    ElMessage.warning('正文不能为空')
    return
  }
  if (form.category === 'solution' && !form.problemID) {
    ElMessage.warning('题解必须关联题目')
    return
  }

  saving.value = true
  try {
    const payload = {
      category: form.category,
      problemID: form.problemID || '',
      title: form.title.trim(),
      content: form.content,
    }
    if (isEdit.value) {
      await updatePost(route.params.id, payload)
      ElMessage.success(userStore.isAdmin ? '已保存' : '已保存，修改后需重新审核')
      router.push({ name: 'BlogDetail', params: { id: route.params.id } })
      return
    }
    const res = await createPost(payload)
    ElMessage.success(res.data?.reviewStatus === 0 ? '已提交，等待管理员审核' : '发布成功')
    router.push({ name: 'BlogDetail', params: { id: res.data.id } })
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  await loadPost()
  // 锁定态(从题目页新建)：拉一次题目名用于展示
  if (locked.value && !isEdit.value && form.problemID) {
    try {
      const p = await getProblem(form.problemID)
      problemName.value = p.data.name
    } catch {
      // 题目可能已删除，忽略
    }
  }
})
</script>

<style scoped lang="scss">
.editor-form {
  :deep(.el-form-item__label) {
    color: #374151;
    font-weight: 600;
  }
}
</style>
