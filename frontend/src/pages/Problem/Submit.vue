<template>
  <div class="submit-page page-container">
    <div class="submit-container">
      <div class="submit-header">
        <span>提交代码</span>
      </div>
      <div class="form">
        <div>
          <span class="label">题目</span>
          <span class="f-link">{{ problem.name }}</span>
        </div>
        <div class="form-item">
          <span class="label">语言</span>
          <select v-model="form.language" placeholder="选择语言">
            <option v-for="item in languageList" :key="item.id" :value="item.id">
              {{ item.name }}
            </option>
          </select>
        </div>

        <div class="form-item">
          <span class="label">源代码</span>
          <textarea v-model="form.code" type="textarea" class="editor"></textarea>
        </div>

        <div class="form-actions">
          <el-button type="primary" @click="submitHandler">提交代码</el-button>
          <el-button @click="router.push({ name: 'ProblemDetail', params: { id: route.params.id } })">返回题目</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="js">
import { ref, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { useRoute, useRouter } from "vue-router";
import { submit } from "@/api/user";
import { getProblem } from "@/api/problems";
import { useLanguages } from "@/hooks/useLanguages";


const route = useRoute();
const router = useRouter();
const { languageList, loadLanguages } = useLanguages();

const form = ref({
  label: null,
  language: null,
  code: null,
})

const problem = ref({
  title: "",
  memoryLimitMb: 0,
  timeLimitMs: 0,
  submitCount: 0,
  acceptedCount: 0,
  acceptanceRate: 0,
});

const submitHandler = async () => {
  let formData = {
    problemId: route.params.id,
    language: form.value.language,
    code: form.value.code,
  }
  try {
    const res = await submit(formData);
    ElMessage.success("提交成功");
    router.push({ name: "SubmissionDetail", params: { id: res.data.submissionID } });
  } catch (error) {
    // 业务错误 msg 已由 axios 拦截器统一弹出
    console.error(error);
  }
}


onMounted(async () => {
  await loadLanguages();
  const res = await getProblem(route.params.id);
  problem.value = res.data || {};
});
</script>

<style scoped lang="scss">
.submit-page {
  width: 100%;

  .submit-container {
    width: 70%;
    background-color: #ffffff;
    padding: 20px;
    border-radius: 2px;

    .submit-header {
      font-size: 20px;
      font-weight: normal;
      margin-bottom: 20px;
    }

    .form {
      width: 100%;
      display: flex;
      flex-direction: column;
      gap: 30px;

      .label {
        width: 80px;
        font-weight: 500;
      }

      .form-item {
        display: flex;
        align-items: center;
        justify-content: flex-start;

        select {
          width: 200px;
          height: 25px;
          line-height: 25px;
          border: 1px solid #ccc;
          padding: 0 6px;
        }

        .editor {
          font-family: "Fira Code", monospace;
          font-size: 14px;
          color: black;
          width: 800px;
          height: 300px;
          padding: 10px 15px;
          border-radius: 2px;
          outline: none;
          border-color: #409eff;
        }
      }

      .form-actions {
        display: flex;
        justify-content: space-around;
        gap: 10px;
      }
    }
  }


}
</style>
