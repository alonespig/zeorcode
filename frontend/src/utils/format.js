
// 保留最多 1 位小数，并去掉多余的尾零：26.0 → "26"，1.50 → "1.5"
const trim = (n) => Number(n.toFixed(1)).toString();

export const formatTime = (time) => {
  if (time == null || isNaN(time)) return "-";
  if (time >= 1000) {
    return trim(time / 1000) + " s";
  }
  return trim(time) + " ms";
};

export const formatMemory = (memory) => {
  if (memory == null || isNaN(memory)) return "-";
  if (memory >= 1024) {
    return trim(memory / 1024) + " MB";
  }
  return trim(memory) + " KB";
};
