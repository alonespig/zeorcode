<template>
  <div class="create-contest-page">
    <div class="page-header">
      <div class="title">
        <el-button :icon="ArrowLeft" text @click="goBack" />
        <span>编辑比赛</span>
      </div>
    </div>

    <div class="layout">
      <!-- 左：主区 -->
      <el-form ref="formRef" :model="contestForm" :rules="rules" label-width="100px" class="main">
        <!-- 基本信息 -->
        <div class="block">
          <div class="block-title">基本信息</div>
          <el-form-item label="比赛名称" prop="name">
            <el-input v-model="contestForm.name" clearable maxlength="60" :clear-icon="CloseBold"
              placeholder="请输入比赛名称" />
          </el-form-item>
          <el-form-item label="比赛封面">
            <SquareImageCropper
              v-model="contestForm.coverUrl"
              title="裁剪比赛封面"
              empty-text="选择封面"
              change-text="重新裁剪"
              help-text="可选，PNG/JPG，最大 5MB；不上传则使用默认比赛封面"
              remove-text="移除封面"
              success-message="比赛封面已上传"
              file-prefix="contest-cover"
              preview-alt="比赛封面预览"
            />
          </el-form-item>
          <div class="row">
            <el-form-item label="比赛日期" prop="contestDate">
              <el-date-picker v-model="contestForm.contestDate" :disabled-date="disabledDate" type="date"
                style="width: 200px" placeholder="选择日期" value-format="YYYY-MM-DD" />
            </el-form-item>
            <el-form-item label="开始时间" prop="contestTime">
              <el-time-select v-model="contestForm.contestTime" start="00:00" step="00:15" end="23:45"
                style="width: 160px" placeholder="选择时间" />
            </el-form-item>
          </div>
          <el-form-item label="比赛时长" prop="duration">
            <el-input-number v-model="contestForm.duration" :min="1" :max="600" :step="1" />
            <span class="unit">分钟</span>
          </el-form-item>
          <el-form-item label="比赛赛制" prop="ruleType">
            <el-radio-group v-model="contestForm.ruleType" :disabled="started">
              <el-radio v-for="o in RULE_OPTIONS" :key="o.value" :value="o.value">{{ o.label }}</el-radio>
            </el-radio-group>
            <span v-if="started" class="unit">比赛已开始，赛制不可修改</span>
          </el-form-item>
          <el-form-item label="计入 Rating">
            <el-switch v-model="contestForm.rated" />
            <span class="unit">开启后，比赛结束将按名次结算 rating</span>
          </el-form-item>
        </div>

        <!-- 题目列表 -->
        <div class="block">
          <div class="block-header">
            <span class="block-title">题目列表</span>
            <div class="block-tools">
              <el-input v-model="queryPid" placeholder="输入题目ID" style="width: 150px" :clear-icon="CloseBold" clearable
                @keyup.enter="addProblem" />
              <el-button :icon="Plus" @click="addProblem">添加</el-button>
              <el-button type="primary" :icon="Search" @click="openPicker">从题库选择</el-button>
            </div>
          </div>

          <el-empty v-if="problemList.length === 0" description="还没有添加题目，可输入 ID 或从题库选择" :image-size="70" />
          <table v-else class="problem-table">
            <colgroup>
              <col style="width: 44px">
              <col style="width: 48px">
              <col>
              <col style="width: 100px">
              <col style="width: 100px">
              <col style="width: 180px">
              <col style="width: 64px">
              <col v-if="isScored" style="width: 90px">
              <col style="width: 60px">
            </colgroup>
            <thead>
              <tr>
                <th></th>
                <th class="center">#</th>
                <th class="left">题目名称</th>
                <th class="center">时间限制</th>
                <th class="center">内存限制</th>
                <th class="left">标签</th>
                <th class="center">气球</th>
                <th v-if="isScored" class="center">分数</th>
                <th class="center">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(problem, index) in problemList" :key="problem.id" draggable="true"
                :class="{ dragging: dragIndex === index, 'drag-over': dragOverIndex === index }"
                @dragstart="onDragStart(index)" @dragover.prevent="onDragOver(index)" @dragend="onDragEnd"
                @drop="onDrop(index)">
                <td class="center handle" title="拖拽排序">
                  <el-icon>
                    <Rank />
                  </el-icon>
                </td>
                <td class="center label">{{ labelOf(index) }}</td>
                <td class="left">{{ problem.name }}</td>
                <td class="center">{{ problem.timeLimitMs ? problem.timeLimitMs + 'ms' : '-' }}</td>
                <td class="center">{{ problem.memoryLimitMb ? problem.memoryLimitMb + 'MB' : '-' }}</td>
                <td class="left">
                  <el-tag v-for="tag in problem.tags || []" :key="tag.id" size="small" type="primary"
                    style="margin-right: 4px">{{ tag.name }}</el-tag>
                </td>
                <td class="center">
                  <el-color-picker v-model="problem.balloonColor" :predefine="BALLOON_COLORS" size="small" />
                </td>
                <td v-if="isScored" class="center">
                  <el-input-number v-model="problem.score" :min="1" :max="1000" :controls="false" size="small"
                    style="width: 72px" />
                </td>
                <td class="center">
                  <el-button type="danger" size="small" text :icon="Delete" @click="removeProblem(index)" />
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 比赛说明 -->
        <div class="block">
          <div class="block-title">比赛说明</div>
          <MdEditor v-model="contestForm.description" height="220px" :toolbars="[
            'bold',
            'underline',
            'italic',
            '-',
            'title',
            'link',
            'code',
            'preview'
          ]" />
        </div>
      </el-form>

      <!-- 右：sticky 概览 -->
      <aside class="aside">
        <div class="overview-card">
          <div class="side-title">概览</div>
          <div class="ov-row"><span>比赛名称</span><b class="ellipsis">{{ contestForm.name || '未填写' }}</b></div>
          <div class="ov-row"><span>赛制</span><b>{{ ruleLabel }}</b></div>
          <div class="ov-row"><span>开始时间</span><b>{{ startLabel }}</b></div>
          <div class="ov-row"><span>时长</span><b>{{ contestForm.duration }} 分钟</b></div>
          <div class="ov-row"><span>题目</span><b :class="{ warn: problemList.length === 0 }">{{ problemList.length }}
              道</b></div>
          <el-button type="primary" class="full" :loading="submitting" @click="submitForm">保存修改</el-button>
        </div>
      </aside>
    </div>

    <!-- 从题库选择 -->
    <el-dialog v-model="pickerVisible" title="从题库选择" width="720px" top="8vh">
      <div class="picker-search">
        <el-input v-model="pickerQ" placeholder="搜索题目名称" clearable style="width: 260px"
          @keyup.enter="loadPicker(1)" />
        <el-button type="primary" :icon="Search" @click="loadPicker(1)">搜索</el-button>
      </div>
      <table class="f-table picker-table">
        <colgroup>
          <col style="width: 52px">
          <col style="width: 64px">
          <col>
          <col style="width: 240px">
        </colgroup>
        <thead>
          <tr>
            <th></th>
            <th class="center">#</th>
            <th class="left">题目名称</th>
            <th class="left">标签</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in pickerList" :key="item.id" class="pick-row" @click="togglePick(item.id)">
            <td class="center">
              <el-checkbox :model-value="isPicked(item.id)" :disabled="inList(item.id)" @click.stop="togglePick(item.id)" />
            </td>
            <td class="center">{{ item.id }}</td>
            <td class="left">{{ item.name }}</td>
            <td class="left">
              <el-tag v-for="t in item.tags || []" :key="t.id" size="small" type="primary" style="margin-right: 4px">{{
                t.name }}</el-tag>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="picker-foot">
        <el-pagination size="small" layout="prev, pager, next" :page-size="pickerPageSize" :total="pickerTotal"
          :current-page="pickerPage" @current-change="loadPicker" />
      </div>
      <template #footer>
        <span class="picked-count">已选 {{ picked.length }} 道</span>
        <el-button @click="pickerVisible = false">取消</el-button>
        <el-button type="primary" :loading="picking" @click="confirmPick">确定添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, CloseBold, Search, Plus, Delete, Rank } from '@element-plus/icons-vue'
