<script setup>
import {
  nextTick,
  onBeforeUnmount,
  reactive,
  shallowRef,
  useTemplateRef,
} from "vue";
import { Delete, Plus, RefreshLeft } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import { uploadImage } from "@/api/upload";

const imageUrl = defineModel({ type: String, default: "" });
const props = defineProps({
  title: { type: String, default: "裁剪图片" },
  emptyText: { type: String, default: "选择图片" },
  changeText: { type: String, default: "重新裁剪" },
  helpText: { type: String, default: "PNG/JPG，最大 5MB，保存为 1:1 正方形" },
  removeText: { type: String, default: "移除图片" },
  successMessage: { type: String, default: "图片已上传" },
  filePrefix: { type: String, default: "image" },
  previewAlt: { type: String, default: "图片预览" },
  previewSize: { type: Number, default: 112 },
  round: { type: Boolean, default: false },
  uploadRequest: { type: Function, default: uploadImage },
});

const OUTPUT_SIZE = 720;
const MAX_FILE_SIZE = 5 * 1024 * 1024;
const ALLOWED_IMAGE_TYPES = new Set(["image/jpeg", "image/png"]);

const cropCanvas = useTemplateRef("cropCanvas");
const dialogVisible = shallowRef(false);
const uploading = shallowRef(false);
const zoom = shallowRef(1);
const sourceImage = shallowRef(null);
const sourceObjectUrl = shallowRef("");
const offset = reactive({ x: 0, y: 0 });
const drag = reactive({ active: false, x: 0, y: 0 });

function clamp(value, min, max) {
  return Math.min(max, Math.max(min, value));
}

function imageMetrics() {
  const image = sourceImage.value;
  if (!image) return null;
  const baseScale = Math.max(OUTPUT_SIZE / image.naturalWidth, OUTPUT_SIZE / image.naturalHeight);
  const scale = baseScale * zoom.value;
  const width = image.naturalWidth * scale;
  const height = image.naturalHeight * scale;
  return {
    width,
    height,
    maxX: Math.max(0, (width - OUTPUT_SIZE) / 2),
    maxY: Math.max(0, (height - OUTPUT_SIZE) / 2),
  };
}

function renderCrop() {
  const canvas = cropCanvas.value;
  const image = sourceImage.value;
  const metrics = imageMetrics();
  if (!canvas || !image || !metrics) return;

  offset.x = clamp(offset.x, -metrics.maxX, metrics.maxX);
  offset.y = clamp(offset.y, -metrics.maxY, metrics.maxY);

  const context = canvas.getContext("2d");
  context.clearRect(0, 0, OUTPUT_SIZE, OUTPUT_SIZE);
  context.imageSmoothingEnabled = true;
  context.imageSmoothingQuality = "high";
  context.drawImage(
    image,
    (OUTPUT_SIZE - metrics.width) / 2 + offset.x,
    (OUTPUT_SIZE - metrics.height) / 2 + offset.y,
    metrics.width,
    metrics.height,
  );
}

function handleImageChange(uploadFile) {
  const file = uploadFile.raw;
  if (!file) return;
  if (!ALLOWED_IMAGE_TYPES.has(file.type)) {
    ElMessage.warning("请选择 PNG 或 JPG 图片");
    return;
  }
  if (file.size > MAX_FILE_SIZE) {
    ElMessage.warning("图片大小不能超过 5MB");
    return;
  }

  releaseSource();
  const objectUrl = URL.createObjectURL(file);
  sourceObjectUrl.value = objectUrl;
  const image = new Image();
  image.onload = async () => {
    sourceImage.value = image;
    zoom.value = 1;
    offset.x = 0;
    offset.y = 0;
    dialogVisible.value = true;
    await nextTick();
    renderCrop();
  };
  image.onerror = () => {
    ElMessage.error("图片读取失败，请重新选择");
    releaseSource();
  };
  image.src = objectUrl;
}

function pointerScale(event) {
  const rect = event.currentTarget.getBoundingClientRect();
  return OUTPUT_SIZE / rect.width;
}

function startDrag(event) {
  drag.active = true;
  drag.x = event.clientX;
  drag.y = event.clientY;
  event.currentTarget.setPointerCapture(event.pointerId);
}

function moveImage(event) {
  if (!drag.active) return;
  const scale = pointerScale(event);
  offset.x += (event.clientX - drag.x) * scale;
  offset.y += (event.clientY - drag.y) * scale;
  drag.x = event.clientX;
  drag.y = event.clientY;
  renderCrop();
}

function stopDrag(event) {
  if (!drag.active) return;
  drag.active = false;
  if (event.currentTarget.hasPointerCapture(event.pointerId)) {
    event.currentTarget.releasePointerCapture(event.pointerId);
  }
}

function resetCrop() {
  zoom.value = 1;
  offset.x = 0;
  offset.y = 0;
  renderCrop();
}

