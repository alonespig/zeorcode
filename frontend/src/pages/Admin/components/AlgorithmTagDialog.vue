<script setup>
import { reactive, useTemplateRef, watch } from "vue";

const props = defineProps({
  title: {
    type: String,
    required: true,
  },
  initialName: {
    type: String,
    default: "",
  },
  submitting: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(["submit"]);
const visible = defineModel({ type: Boolean, default: false });
const formRef = useTemplateRef("formRef");
const form = reactive({ name: "" });
const rules = {
  name: [
    { required: true, message: "请输入标签名称", trigger: "blur" },
    { max: 64, message: "标签名称不能超过 64 个字符", trigger: "blur" },
  ],
};

watch(visible, (isVisible) => {
  if (!isVisible) return;
  form.name = props.initialName;
  formRef.value?.clearValidate();
});

const submit = async () => {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }
  emit("submit", form.name.trim());
};
</script>

<template>
  <el-dialog v-model="visible" :title="title" width="460px" :close-on-click-modal="!submitting">
    <el-form ref="formRef" :model="form" :rules="rules" label-width="88px" @submit.prevent="submit">
      <el-form-item label="标签名称" prop="name">
        <el-input v-model="form.name" maxlength="64" show-word-limit placeholder="例如：动态规划" autofocus />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button :disabled="submitting" @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
    </template>
  </el-dialog>
</template>
