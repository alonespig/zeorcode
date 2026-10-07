<script setup lang="ts">
import { reactive, ref } from "vue";
import { ElMessageBox, type FormInstance, type FormRules } from "element-plus";
import { message } from "@/utils/message";
import { broadcastNotification } from "@/api/admin/notifications";

defineOptions({ name: "AdminNotifications" });

const formRef = ref<FormInstance>();
const submitting = ref(false);
const form = reactive({ title: "", content: "", link: "" });

const rules: FormRules = {
  title: [{ required: true, message: "请输入通知标题", trigger: "blur" }],
  content: [{ required: true, message: "请输入通知内容", trigger: "blur" }]
};

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid || submitting.value) return;

  const title = form.title.trim();
  const content = form.content.trim();
  try {
    await ElMessageBox.confirm(
      `确认向全部用户发送系统通知「${title}」？发送后当前系统不支持撤回。`,
      "发布系统通知",
      {
        type: "warning",
        confirmButtonText: "确认发布"
      }
    );
  } catch {
    return;
  }

  submitting.value = true;
  try {
    await broadcastNotification({
      title,
      content,
      link: form.link.trim()
    });
    message("系统通知已发布", { type: "success" });
    form.title = "";
    form.content = "";
    form.link = "";
    formRef.value?.clearValidate();
  } catch {
    // 错误信息已由 http 拦截器统一提示
  } finally {
    submitting.value = false;
  }
}
</script>

<template>
  <div class="notification-page">
    <header class="page-header">
      <div>
        <h2>发布系统通知</h2>
        <p>通知将发送到所有用户的站内消息中。</p>
      </div>
    </header>

    <div class="content-layout">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="publish-form"
      >
        <el-form-item label="通知标题" prop="title">
          <el-input
            v-model="form.title"
            maxlength="100"
            show-word-limit
            placeholder="请输入通知标题"
          />
        </el-form-item>
        <el-form-item label="通知内容" prop="content">
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="10"
            maxlength="2000"
            show-word-limit
            placeholder="请输入通知内容"
          />
        </el-form-item>
        <el-form-item label="跳转链接">
          <el-input v-model="form.link" placeholder="可选，例如 /contest/123" />
        </el-form-item>
        <div class="form-actions">
          <el-button type="primary" :loading="submitting" @click="submit">
            发布通知
          </el-button>
        </div>
      </el-form>

      <aside class="preview-panel">
        <h3>通知预览</h3>
        <div class="preview-card">
          <div class="preview-type">系统通知</div>
          <strong>{{ form.title.trim() || "通知标题" }}</strong>
          <p>{{ form.content.trim() || "通知内容将在这里显示。" }}</p>
          <span v-if="form.link.trim()" class="preview-link"> 查看详情 → </span>
        </div>
        <el-alert
          title="发布会为每位用户创建一条通知，当前接口不支持撤回或删除。"
          type="warning"
          show-icon
          :closable="false"
        />
      </aside>
    </div>
  </div>
</template>

<style scoped>
.notification-page {
  padding: 24px;
  background: var(--el-bg-color);
  border-radius: 4px;
}

.page-header {
  padding-bottom: 20px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.page-header h2,
.preview-panel h3 {
  margin: 0;
}

.page-header p {
  margin: 6px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.content-layout {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(280px, 1fr);
  gap: 28px;
  padding-top: 24px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.preview-panel h3 {
  margin-bottom: 14px;
  font-size: 16px;
}

.preview-card {
  padding: 18px;
  margin-bottom: 16px;
  border: 1px solid var(--el-border-color-lighter);
  border-left: 3px solid var(--el-color-primary);
}

.preview-type,
.preview-link {
  color: var(--el-color-primary);
  font-size: 12px;
}

.preview-card strong {
  display: block;
  margin-top: 8px;
}

.preview-card p {
  min-height: 72px;
  margin: 10px 0;
  color: var(--el-text-color-regular);
  line-height: 1.7;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

@media (max-width: 900px) {
  .content-layout {
    grid-template-columns: 1fr;
  }
}
</style>
