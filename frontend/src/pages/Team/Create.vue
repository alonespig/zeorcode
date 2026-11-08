<template>
  <div class="team-create-page page-container">
    <section class="f-panel create-panel">
      <header class="page-header">
        <el-button :icon="ArrowLeft" @click="cancel">返回</el-button>
        <el-button type="primary" :loading="creating" @click="submit">创建</el-button>
      </header>

      <el-form :model="form" label-position="top">
        <section class="form-section">
          <el-form-item label="团队名称" required>
            <el-input v-model="form.name" maxlength="60" show-word-limit placeholder="团队名称" />
          </el-form-item>
          <el-form-item label="团队封面">
            <SquareImageCropper v-model="form.coverUrl" title="裁剪团队封面" remove-text="移除封面"
              success-message="团队封面已上传" file-prefix="team-cover" preview-alt="团队封面预览" />
          </el-form-item>
          <el-form-item label="公开度">
            <el-radio-group v-model="form.visibility">
              <el-radio :value="0">公开（任何人可加入）</el-radio>
              <el-radio :value="1">非公开（需邀请码）</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item v-if="form.visibility === 1" label="邀请码">
            <el-input v-model="form.inviteCode" maxlength="32" placeholder="非公开团队必填" />
          </el-form-item>
          <el-form-item label="团队简介">
            <MdEditor v-model="form.description" height="220px" editor-id="team-create-editor" />
          </el-form-item>
        </section>
      </el-form>

      <div class="form-actions">
        <el-button @click="cancel">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submit">创建</el-button>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { ArrowLeft } from "@element-plus/icons-vue";
import { createTeam } from "@/api/team";
import SquareImageCropper from "@/components/upload/SquareImageCropper.vue";

const router = useRouter();
const creating = ref(false);
const form = ref(blankForm());

function blankForm() {
  return { name: "", coverUrl: "", description: "", visibility: 0, inviteCode: "" };
}

const cancel = () => router.push({ name: "TeamList" });

const submit = async () => {
  if (!form.value.name.trim()) {
    ElMessage.warning("请填写团队名称");
    return;
  }
  if (form.value.visibility === 1 && !form.value.inviteCode.trim()) {
    ElMessage.warning("非公开团队必须设置邀请码");
    return;
  }
  creating.value = true;
  try {
    const res = await createTeam({
      ...form.value,
      name: form.value.name.trim(),
      inviteCode: form.value.inviteCode.trim(),
    });
    ElMessage.success("创建成功");
    router.push(`/team/${res.data.id}`);
  } catch (err) {
    console.error(err);
  } finally {
    creating.value = false;
  }
};
</script>

<style scoped>
.team-create-page {
  padding-top: 8px;
  padding-bottom: 28px;
}

.create-panel {
  padding: 24px 28px 28px;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}

.form-section + .form-section {
  margin-top: 28px;
}

.form-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-top: 24px;
}

.form-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

@media (max-width: 768px) {
  .create-panel {
    padding: 20px 16px;
  }
}
</style>
