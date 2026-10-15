import { computed, onBeforeUnmount, ref, unref, watch } from "vue";
import { ElMessageBox } from "element-plus";

const DRAFT_PREFIX = "submission-code:v2:";
const LAST_LANGUAGE_KEY = "submission-editor-last-language";
const DRAFT_TTL_MS = 30 * 24 * 60 * 60 * 1000;
const SAVE_DELAY_MS = 300;
let expiredDraftsPruned = false;

const isExpired = (updatedAt) => !updatedAt || Date.now() - updatedAt > DRAFT_TTL_MS;

const readDraft = (key) => {
  const raw = localStorage.getItem(key);
  if (raw === null) return null;
  try {
    const draft = JSON.parse(raw);
    if (typeof draft.code !== "string" || isExpired(Number(draft.updatedAt))) {
      localStorage.removeItem(key);
      return null;
    }
    return draft.code;
  } catch {
    localStorage.removeItem(key);
    return null;
  }
};

const pruneExpiredDrafts = () => {
  if (expiredDraftsPruned) return;
  expiredDraftsPruned = true;
  for (let index = localStorage.length - 1; index >= 0; index -= 1) {
    const key = localStorage.key(index);
    if (key?.startsWith(DRAFT_PREFIX)) readDraft(key);
  }
};

export function useSubmissionEditor({ contextKey, languages }) {
  const language = ref(null);
  const code = ref("");
  const activeContext = ref("");
  let hydrating = false;
  let saveTimer = null;

  pruneExpiredDrafts();

  const languageItems = () => unref(languages) || [];
  const normalizedContext = () => String(unref(contextKey) || "");
  const draftKey = (context, languageID) => `${DRAFT_PREFIX}${context}:${languageID}`;
  const templateFor = (languageID) =>
    languageItems().find((item) => Number(item.id) === Number(languageID))?.defaultCode || "";

  const writeDraft = (context, languageID, value) => {
    if (!context || !languageID) return;
    localStorage.setItem(
      draftKey(context, languageID),
      JSON.stringify({ code: value, updatedAt: Date.now() })
    );
  };

  const flushDraft = () => {
    if (saveTimer) {
      clearTimeout(saveTimer);
      saveTimer = null;
    }
    if (!hydrating) writeDraft(activeContext.value, language.value, code.value);
  };

  const loadCode = (context, languageID) => {
    const saved = readDraft(draftKey(context, languageID));
    hydrating = true;
    code.value = saved === null ? templateFor(languageID) : saved;
    hydrating = false;
  };

  const hydrate = () => {
    const context = normalizedContext();
    const items = languageItems();
    if (!context || !items.length) return;

    const available = new Set(items.map((item) => Number(item.id)));
    const remembered = Number(localStorage.getItem(LAST_LANGUAGE_KEY));
    const selected = available.has(Number(language.value))
      ? Number(language.value)
      : available.has(remembered)
        ? remembered
        : Number(items[0].id);

    if (activeContext.value !== context) {
      if (activeContext.value) flushDraft();
      activeContext.value = context;
      hydrating = true;
      language.value = selected;
      hydrating = false;
      loadCode(context, selected);
      return;
    }

    if (!available.has(Number(language.value))) language.value = selected;
  };

  watch(
    language,
    (next, previous) => {
      if (hydrating || !next) return;
      if (previous) {
        if (saveTimer) clearTimeout(saveTimer);
        saveTimer = null;
        writeDraft(activeContext.value, previous, code.value);
      }
      localStorage.setItem(LAST_LANGUAGE_KEY, String(next));
      loadCode(activeContext.value, next);
    },
    { flush: "sync" }
  );

  watch(
    code,
    (value) => {
      if (hydrating || !activeContext.value || !language.value) return;
      if (saveTimer) clearTimeout(saveTimer);
      saveTimer = setTimeout(() => {
        writeDraft(activeContext.value, language.value, value);
        saveTimer = null;
      }, SAVE_DELAY_MS);
    },
    { flush: "sync" }
  );

  watch(
    [
      () => normalizedContext(),
      () => languageItems().map((item) => `${item.id}:${item.defaultCode || ""}`).join("\u0000"),
    ],
    hydrate,
    { immediate: true, flush: "post" }
  );

  const hasTemplate = computed(() => Boolean(templateFor(language.value)));

  const restoreTemplate = async () => {
    const template = templateFor(language.value);
    if (!template) return;
    try {
      await ElMessageBox.confirm("当前编辑器内容将被该语言的默认模板覆盖，是否继续？", "恢复模板", {
        confirmButtonText: "确认恢复",
        cancelButtonText: "取消",
        type: "warning",
        lockScroll: false,
      });
    } catch {
      return;
    }
    code.value = template;
  };

  onBeforeUnmount(flushDraft);

  return { language, code, hasTemplate, restoreTemplate };
}
