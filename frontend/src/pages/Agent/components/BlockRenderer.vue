<template>
  <div class="block-list">
    <template v-for="(block, index) in blocks" :key="`${block.type}-${block.requestId || index}`">
      <ChoiceBlock
        v-if="block.type === 'single_select' || block.type === 'multi_select'"
        :block="block"
        :disabled="block.requestId !== pendingRequestId"
        :submitting="running"
        @submit="(requestId, value) => $emit('answer', requestId, value)"
      />
      <HomeworkSettingsBlock
        v-else-if="block.type === 'homework_settings'"
        :block="block"
        :disabled="block.requestId !== pendingRequestId"
        :submitting="running"
        @submit="(requestId, value) => $emit('answer', requestId, value)"
      />
      <DataTableBlock v-else-if="block.type === 'data_table'" :block="block" />
      <NoticeBlock
        v-else-if="block.type === 'notice'"
        :block="block"
        :disabled="block.requestId !== pendingRequestId"
        :submitting="running"
        @submit="(requestId, value) => $emit('answer', requestId, value)"
      />
      <ActionPreviewBlock
        v-else-if="block.type === 'action_preview'"
        :block="block"
        :disabled="block.payload?.actionId !== pendingActionId"
        :submitting="running"
        @approve="(actionId, version) => $emit('approve', actionId, version)"
        @reject="(actionId, version) => $emit('reject', actionId, version)"
      />
      <ActionResultBlock v-else-if="block.type === 'action_result'" :block="block" />
    </template>
  </div>
</template>

<script setup>
import ChoiceBlock from './blocks/ChoiceBlock.vue'
import HomeworkSettingsBlock from './blocks/HomeworkSettingsBlock.vue'
import DataTableBlock from './blocks/DataTableBlock.vue'
import NoticeBlock from './blocks/NoticeBlock.vue'
import ActionPreviewBlock from './blocks/ActionPreviewBlock.vue'
import ActionResultBlock from './blocks/ActionResultBlock.vue'

defineProps({
  blocks: { type: Array, default: () => [] },
  pendingRequestId: { type: Number, default: 0 },
  pendingActionId: { type: Number, default: 0 },
  running: Boolean,
})
defineEmits(['answer', 'approve', 'reject'])
</script>
