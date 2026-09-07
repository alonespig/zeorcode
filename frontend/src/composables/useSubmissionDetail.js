import { onScopeDispose, readonly, shallowRef, toValue, watch } from "vue";
import { getSubmitDetail } from "@/api/user";
import { JUDGE_STATUS } from "@/constants/index";

const POLL_INTERVAL = 1500;

const emptyDetail = () => ({
  user: {},
  problem: {},
  submission: { status: JUDGE_STATUS.PENDING },
  caseResults: [],
});

const isProcessing = (status) =>
  status === JUDGE_STATUS.PENDING || status === JUDGE_STATUS.JUDGING;

export function useSubmissionDetail(submissionId) {
  const detail = shallowRef(emptyDetail());
  const loading = shallowRef(false);
  const loadError = shallowRef(null);
  let pollTimer = null;
  let requestSerial = 0;

  function stopPolling() {
    if (pollTimer !== null) {
      window.clearTimeout(pollTimer);
      pollTimer = null;
    }
  }

  async function load({ showLoading = false } = {}) {
    const id = toValue(submissionId);
    if (!id) return null;

    const serial = ++requestSerial;
    if (showLoading) loading.value = true;
    try {
      const res = await getSubmitDetail(id);
      if (serial !== requestSerial) return null;
      detail.value = {
        ...emptyDetail(),
        ...res.data,
        submission: res.data?.submission || {},
        caseResults: res.data?.caseResults || [],
      };
      loadError.value = null;
      return detail.value.submission.status;
    } catch (error) {
      if (serial === requestSerial) loadError.value = error;
      return null;
    } finally {
      if (serial === requestSerial && showLoading) loading.value = false;
    }
  }

  function schedulePolling() {
    stopPolling();
    pollTimer = window.setTimeout(async () => {
      pollTimer = null;
      const status = await load();
      if (isProcessing(status)) schedulePolling();
    }, POLL_INTERVAL);
  }

  watch(
    () => toValue(submissionId),
    async (_id, _previousId, onCleanup) => {
      stopPolling();
      onCleanup(stopPolling);
      const status = await load({ showLoading: true });
      if (isProcessing(status)) schedulePolling();
    },
    { immediate: true },
  );

  onScopeDispose(stopPolling);

  return {
    detail: readonly(detail),
    loading: readonly(loading),
    loadError: readonly(loadError),
    refresh: () => load({ showLoading: true }),
  };
}
