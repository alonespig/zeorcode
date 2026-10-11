<script setup>
import { computed, onBeforeUnmount, onMounted, shallowRef } from "vue";
import { useRoute, useRouter } from "vue-router";
import AgentView from "./AgentView.vue";
import AgentErrorBoundary from "./components/AgentErrorBoundary.vue";

const route = useRoute();
const router = useRouter();
const agentViewKey = shallowRef(0);
const startNewConversation = computed(() => route.query.new === "1");

const returnPath = computed(() => {
  const source = Array.isArray(route.query.from) ? route.query.from[0] : route.query.from;
  if (typeof source !== "string" || !source.startsWith("/") || source.startsWith("//")) {
    return "/home";
  }
  if (source.split("?")[0] === "/agent") return "/home";
  return source;
});

const returnLabel = computed(() => "返回原页面");

const goBack = () => {
  router.push(returnPath.value);
};

const retryAgent = () => {
  agentViewKey.value += 1;
};

const consumeNewConversationFlag = () => {
  const query = { ...route.query };
  delete query.new;
  router.replace({ query });
};

const pageLockClass = "agent-route-active";

onMounted(() => {
  document.documentElement.classList.add(pageLockClass);
});

onBeforeUnmount(() => {
  document.documentElement.classList.remove(pageLockClass);
});
</script>

<template>
  <section class="agent-route">
    <AgentErrorBoundary :key="agentViewKey" @retry="retryAgent">
      <AgentView
        :key="agentViewKey"
        :return-label="returnLabel"
        :start-new-conversation="startNewConversation"
        @back="goBack"
        @new-conversation-started="consumeNewConversationFlag"
      />
    </AgentErrorBoundary>
  </section>
</template>

<style scoped>
.agent-route {
  position: fixed;
  inset: 0;
  width: 100%;
  height: 100dvh;
  overflow: hidden;
  background: #fff;
  box-sizing: border-box;
}

:global(html.agent-route-active),
:global(html.agent-route-active body),
:global(html.agent-route-active #app) {
  height: 100%;
  overflow: hidden;
}
</style>
