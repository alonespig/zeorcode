<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRouter, onBeforeRouteLeave } from "vue-router";
import { ArrowLeft, Rank } from "@element-plus/icons-vue";
import { ElMessageBox } from "element-plus";
import { message } from "@/utils/message";
import { getProblem, getTagList, type TagItem } from "@/api/admin/problems";
import {
  createProblemSet,
  getProblemSetDetail,
  updateProblemSet,
  type SaveProblemSetParams
} from "@/api/admin/problemsets";
import ProblemPickerDialog from "./ProblemPickerDialog.vue";

defineOptions({ name: "ProblemSetForm" });

const props = defineProps<{
  /** 编辑时传入题单对外 id；新建时不传 */
  problemSetId?: number;
}>();

interface ProblemSetFormModel {
  title: string;
  description: string;
  /** 0 草稿 / 1 已发布 */
  published: number;
  /** 0 公开 / 1 需邀请码 */
  visibility: number;
  inviteCode: string;
  tagIds: number[];
}

const router = useRouter();
const isEdit = computed(() => props.problemSetId !== undefined);
// 编辑回填时记录的原始可见性：0 公开 / 1 需邀请码
const originalVisibility = ref(0);
// 邀请码输入框提示：编辑原邀请码题单且留空时保留原码，其余场景需填写
const inviteCodePlaceholder = computed(() =>
  isEdit.value && originalVisibility.value === 1
    ? "留空则保留原邀请码"
    : "请输入邀请码"
);

const tagOptions = ref<TagItem[]>([]);
const problems = ref<{ id: string; name: string }[]>([]);
const form = reactive<ProblemSetFormModel>({
  title: "",
  description: "",
  published: 1,
  visibility: 0,
  inviteCode: "",
  tagIds: []
});

const loadingDetail = ref(false);
const saving = ref(false);
const dirty = ref(false);

// 除初次回填外，任何表单或题目变化都标记为未保存
watch(
  [form, problems],
  () => {
    if (!loadingDetail.value) dirty.value = true;
  },
  { deep: true, flush: "sync" }
);

const pidInput = ref("");
const addingPid = ref(false);
const pickerVisible = ref(false);
const existingIds = computed(() => problems.value.map(p => p.id));

const dragIndex = ref(-1);
const dragOverIndex = ref(-1);

async function loadTags() {
  try {
    const res = await getTagList();
    tagOptions.value = res.data?.tags ?? [];
  } catch {
    // 错误信息已由 http 拦截器统一提示
  }
}

async function loadDetail() {
  if (props.problemSetId === undefined) return;
  loadingDetail.value = true;
  try {
    const res = await getProblemSetDetail(props.problemSetId);
    const d = res.data;
    if (!d) return;
    form.title = d.title;
    form.description = d.description ?? "";
    form.published = d.published;
    form.visibility = d.visibility;
    form.inviteCode = "";
    originalVisibility.value = d.visibility;
    form.tagIds = (d.tags ?? []).map(t => t.id);
    problems.value = (d.problems ?? []).map(p => ({ id: p.id, name: p.name }));
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    loadingDetail.value = false;
    dirty.value = false;
  }
}

async function addProblemById() {
  const id = pidInput.value.trim();
  if (!id) {
    message("请输入题号", { type: "warning" });
    return;
  }
  if (problems.value.some(p => p.id === id)) {
    message("题目已添加", { type: "warning" });
    return;
  }
  addingPid.value = true;
  try {
    const res = await getProblem(id);
    const problem = res.data;
    if (!problem) return;
    problems.value.push({ id: problem.id, name: problem.name });
    pidInput.value = "";
  } catch {
    // 错误信息已由 http 拦截器统一提示（如题目不存在）
  } finally {
    addingPid.value = false;
  }
}

function removeProblem(index: number) {
  problems.value.splice(index, 1);
}

function addPickedProblems(items: { id: string; name: string }[]) {
  items.forEach(item => {
    if (!problems.value.some(p => p.id === item.id)) {
      problems.value.push({ id: item.id, name: item.name });
    }
  });
}

function onDragStart(index: number) {
  dragIndex.value = index;
}

function onDragOver(index: number) {
  dragOverIndex.value = index;
}

function onDragEnd() {
  dragIndex.value = -1;
  dragOverIndex.value = -1;
}

function onDrop(index: number) {
  const from = dragIndex.value;
  if (from !== -1 && from !== index) {
    const [moved] = problems.value.splice(from, 1);
    problems.value.splice(index, 0, moved);
  }
  onDragEnd();
}

