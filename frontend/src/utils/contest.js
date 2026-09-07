export function contestProblemLabel(index) {
  if (!Number.isInteger(index) || index < 0) return "";
  let value = index;
  let label = "";
  while (value >= 0) {
    label = String.fromCharCode(65 + (value % 26)) + label;
    value = Math.floor(value / 26) - 1;
  }
  return label;
}
