import { shallowRef } from "vue";
import { getLanguages } from "@/api/languages";

const cachedLanguages = shallowRef([]);
let loadingPromise = null;

export function invalidateLanguageCache() {
  cachedLanguages.value = [];
}

export function useLanguages() {
  const languageList = cachedLanguages;

  const loadLanguages = async ({ force = false } = {}) => {
    if (!force && cachedLanguages.value.length > 0) {
      return cachedLanguages.value;
    }

    if (!loadingPromise) {
      loadingPromise = getLanguages()
        .then((res) => {
          cachedLanguages.value = res.data || [];
          return cachedLanguages.value;
        })
        .finally(() => {
          loadingPromise = null;
        });
    }

    return loadingPromise;
  };

  return {
    languageList,
    loadLanguages,
  };
}
