<template>
  <el-drawer v-model="visible" :title="isEdit ? '编辑作业' : '布置作业'" direction="rtl" size="min(1000px, 92vw)"
    :close-on-click-modal="false">
    <HomeworkForm ref="formRef" :team-id="teamId" :homework-id="homeworkId" @saved="onSaved"
      @cancel="visible = false" />

    <template #footer>
      <div class="drawer-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="formRef?.saving" @click="save">
          {{ isEdit ? "保存" : "布置作业" }}
        </el-button>
      </div>
    </template>
  </el-drawer>
</template>

<script setup>
import { computed, ref } from "vue";
import HomeworkForm from "./HomeworkForm.vue";

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  teamId: { type: [String, Number], default: "" },
  homeworkId: { type: [String, Number], default: "" },
});
const emit = defineEmits(["update:modelValue", "saved"]);

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});

const isEdit = computed(() => Boolean(props.homeworkId));

const formRef = ref(null);

const save = () => formRef.value?.save();

const onSaved = () => {
  emit("saved");
  visible.value = false;
};
</script>

<style scoped>
.drawer-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px 0;
}
</style>