async function save() {
  if (!form.title.trim()) {
    message("请填写标题", { type: "warning" });
    return;
  }
  // 邀请码可见时必须填写；仅编辑原邀请码题单时留空表示保留原邀请码
  if (
    form.visibility === 1 &&
    (!isEdit.value || originalVisibility.value === 0) &&
    !form.inviteCode.trim()
  ) {
    message("选择邀请码可见时必须填写邀请码", { type: "warning" });
    return;
  }
  const payload: SaveProblemSetParams = {
    title: form.title.trim(),
    description: form.description,
    published: form.published,
    visibility: form.visibility,
    inviteCode: form.inviteCode.trim(),
    tagIds: form.tagIds,
    problems: problems.value.map(p => p.id)
  };
  saving.value = true;
  try {
    if (isEdit.value && props.problemSetId !== undefined) {
      await updateProblemSet(props.problemSetId, payload);
    } else {
      await createProblemSet(payload);
    }
    message("已保存", { type: "success" });
    dirty.value = false;
    router.push("/content/problemsets");
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    saving.value = false;
  }
}

function goBack() {
  router.push("/content/problemsets");
}

onBeforeRouteLeave(async () => {
  if (!dirty.value) return true;
  try {
    await ElMessageBox.confirm("当前修改尚未保存，确定要离开吗？", "提示", {
      type: "warning"
    });
    return true;
  } catch {
    return false;
  }
});

onMounted(() => {
  loadTags();
  loadDetail();
});
</script>

<template>
  <div>
    <div class="mb-4 flex items-center gap-3">
      <el-button :icon="ArrowLeft" @click="goBack">返回</el-button>
      <span class="text-base font-medium">
        {{ isEdit ? "编辑题单" : "新建题单" }}
      </span>
    </div>

    <el-card v-loading="loadingDetail" shadow="never">
      <el-form :model="form" label-width="96px" class="max-w-3xl">
        <el-form-item label="标题" required>
          <el-input
            v-model="form.title"
            maxlength="100"
            show-word-limit
            placeholder="题单标题"
          />
        </el-form-item>

        <el-form-item label="标签">
          <el-select
            v-model="form.tagIds"
            multiple
            filterable
            clearable
            placeholder="选择标签（可多选）"
            style="width: 100%"
          >
            <el-option
              v-for="tag in tagOptions"
              :key="tag.id"
              :label="tag.name"
              :value="tag.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="8"
            placeholder="题单描述（支持 Markdown）"
          />
        </el-form-item>

        <el-form-item label="访问方式">
          <el-radio-group v-model="form.visibility">
            <el-radio :value="0">公开</el-radio>
            <el-radio :value="1">需要邀请码</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item v-if="form.visibility === 1" label="邀请码">
          <el-input
            v-model="form.inviteCode"
            maxlength="32"
            :placeholder="inviteCodePlaceholder"
            style="max-width: 240px"
          />
        </el-form-item>

        <el-form-item label="发布状态">
          <el-radio-group v-model="form.published">
            <el-radio :value="1">已发布</el-radio>
            <el-radio :value="0">草稿</el-radio>
          </el-radio-group>
          <div class="text-xs text-gray-400">草稿在前台完全不可见</div>
        </el-form-item>

        <el-form-item label="题目">
          <div class="w-full">
            <div class="mb-2 flex flex-wrap items-center gap-2">
              <el-input
                v-model="pidInput"
                class="w-48"
                placeholder="输入题号，如 P1001"
                :disabled="addingPid"
                @keyup.enter="addProblemById"
              />
              <el-button :loading="addingPid" @click="addProblemById">
                添加
              </el-button>
              <el-button type="primary" @click="pickerVisible = true">
                从题库选择
              </el-button>
              <span class="text-xs text-gray-400">
                已选 {{ problems.length }} 道 · 拖动行可调整顺序
              </span>
            </div>

            <div
              v-if="problems.length"
              class="divide-y divide-gray-200 rounded border border-gray-200"
            >
              <div
                v-for="(problem, index) in problems"
                :key="problem.id"
                class="flex items-center gap-3 px-3 py-2"
                :class="{
                  'bg-gray-50': dragOverIndex === index && dragIndex !== index,
                  'opacity-50': dragIndex === index
                }"
                draggable="true"
                @dragstart="onDragStart(index)"
                @dragover.prevent="onDragOver(index)"
                @dragend="onDragEnd"
                @drop="onDrop(index)"
              >
                <el-icon class="cursor-grab text-gray-400"><Rank /></el-icon>
                <span class="w-6 text-right text-gray-400">
                  {{ index + 1 }}
                </span>
                <span class="w-24 font-mono">{{ problem.id }}</span>
                <span class="flex-1 truncate">{{ problem.name }}</span>
                <el-button
                  link
                  type="danger"
                  size="small"
                  @click="removeProblem(index)"
                >
                  移除
                </el-button>
              </div>
            </div>
            <el-empty v-else description="还没有添加题目" :image-size="60" />
          </div>
        </el-form-item>
      </el-form>
    </el-card>

    <div class="mt-4 flex items-center gap-2">
      <el-button type="primary" :loading="saving" @click="save">
        保存
      </el-button>
      <el-button :disabled="saving" @click="goBack">取消</el-button>
    </div>

    <ProblemPickerDialog
      v-model="pickerVisible"
      :existing-ids="existingIds"
      @add="addPickedProblems"
    />
  </div>
</template>
