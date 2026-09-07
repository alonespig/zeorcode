export const CONTEST_DRAFT_KEY = "create-contest-draft";

// 气球预设色板：创建/编辑比赛时按顺序自动分配，保证每题默认不同色（可手动改）
export const BALLOON_COLORS = [
  "#e74c3c",
  "#3498db",
  "#2ecc71",
  "#f1c40f",
  "#9b59b6",
  "#e67e22",
  "#1abc9c",
  "#ff7faa",
  "#34495e",
  "#16a085",
  "#d35400",
  "#8e44ad",
];

export const CONTEST_TYPE = {
  ACM: 1,
  OI: 2,
  IOI: 3,
  CF: 4,
};

// 赛制选项（value 与后端 consts.ContestType 对齐：1 ACM / 2 OI / 3 IOI / 4 CF 动态分）
export const RULE_OPTIONS = [
  { value: CONTEST_TYPE.ACM, label: "ACM" },
  { value: CONTEST_TYPE.OI, label: "OI" },
  { value: CONTEST_TYPE.IOI, label: "IOI" },
  { value: CONTEST_TYPE.CF, label: "CF" },
];

export const CONTEST_STATUS = {
  NOT_STARTED: 0,
  RUNNING: 1,
  ENDED: 2,
};

export const CONTEST_STATUS_TEXT = {
  [CONTEST_STATUS.NOT_STARTED]: "未开始",
  [CONTEST_STATUS.RUNNING]: "进行中",
  [CONTEST_STATUS.ENDED]: "已结束",
};

export const ruleLabelOf = (value) => RULE_OPTIONS.find((o) => o.value === value)?.label || "-";

// 竞赛 rating 段位（CF 风）：从高到低匹配
export const RATING_TIERS = [
  { min: 2400, name: "Grandmaster", color: "#ff0000" },
  { min: 2100, name: "Master", color: "#ff8c00" },
  { min: 1900, name: "Candidate Master", color: "#aa00aa" },
  { min: 1600, name: "Expert", color: "#0000ff" },
  { min: 1400, name: "Specialist", color: "#03a89e" },
  { min: 1200, name: "Pupil", color: "#008000" },
  { min: 0, name: "Newbie", color: "#808080" },
];

export const ratingTier = (rating) => {
  // 0 / 空 = 未参加过 rated 比赛，显示未定级灰
  if (!rating || rating <= 0) return { name: "未定级", color: "#9aa0a8" };
  return RATING_TIERS.find((t) => rating >= t.min) || RATING_TIERS[RATING_TIERS.length - 1];
};

export const PendingCode = 0;
export const AcceptedCode = 1;
export const MemoryLimitExceededCode = 2;
export const TimeLimitExceededCode = 3;
export const RuntimeErrorCode = 4;
export const WrongAnswerCode = 5;
export const CompileErrorCode = 6;
export const UnknownErrorCode = 7;

// export const formatTime = (time) => {
//   if (time == null || isNaN(time)) return "-";
//   if (time >= 1000) {
//     return (time / 1000).toFixed(1) + "s";
//   }
//   return time.toFixed(1) + "ms";
// };

// export const formatMemory = (memory) => {
//   if (memory == null || isNaN(memory)) return "-";
//   if (memory >= 1024) {
//     return (memory / 1024).toFixed(1) + "MB";
//   }
//   return memory.toFixed(1) + "KB";
// };

export const JUDGE_STATUS = {
  PENDING: 0,
  ACCEPTED: 1,
  MEMORY_LIMIT_EXCEEDED: 2,
  TIME_LIMIT_EXCEEDED: 3,
  RUNTIME_ERROR: 4,
  WRONG_ANSWER: 5,
  COMPILE_ERROR: 6,
  UNKNOWN: 7,
  PRESENTATION_ERROR: 8,
  SYSTEM_ERROR: 9,
  JUDGING: 10,
};

export const JUDGE_STATUS_TEXT = {
  [JUDGE_STATUS.PENDING]: "Pending",
  [JUDGE_STATUS.JUDGING]: "Judging",
  [JUDGE_STATUS.ACCEPTED]: "Accepted",
  [JUDGE_STATUS.WRONG_ANSWER]: "Wrong Answer",
  [JUDGE_STATUS.TIME_LIMIT_EXCEEDED]: "Time Limit Exceeded",
  [JUDGE_STATUS.MEMORY_LIMIT_EXCEEDED]: "Memory Limit Exceeded",
  [JUDGE_STATUS.RUNTIME_ERROR]: "Runtime Error",
  [JUDGE_STATUS.COMPILE_ERROR]: "Compilation Error",
  [JUDGE_STATUS.PRESENTATION_ERROR]: "Presentation Error",
  [JUDGE_STATUS.SYSTEM_ERROR]: "System Error",
  [JUDGE_STATUS.UNKNOWN]: "Unknown",
};

export const JUDGE_STATUS_CLASS = {
  [JUDGE_STATUS.PENDING]: "pending",
  [JUDGE_STATUS.JUDGING]: "judging",
  [JUDGE_STATUS.ACCEPTED]: "accepted",
  [JUDGE_STATUS.WRONG_ANSWER]: "wrong-answer",
  [JUDGE_STATUS.TIME_LIMIT_EXCEEDED]: "time-limit",
  [JUDGE_STATUS.MEMORY_LIMIT_EXCEEDED]: "memory-limit",
  [JUDGE_STATUS.RUNTIME_ERROR]: "runtime-error",
  [JUDGE_STATUS.COMPILE_ERROR]: "compile-error",
  [JUDGE_STATUS.PRESENTATION_ERROR]: "presentation-error",
  [JUDGE_STATUS.SYSTEM_ERROR]: "system-error",
  [JUDGE_STATUS.UNKNOWN]: "unknown",
};

// export const categoryList = [
//   '博客', '通知', '题解', '求助',
// ]

export const categoryList = [
  {
    id: "announcement",
    name: "通知",
  },
  {
    id: "blog",
    name: "博客",
  },
  {
    id: "solution",
    name: "题解",
  },
  {
    id: "help",
    name: "讨论",
  },
];

export const categoryTitleMap = {
  blog: "博客",
  announcement: "通知",
  solution: "题解",
  help: "讨论",
};

export const categoryStyleMap = {
  solution: "border-orange-200 bg-orange-50 text-orange-700",
  help: "border-sky-200 bg-sky-50 text-sky-700",
  blog: "border-indigo-200 bg-indigo-50 text-indigo-700",
  announcement: "border-red-200 bg-red-50 text-red-700",
  default: "border-slate-200 bg-slate-50 text-slate-600",
};
