<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { CircleCheckFilled, Warning } from "@element-plus/icons-vue";
import type { AgentBlock, AgentTableColumn } from "@/api/admin/agent";

defineOptions({ name: "AgentBlockRenderer" });
const props = defineProps<{
  blocks: AgentBlock[];
  pendingRequestId: number;
  pendingActionId: number;
  running: boolean;
}>();
const emit = defineEmits<{
  answer: [requestId: number, value: unknown];
  approve: [actionId: number, draftVersion: number];
  reject: [actionId: number, draftVersion: number];
}>();

interface HomeworkSettings {
  title: string;
  description: string;
  problemCount: number;
  difficulty: number;
  repeatPolicy: string;
}

const values = reactive<Record<number, unknown>>({});
const settings = reactive<Record<number, HomeworkSettings>>({});
const ranges = reactive<Record<number, [string, string] | []>>({});
const submitting = computed(() => props.running);
const defaultTimes: [Date, Date] = [
  new Date(2000, 1, 1, 8),
  new Date(2000, 1, 1, 23, 59)
];
const expandedUnknown = ref<number[]>([]);

function requestId(block: AgentBlock) {
  return block.requestId || 0;
}

function isInteractionDisabled(block: AgentBlock) {
  return requestId(block) !== props.pendingRequestId;
}

function updateValue(block: AgentBlock, value: unknown) {
  values[requestId(block)] = value;
}

function multipleValue(block: AgentBlock) {
  const value = values[requestId(block)];
  return Array.isArray(value) ? value : [];
}

function singleValue(block: AgentBlock) {
  const value = values[requestId(block)];
  return typeof value === "string" ||
    typeof value === "number" ||
    typeof value === "boolean"
    ? value
    : undefined;
}

function ensureSettings(block: AgentBlock) {
  const id = requestId(block);
  if (!settings[id]) {
    const payload = block.payload || {};
    settings[id] = {
      title: String(payload.defaultTitle || ""),
      description: "",
      problemCount: Number(payload.defaultProblemCount || 8),
      difficulty: Number(payload.defaultDifficulty || 0),
      repeatPolicy: String(payload.defaultRepeatPolicy || "exclude_all")
    };
    const start = String(payload.defaultStartTime || "");
    const end = String(payload.defaultEndTime || "");
    ranges[id] = start && end ? [start, end] : [];
  }
  return settings[id];
}

function submitSettings(block: AgentBlock) {
  const id = requestId(block);
  const form = ensureSettings(block);
  const range = ranges[id];
  if (!String(form.title || "").trim() || range?.length !== 2) return;
  emit("answer", id, {
    ...form,
    title: String(form.title).trim(),
    startTime: range[0],
    endTime: range[1]
  });
}

function columnStyle(column: AgentTableColumn) {
  return {
    width: column.width ? `${column.width}px` : undefined,
    textAlign: column.align || "left"
  };
}

function displayCell(value: unknown, formatter?: string) {
  if (formatter === "difficulty") {
    return (
      ({ 1: "入门", 2: "普及", 3: "提高" } as Record<number, string>)[
        Number(value)
      ] || "未知"
    );
  }
  if (Array.isArray(value)) return value.join("、");
  if (value === null || value === undefined || value === "") return "—";
  return String(value);
}

function payloadNumber(block: AgentBlock, key: string) {
  return Number(block.payload?.[key] || 0);
}
</script>

