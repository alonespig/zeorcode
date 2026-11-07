<script setup lang="ts">
import { computed, ref } from "vue";
import ProblemSetForm from "./ProblemSetForm.vue";

defineOptions({ name: "ProblemSetDrawer" });

const props = defineProps<{
  modelValue: boolean;
  /** 编辑时传入题单对外 id；新建时不传 */
  problemSetId?: number;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: boolean): void;
  (e: "saved"): void;
}>();

const visible = computed({
  get: () => props.modelValue,
  set: value => emit("update:modelValue", value)
});

const isEdit = computed(() => props.problemSetId !== undefined);
const formRef = ref<InstanceType<typeof ProblemSetForm> | null>(null);

function save() {
  formRef.value?.save();
}

function onSaved() {
  emit("saved");
  visible.value = false;
}
</script>

<template>
  <el-drawer
    v-model="visible"
    :title="isEdit ? '编辑题单' : '新建题单'"
    direction="rtl"
    size="min(1000px, 92vw)"
    :close-on-click-modal="false"
  >
    <ProblemSetForm
      ref="formRef"
      embedded
      :problem-set-id="problemSetId"
      @saved="onSaved"
    />

    <template #footer>
      <div class="drawer-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="formRef?.saving" @click="save">
          {{ isEdit ? "保存" : "创建" }}
        </el-button>
      </div>
    </template>
  </el-drawer>
</template>

<style scoped>
.drawer-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px 0;
}
</style>
