import DOMPurify from "dompurify";
import MarkdownIt from "markdown-it";
import texmath from "markdown-it-texmath";
import katex from "katex";
import hljs from "highlight.js/lib/common";
import "katex/dist/katex.min.css";
import "highlight.js/styles/github.css";

const md = new MarkdownIt({
  html: false,
  breaks: true,
  highlight(code, language) {
    if (language && hljs.getLanguage(language)) {
      return `<pre class="hljs"><code>${hljs.highlight(code, { language }).value}</code></pre>`;
    }
    return `<pre class="hljs"><code>${md.utils.escapeHtml(code)}</code></pre>`;
  }
});

md.use(texmath, {
  engine: katex,
  delimiters: "dollars",
  katexOptions: { throwOnError: false }
});

// 行内公式 $ x $ → $x$：texmath 的 dollars 规则不认紧贴 $ 的空格，
// 而 HDU / AtCoder 等源站用 MathJax(容忍空格)。不归一化会导致该公式失效，
// 且落单的 $ 还会和后面的公式错误配对、级联带歪。$$display$$ 不受影响。
const normalizeInlineMath = (s: string) =>
  s.replace(/\$([^$\n]+?)\$/g, (_, inner: string) => `$${inner.trim()}$`);

export function renderMarkdown(raw?: string | null): string {
  const text = normalizeInlineMath(String(raw || "").replace(/\r\n?/g, "\n"));
  return DOMPurify.sanitize(md.render(text));
}