import { getProblem, getProblemList } from '@/api/problems'
import { getContestInfo, updateContest } from '@/api/contest'
import { ElMessage } from 'element-plus'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { BALLOON_COLORS, CONTEST_TYPE, RULE_OPTIONS, ruleLabelOf } from '@/constants/index'
import { contestProblemLabel } from '@/utils/contest'
import SquareImageCropper from '@/components/upload/SquareImageCropper.vue'

const route = useRoute()
const router = useRouter()
const formRef = ref(null)

const contestForm = ref({
  name: '',
  coverUrl: '',
  contestDate: null,
  contestTime: null,
  duration: 120,
  ruleType: 1,
  rated: false,
  description: '',
})
const problemList = ref([])
const queryPid = ref(null)
const submitting = ref(false)
const started = ref(false) // 比赛是否已开始（开始后锁定赛制）

const pickColor = () => BALLOON_COLORS[problemList.value.length % BALLOON_COLORS.length]

const goBack = () => router.back()

// ---------- 概览 ----------
const ruleLabel = computed(() => ruleLabelOf(contestForm.value.ruleType))
const isScored = computed(() => contestForm.value.ruleType !== CONTEST_TYPE.ACM)
const startLabel = computed(() => {
  const { contestDate, contestTime } = contestForm.value
  return contestDate && contestTime ? `${contestDate} ${contestTime}` : '未设置'
})