<template>
  <div class="block-list">
    <template
      v-for="(block, index) in blocks"
      :key="`${block.type}-${requestId(block)}-${index}`"
    >
      <section
        v-if="block.type === 'single_select' || block.type === 'multi_select'"
        class="block-card"
      >
        <h4>{{ block.title }}</h4>
        <p v-if="block.description">{{ block.description }}</p>
        <el-checkbox-group
          v-if="block.type === 'multi_select'"
          :model-value="multipleValue(block)"
          class="option-list"
          :disabled="isInteractionDisabled(block)"
          @update:model-value="updateValue(block, $event)"
        >
          <el-checkbox
            v-for="option in block.options || []"
            :key="String(option.value)"
            :value="option.value"
            :disabled="option.disabled"
            border
          >
            {{ option.label }}
            <el-tag v-if="option.recommended" size="small" type="success"
              >推荐</el-tag
            >
            <small v-if="option.description">{{ option.description }}</small>
          </el-checkbox>
        </el-checkbox-group>
        <el-radio-group
          v-else
          :model-value="singleValue(block)"
          class="option-list"
          :disabled="isInteractionDisabled(block)"
          @update:model-value="updateValue(block, $event)"
        >
          <el-radio
            v-for="option in block.options || []"
            :key="String(option.value)"
            :value="option.value"
            :disabled="option.disabled"
            border
          >
            {{ option.label }}
            <el-tag v-if="option.recommended" size="small" type="success"
              >推荐</el-tag
            >
            <small v-if="option.description">{{ option.description }}</small>
          </el-radio>
        </el-radio-group>
        <footer>
          <span v-if="isInteractionDisabled(block)">此选项已处理</span>
          <el-button
            v-else
            type="primary"
            :loading="submitting"
            :disabled="values[requestId(block)] === undefined"
            @click="emit('answer', requestId(block), values[requestId(block)])"
            >确认选择</el-button
          >
        </footer>
      </section>

      <section
        v-else-if="block.type === 'homework_settings'"
        class="block-card"
      >
        <h4>{{ block.title }}</h4>
        <p v-if="block.description">{{ block.description }}</p>
        <el-form label-position="top" class="settings-form">
          <div class="form-grid">
            <el-form-item label="作业名称">
              <el-input
                v-model="ensureSettings(block).title"
                maxlength="100"
                :disabled="isInteractionDisabled(block)"
              />
            </el-form-item>
            <el-form-item label="题目数量">
              <el-input-number
                v-model="ensureSettings(block).problemCount"
                :min="1"
                :max="30"
                controls-position="right"
                :disabled="isInteractionDisabled(block)"
              />
            </el-form-item>
            <el-form-item label="难度">
              <el-select
                v-model="ensureSettings(block).difficulty"
                :disabled="isInteractionDisabled(block)"
              >
                <el-option label="不限难度" :value="0" />
                <el-option label="入门" :value="1" />
                <el-option label="普及" :value="2" />
                <el-option label="提高" :value="3" />
              </el-select>
            </el-form-item>
            <el-form-item label="历史作业题目">
              <el-select
                v-model="ensureSettings(block).repeatPolicy"
                :disabled="isInteractionDisabled(block)"
              >
                <el-option label="排除以前使用过的题目" value="exclude_all" />
                <el-option label="允许重复使用" value="allow_history" />
              </el-select>
            </el-form-item>
          </div>
          <el-form-item label="开始和结束时间">
            <el-date-picker
              v-model="ranges[requestId(block)]"
              type="datetimerange"
              value-format="YYYY-MM-DD HH:mm:ss"
              format="YYYY-MM-DD HH:mm"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              :default-time="defaultTimes"
              :disabled="isInteractionDisabled(block)"
            />
          </el-form-item>
          <el-form-item label="作业说明（可选）">
            <el-input
              v-model="ensureSettings(block).description"
              type="textarea"
              :rows="2"
              maxlength="1000"
              :disabled="isInteractionDisabled(block)"
            />
          </el-form-item>
        </el-form>
        <footer>
          <span v-if="isInteractionDisabled(block)">此设置已提交</span>
          <el-button
            v-else
            type="primary"
            :loading="submitting"
            @click="submitSettings(block)"
          >
            生成题目方案
          </el-button>
        </footer>
      </section>

      <section v-else-if="block.type === 'data_table'" class="table-card">
        <header>
          <h4>{{ block.title }}</h4>
          <span>{{ block.rows?.length || 0 }} 项</span>
        </header>
        <div class="table-scroll">
          <table>
            <thead>
              <tr>
                <th
                  v-for="column in block.columns || []"
                  :key="column.key"
                  :style="columnStyle(column)"
                >
                  {{ column.label }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(row, rowIndex) in block.rows || []"
                :key="String(row.id || rowIndex)"
              >
                <td
                  v-for="column in block.columns || []"
                  :key="column.key"
                  :style="{ textAlign: column.align || 'left' }"
                >
                  {{ displayCell(row[column.key], column.formatter) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section
        v-else-if="block.type === 'notice'"
        class="block-card notice-card"
      >
        <div class="notice-title">
          <el-icon><Warning /></el-icon>
          <div>
            <h4>{{ block.title }}</h4>
            <p>{{ block.description }}</p>
          </div>
        </div>
        <div class="notice-actions">
          <el-button
            v-for="option in block.options || []"
            :key="String(option.value)"
            :disabled="
              isInteractionDisabled(block) || submitting || option.disabled
            "
            @click="emit('answer', requestId(block), option.value)"
            >{{ option.label }}</el-button
          >
        </div>
      </section>

      <section
        v-else-if="block.type === 'action_preview'"
        class="block-card preview-card"
      >
        <h4>{{ block.title }}</h4>
        <p>{{ block.description }}</p>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="团队">{{
            block.payload?.teamName
          }}</el-descriptions-item>
          <el-descriptions-item label="作业名称">{{
            block.payload?.title
          }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{
            block.payload?.startTime
          }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{
            block.payload?.endTime
          }}</el-descriptions-item>
          <el-descriptions-item label="题目数量"
            >{{ block.payload?.problemCount }} 题</el-descriptions-item
          >
          <el-descriptions-item label="知识点">{{
            displayCell(block.payload?.knowledgeTags)
          }}</el-descriptions-item>
        </el-descriptions>
        <footer>
          <span v-if="payloadNumber(block, 'actionId') !== pendingActionId"
            >此操作已处理</span
          >
          <template v-else>
            <el-button
              :disabled="submitting"
              @click="
                emit(
                  'reject',
                  payloadNumber(block, 'actionId'),
                  payloadNumber(block, 'draftVersion')
                )
              "
              >取消</el-button
            >
            <el-button
              type="primary"
              :loading="submitting"
              @click="
                emit(
                  'approve',
                  payloadNumber(block, 'actionId'),
                  payloadNumber(block, 'draftVersion')
                )
              "
              >确认创建</el-button
            >
          </template>
        </footer>
      </section>

      <section v-else-if="block.type === 'action_result'" class="result-card">
        <el-icon><CircleCheckFilled /></el-icon>
        <div>
          <h4>{{ block.title }}</h4>
          <p>{{ block.payload?.title }}</p>
        </div>
        <a
          v-if="block.payload?.link"
          :href="String(block.payload.link)"
          target="_blank"
          rel="noopener"
          >查看作业</a
        >
      </section>

      <el-collapse v-else v-model="expandedUnknown" class="unknown-block">
        <el-collapse-item
          :name="index"
          :title="block.title || `未识别内容：${block.type}`"
        >
          <pre>{{ JSON.stringify(block.payload || block, null, 2) }}</pre>
        </el-collapse-item>
      </el-collapse>
    </template>
  </div>
</template>

<style scoped>
.block-list {
  margin-top: 12px;
}
.block-card {
  margin-top: 12px;
  padding: 16px;
  border: 1px solid var(--el-border-color-light);
  background: var(--el-fill-color-extra-light);
}
h4 {
  margin: 0;
  color: var(--el-text-color-primary);
  font-size: 14px;
}
.block-card > p,
.notice-title p {
  margin: 5px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
  line-height: 1.6;
}
.option-list {
  width: 100%;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  margin-top: 13px;
}
.option-list :deep(.el-checkbox),
.option-list :deep(.el-radio) {
  width: 100%;
  height: auto;
  min-height: 42px;
  margin: 0;
  padding: 8px 10px;
}
.option-list small {
  display: block;
  margin-top: 2px;
  color: var(--el-text-color-secondary);
  white-space: normal;
}
footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 13px;
}
footer span {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}
.settings-form {
  margin-top: 14px;
}
.form-grid {
  display: grid;
  grid-template-columns: 1.5fr 0.7fr;
  gap: 0 14px;
}
.settings-form :deep(.el-select),
.settings-form :deep(.el-input-number),
.settings-form :deep(.el-date-editor) {
  width: 100%;
}
.table-card {
  margin-top: 12px;
  overflow: hidden;
  border: 1px solid var(--el-border-color-light);
}
.table-card header {
  height: 43px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 14px;
  background: var(--el-fill-color-extra-light);
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.table-card header span {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.table-scroll {
  overflow-x: auto;
}
table {
  width: 100%;
  min-width: 640px;
  border-collapse: collapse;
}
th,
td {
  height: 42px;
  padding: 7px 11px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  font-size: 12px;
}
th {
  color: var(--el-text-color-secondary);
  background: var(--el-fill-color-extra-light);
}
td {
  color: var(--el-text-color-regular);
}
tbody tr:last-child td {
  border-bottom: 0;
}
.notice-card {
  border-color: var(--el-color-warning-light-7);
  background: var(--el-color-warning-light-9);
}
.notice-title {
  display: flex;
  gap: 9px;
  color: var(--el-color-warning);
}
.notice-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}
.preview-card {
  border-left: 3px solid var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}
.preview-card :deep(.el-descriptions) {
  margin-top: 13px;
}
.result-card {
  min-height: 66px;
  display: flex;
  align-items: center;
  gap: 11px;
  margin-top: 12px;
  padding: 14px 16px;
  border: 1px solid var(--el-color-success-light-7);
  background: var(--el-color-success-light-9);
}
.result-card > .el-icon {
  color: var(--el-color-success);
  font-size: 25px;
}
.result-card div {
  min-width: 0;
  flex: 1;
}
.result-card p {
  margin: 4px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.result-card a {
  color: var(--el-color-success);
  font-size: 12px;
}
.unknown-block {
  margin-top: 12px;
}
.unknown-block pre {
  overflow-x: auto;
  font-size: 11px;
}
@media (max-width: 720px) {
  .option-list,
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>
