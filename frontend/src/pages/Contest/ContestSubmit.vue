<template>
  <div class="form">
    <div class="form-item">
      <span class="label">题目</span>
      <select v-model="form.label" placeholder="选择题目">
        <option v-for="item in problemList" :key="item.label" :value="item.label">
          {{ item.name }}
        </option>
      </select>
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
      <el-button @click="form = { label: null, language: null, code: null }">重置</el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getContestProblemList,submit } from "@/api/contest";
import { useLanguages } from "@/hooks/useLanguages";

const route = useRoute();
const router = useRouter();
const params = computed(() => route.params);

const problemList = ref([])
const { languageList, loadLanguages } = useLanguages();

const form = ref({
  label: route.query?.id || null,
  language: null,
  code: null,
})

const submitHandler = async () => {
  let formData = {
    ...form.value,
    contestId: Number(params.value.id),
  }
  await submit(formData)
  router.push({
    path: `/contest/${params.value.id}/submission`,
  })
}


const getProblemList = async () => {
  const res = await getContestProblemList(params.value.id);
  problemList.value = res.data.problemList.map(item => ({
    name: item.label + " - " + item.name,
    label: item.label,
  }))
}

onMounted(() => {
  loadLanguages()
  getProblemList()
})

</script>

<style scoped lang="scss">
.form {
  max-width: 900px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 30px;

  select {
    width: 300px;
    height: 25px;
    line-height: 25px;
    border: 1px solid #ccc;
    padding: 0 5px;
  }

  .label {
    width: 80px;
    font-weight: 500;
    // color: #606266;
  }

  .form-item {
    display: flex;
    align-items: center;

    .editor {
      font-family: "Fira Code", monospace;
      font-size: 14px;
      background: #fff;
      color: black;
      // border-radius: 8px;
    }

    textarea {
      width: 800px;
      height: 600px;
      padding: 10px 10px;
      border: 1px solid #ddd;
      border-radius: 2px;
      outline: none;
    }

    textarea:focus {
      border-color: #409eff;
      box-shadow: 0 0 5px rgba(64, 158, 255, 0.3);
    }
  }

  .form-actions {
    display: flex;
    justify-content: space-around;
    gap: 10px;
  }
}
</style>