const labelOf = contestProblemLabel

// ---------- 题目对象 ----------
const toProblem = (d) => ({
  id: d.id,
  name: d.name,
  timeLimitMs: d.timeLimitMs,
  memoryLimitMb: d.memoryLimitMb,
  tags: d.tags || [],
  score: 100,
})

const addProblem = async () => {
  const id = (queryPid.value || '').trim()
  if (!id) {
    ElMessage.error('请输入题目编号')
    return
  }
  if (problemList.value.some((item) => String(item.id) === id)) {
    ElMessage.error('题目已添加')
    return
  }
  try {
    const res = await getProblem(id)
    const p = toProblem(res.data)
    p.balloonColor = pickColor()
    problemList.value.push(p)
    queryPid.value = null
  } catch (err) {
    console.error(err)
  }
}

const removeProblem = (index) => {
  problemList.value.splice(index, 1)
}

// ---------- 拖拽排序 ----------
const dragIndex = ref(-1)
const dragOverIndex = ref(-1)
const onDragStart = (i) => { dragIndex.value = i }
const onDragOver = (i) => { dragOverIndex.value = i }
const onDragEnd = () => { dragIndex.value = -1; dragOverIndex.value = -1 }
const onDrop = (i) => {
  const from = dragIndex.value
  if (from !== -1 && from !== i) {
    const list = problemList.value
    const [moved] = list.splice(from, 1)
    list.splice(i, 0, moved)
  }
  onDragEnd()
}

// ---------- 从题库选择 ----------
const pickerVisible = ref(false)
const picking = ref(false)
const pickerQ = ref('')
const pickerList = ref([])
const pickerPage = ref(1)
const pickerPageSize = 8
const pickerTotal = ref(0)
const picked = ref([])

const inList = (id) => problemList.value.some((p) => p.id === id)
const isPicked = (id) => picked.value.includes(id)
const togglePick = (id) => {
  if (inList(id)) return
  const idx = picked.value.indexOf(id)
  if (idx === -1) picked.value.push(id)
  else picked.value.splice(idx, 1)
}

const loadPicker = async (page = 1) => {
  pickerPage.value = page
  try {
    const res = await getProblemList({
      page,
      pageSize: pickerPageSize,
      q: pickerQ.value || undefined,
    })
    pickerList.value = res.data.list || []
    pickerTotal.value = res.data.total || 0
  } catch (err) {
    console.error(err)
  }
}

const openPicker = () => {
  picked.value = problemList.value.map((p) => p.id)
  pickerQ.value = ''
  pickerVisible.value = true
  loadPicker(1)
}

const confirmPick = async () => {
  const toAdd = picked.value.filter((id) => !inList(id))
  if (toAdd.length === 0) {
    pickerVisible.value = false
    return
  }
  picking.value = true
  try {
    const results = await Promise.all(toAdd.map((id) => getProblem(id)))
    results.forEach((res) => {
      const p = toProblem(res.data)
      p.balloonColor = pickColor()
      problemList.value.push(p)
    })
    pickerVisible.value = false
  } catch (err) {
    console.error(err)
  } finally {
    picking.value = false
  }
}

// ---------- 校验规则 ----------
const disabledDate = (time) => time < new Date(new Date().setHours(0, 0, 0, 0))

const rules = {
  name: [{ required: true, message: '请输入比赛名称', trigger: 'blur' }],
  contestDate: [{ required: true, message: '请选择比赛日期', trigger: 'change' }],
  contestTime: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  duration: [{ required: true, message: '请输入比赛时长', trigger: 'blur' }],
}