function canvasBlob() {
  return new Promise((resolve, reject) => {
    cropCanvas.value?.toBlob(
      (blob) => (blob ? resolve(blob) : reject(new Error("无法生成裁剪图片"))),
      "image/jpeg",
      0.9,
    );
  });
}

async function confirmCrop() {
  if (!sourceImage.value || !cropCanvas.value) return;
  uploading.value = true;
  try {
    renderCrop();
    const blob = await canvasBlob();
    const file = new File([blob], `${props.filePrefix}-${Date.now()}.jpg`, { type: "image/jpeg" });
    const formData = new FormData();
    formData.append("file", file);
    const response = await props.uploadRequest(formData);
    imageUrl.value = response.data.url;
    dialogVisible.value = false;
    ElMessage.success(props.successMessage);
  } catch (error) {
    console.error(error);
    if (error instanceof Error && error.message === "无法生成裁剪图片") {
      ElMessage.error("图片处理失败，请重新选择");
    }
  } finally {
    uploading.value = false;
  }
}

function removeImage() {
  imageUrl.value = "";
}

function releaseSource() {
  if (sourceObjectUrl.value) URL.revokeObjectURL(sourceObjectUrl.value);
  sourceObjectUrl.value = "";
  sourceImage.value = null;
  drag.active = false;
}

onBeforeUnmount(releaseSource);
</script>

<template>
  <div class="image-field">
    <el-upload
      class="image-upload"
      action="#"
      accept=".png,.jpg,.jpeg"
      :auto-upload="false"
      :show-file-list="false"
      :on-change="handleImageChange"
    >
      <div
        class="image-preview"
        :class="{ 'image-preview--filled': imageUrl, 'image-preview--round': round }"
        :style="{ width: `${previewSize}px`, height: `${previewSize}px` }"
      >
        <img v-if="imageUrl" :src="imageUrl" :alt="previewAlt" />
        <div v-else class="image-placeholder">
          <el-icon :size="24"><Plus /></el-icon>
          <span>{{ emptyText }}</span>
        </div>
        <span v-if="imageUrl" class="image-change">{{ changeText }}</span>
      </div>
    </el-upload>

    <div class="image-help" :style="{ minHeight: `${previewSize}px` }">
      <span>{{ helpText }}</span>
      <el-button v-if="imageUrl" type="danger" text :icon="Delete" @click="removeImage">
        {{ removeText }}
      </el-button>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="title"
      width="min(500px, calc(100vw - 32px))"
      :close-on-click-modal="false"
      append-to-body
      @closed="releaseSource"
    >
      <div class="crop-workspace">
        <canvas
          ref="cropCanvas"
          class="crop-canvas"
          :width="OUTPUT_SIZE"
          :height="OUTPUT_SIZE"
          @pointerdown="startDrag"
          @pointermove="moveImage"
          @pointerup="stopDrag"
          @pointercancel="stopDrag"
        ></canvas>
        <p class="crop-tip">拖动图片调整位置，使用滑块缩放</p>
        <div class="zoom-row">
          <span>缩放</span>
          <el-slider v-model="zoom" :min="1" :max="3" :step="0.01" :show-tooltip="false" @input="renderCrop" />
          <el-button text :icon="RefreshLeft" @click="resetCrop">重置</el-button>
        </div>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="confirmCrop">裁剪并上传</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.image-field { display: flex; align-items: flex-start; gap: 14px; }
.image-upload { flex: 0 0 auto; }
.image-preview { position: relative; overflow: hidden; border: 1px dashed #bdc9d8; border-radius: 4px; display: grid; place-items: center; color: #758196; background: #f8fafc; cursor: pointer; transition: border-color 150ms ease, color 150ms ease, background-color 150ms ease; }
.image-preview--round { border-radius: 50%; }
.image-preview:hover { border-color: #1769e0; color: #1769e0; background: #f3f7fd; }
.image-preview img { width: 100%; height: 100%; display: block; object-fit: cover; }
.image-placeholder { display: flex; align-items: center; flex-direction: column; gap: 7px; font-size: 13px; }
.image-change { position: absolute; right: 0; bottom: 0; left: 0; padding: 5px 0; color: #fff; background: rgba(18, 29, 45, 0.68); font-size: 12px; text-align: center; transform: translateY(100%); transition: transform 150ms ease; }
.image-preview--filled:hover .image-change { transform: translateY(0); }
.image-help { display: flex; align-items: flex-start; flex-direction: column; justify-content: space-between; color: #7b8797; font-size: 13px; line-height: 1.6; }
.crop-workspace { display: flex; align-items: center; flex-direction: column; }
.crop-canvas { width: min(360px, 100%); height: auto; aspect-ratio: 1; border: 1px solid #dce3ec; display: block; background: #edf1f5; cursor: grab; touch-action: none; }
.crop-canvas:active { cursor: grabbing; }
.crop-tip { margin: 10px 0 14px; color: #788496; font-size: 13px; }
.zoom-row { width: min(360px, 100%); display: grid; grid-template-columns: 36px 1fr 62px; align-items: center; gap: 10px; color: #566276; font-size: 14px; }
</style>
