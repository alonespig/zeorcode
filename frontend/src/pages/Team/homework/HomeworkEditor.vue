<script setup>
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ArrowLeft } from "@element-plus/icons-vue";
import HomeworkForm from "./HomeworkForm.vue";

const route = useRoute();
const router = useRouter();
const formRef = ref(null);

const isEdit = computed(() => Boolean(route.params.hid));

const backToList = () => {
  router.push({ name: "TeamHomework", params: { id: route.params.id } });
};

const save = () => formRef.value?.save();
</script>

<template>
  <div class="homework-editor-page">
    <section class="f-panel editor-shell">
      <header class="editor-header">
        <el-button :icon="ArrowLeft" @click="backToList">返回</el-button>
        <el-button type="primary" :loading="formRef?.saving" @click="save">
          {{ isEdit ? "保存修改" : "布置作业" }}
        </el-button>
      </header>

      <HomeworkForm ref="formRef" :team-id="route.params.id" :homework-id="route.params.hid"
        @saved="backToList" @cancel="backToList" />

      <div class="editor-actions">
        <el-button @click="backToList">取消</el-button>
        <el-button type="primary" :loading="formRef?.saving" @click="save">
          {{ isEdit ? "保存修改" : "布置作业" }}
        </el-button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.homework-editor-page {
  min-width: 0;
}

.editor-shell {
  padding: 24px 28px 28px;
  overflow: hidden;
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}

.editor-actions {
  display: flex;
  justify-content: center;
  gap: 8px;
  margin-top: 24px;
}

@media (max-width: 768px) {
  .editor-shell {
    padding: 20px 16px;
  }
}
</style>
