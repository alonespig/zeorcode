import MarkdownIt from "markdown-it";
import texmath from "markdown-it-texmath";
import katex from "katex";
import hljs from "@/utils/highlight";
import DOMPurify from "dompurify";
import "katex/dist/katex.min.css";
import "highlight.js/styles/github.css";

const md = new MarkdownIt({
  html: false,
  breaks: true,
  highlight(code, lang) {
    if (lang && hljs.getLanguage(lang)) {
      return `<pre class="hljs"><code>${hljs.highlight(code, { language: lang }).value}</code></pre>`;
    }
    return `<pre class="hljs"><code>${md.utils.escapeHtml(code)}</code></pre>`;
  },
});

md.use(texmath, {
  engine: katex,
  delimiters: "dollars",
  katexOptions: { throwOnError: false },
});

// 行内公式 $ x $ → $x$：texmath 的 dollars 规则不认紧贴 $ 的空格，
// 而 HDU / AtCoder 等源站用 MathJax(容忍空格)。不归一化会导致该公式失效，
// 且落单的 $ 还会和后面的公式错误配对、级联带歪。$$display$$ 不受影响。
const normalizeInlineMath = (s) =>
  s.replace(/\$([^$\n]+?)\$/g, (_, inner) => `$${inner.trim()}$`);

export const renderMarkdown = (raw) => {
  const text = normalizeInlineMath(String(raw || "").replace(/\r\n?/g, "\n"));
  const html = md.render(text);
  return DOMPurify.sanitize(html);
};
