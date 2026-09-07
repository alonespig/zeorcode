import { cpp } from "@codemirror/lang-cpp";
import { java } from "@codemirror/lang-java";
import { python } from "@codemirror/lang-python";

const normalizeLanguage = (name) => String(name || "").trim().toLowerCase();

// 后端 language.name 是编辑器模式的唯一来源；未知语言回退为纯文本。
export function getEditorLanguageExtension(name) {
  const language = normalizeLanguage(name);
  if (["c++", "cpp", "cc", "cxx", "g++"].includes(language)) return cpp();
  if (language === "java") return java();
  if (["python", "python3", "py"].includes(language)) return python();
  return null;
}