// ---------- 加载现有比赛 ----------
const loadContest = async () => {
  const res = await getContestInfo(route.params.id)
  const data = res.data || {}
  contestForm.value.name = data.title || ''
  contestForm.value.coverUrl = data.coverUrl || ''
  contestForm.value.description = data.description || ''
  contestForm.value.ruleType = data.ruleType || 1
  contestForm.value.rated = !!data.rated
  // 比赛是否已开始：开始后锁定赛制
  started.value = !!data.startTime && Date.now() >= data.startTime
  // 起止毫秒 → 日期 + 开始时间 + 时长(分钟)
  if (data.startTime) {
    const s = new Date(data.startTime)
    const pad = (n) => String(n).padStart(2, '0')
    contestForm.value.contestDate = `${s.getFullYear()}-${pad(s.getMonth() + 1)}-${pad(s.getDate())}`
    contestForm.value.contestTime = `${pad(s.getHours())}:${pad(s.getMinutes())}`
    if (data.endTime) {
      contestForm.value.duration = Math.max(1, Math.round((data.endTime - data.startTime) / 60000))
    }
  }
  // 题目：补拉完整信息(时间/内存/标签)，保留已存的气球色
  const list = data.problems || []
  const details = await Promise.all(list.map((p) => getProblem(p.id).catch(() => null)))
  problemList.value = list.map((p, i) => {
    const d = details[i]?.data
    const base = d ? toProblem(d) : { id: p.id, name: p.name, tags: [] }
    base.balloonColor = p.color || ''
    base.score = p.score || 100
    return base
  })
}

onMounted(loadContest)

// ---------- 提交 ----------
const submitForm = async () => {
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  if (problemList.value.length === 0) {
    ElMessage.error('请至少添加一道题目')
    return
  }
  const startMs = new Date(`${contestForm.value.contestDate}T${contestForm.value.contestTime}:00`).getTime()
  const payload = {
    title: contestForm.value.name,
    ruleType: contestForm.value.ruleType,
    rated: contestForm.value.rated,
    description: contestForm.value.description,
    coverUrl: contestForm.value.coverUrl,
    startTime: startMs,
    endTime: startMs + Number(contestForm.value.duration) * 60000,
    problems: problemList.value.map((p) => ({ problemId: p.id, color: p.balloonColor || '', score: p.score || 100 })),
  }

  submitting.value = true
  try {
    await updateContest(route.params.id, payload)
    ElMessage.success('保存成功')
    router.push(`/contest/${route.params.id}`)
  } catch (err) {
    console.error(err)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.create-contest-page {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 20px;
}

.page-header {
  margin-bottom: 14px;

  .title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 22px;
    font-weight: 600;
  }
}

.layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 300px;
  gap: 16px;
  align-items: start;
}

.block {
  background: #fff;
  border: 1px solid var(--el-border-color);
  padding: 16px;
  margin-bottom: 16px;

  :deep(.md-editor) {
    width: 100% !important;
  }
}

.block-title {
  font-weight: 600;
  margin-bottom: 14px;
}

.acm-rule {
  font-weight: 600;
}

.block-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;

  .block-title {
    margin-bottom: 0;
  }

  .block-tools {
    display: flex;
    align-items: center;
    gap: 8px;
  }
}

.row {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}

.unit {
  margin-left: 8px;
  color: var(--el-text-color-regular);
  font-size: 13px;
}

.aside {
  position: sticky;
  top: 16px;
}

.overview-card {
  background: #fff;
  padding: 16px;
  box-shadow: var(--shadow-card);
}

.side-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 8px;
}

.ov-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 9px 0;
  border-bottom: 1px solid #f0f0f0;
  font-size: 14px;
  color: var(--el-text-color-regular);

  b {
    color: var(--el-text-color-primary);
    font-weight: 600;
    text-align: right;
  }

  .ellipsis {
    max-width: 170px;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .warn {
    color: #ED3F14;
  }
}

.full {
  width: 100%;
  margin-top: 16px;
}

.problem-table {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
  font-size: 14px;

  th,
  td {
    padding: 10px;
    border-bottom: 1px solid #e4e7ed;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }

  .left {
    text-align: left;
  }

  .center {
    text-align: center;
  }

  .label {
    font-weight: 600;
  }

  tbody tr {
    transition: background 0.15s ease;

    &:hover {
      background: #f8f9fb;
    }

    &.dragging {
      opacity: 0.4;
    }

    &.drag-over {
      background: #ecf5ff;
      box-shadow: inset 0 2px 0 #409eff;
    }
  }

  .handle {
    cursor: grab;
    color: #b0b3bd;

    &:active {
      cursor: grabbing;
    }
  }
}

.picker-search {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.picker-table {
  font-size: 14px;

  .pick-row {
    cursor: pointer;
  }
}

.picker-foot {
  display: flex;
  justify-content: center;
  margin-top: 12px;
}

.picked-count {
  margin-right: auto;
  color: var(--el-text-color-regular);
  font-size: 13px;
}

@media (max-width: 1100px) {
  .layout {
    grid-template-columns: 1fr;
  }

  .aside {
    position: static;
  }
}
</style>
