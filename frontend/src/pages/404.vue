<template>
  <main class="not-found" aria-labelledby="not-found-title">
    <div class="not-found__content">
      <div class="not-found__scene" aria-hidden="true">
        <svg class="not-found__trail" viewBox="0 0 360 200" fill="none" focusable="false">
          <path d="M42 137C2 49 100 15 191 24s178 54 132 124-181 35-169-4" stroke="#e2d5b8" stroke-width="1.5" stroke-dasharray="5 7" />
          <circle cx="49" cy="63" r="4" fill="#ffcb4c" />
          <circle cx="296" cy="158" r="3" fill="#ffcb4c" />
        </svg>
        <span class="not-found__symbol not-found__symbol--code">&lt;/&gt;</span>
        <span class="not-found__symbol not-found__symbol--question">?</span>
        <span class="not-found__symbol not-found__symbol--braces">{ }</span>
      <svg class="not-found__emoji" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 36 36" aria-hidden="true" focusable="false">
        <circle fill="#FFCB4C" cx="18" cy="17.018" r="17" />
        <path fill="#65471B" d="M14.524 21.036c-.145-.116-.258-.274-.312-.464-.134-.46.13-.918.59-1.021 4.528-1.021 7.577 1.363 7.706 1.465.384.306.459.845.173 1.205-.286.358-.828.401-1.211.097-.11-.084-2.523-1.923-6.182-1.098-.274.061-.554-.016-.764-.184z" />
        <ellipse fill="#65471B" cx="13.119" cy="11.174" rx="2.125" ry="2.656" />
        <ellipse fill="#65471B" cx="24.375" cy="12.236" rx="2.125" ry="2.656" />
        <path fill="#F19020" d="M17.276 35.149s1.265-.411 1.429-1.352c.173-.972-.624-1.167-.624-1.167s1.041-.208 1.172-1.376c.123-1.101-.861-1.363-.861-1.363s.97-.4 1.016-1.539c.038-.959-.995-1.428-.995-1.428s5.038-1.221 5.556-1.341c.516-.12 1.32-.615 1.069-1.694-.249-1.08-1.204-1.118-1.697-1.003-.494.115-6.744 1.566-8.9 2.068l-1.439.334c-.54.127-.785-.11-.404-.512.508-.536.833-1.129.946-2.113.119-1.035-.232-2.313-.433-2.809-.374-.921-1.005-1.649-1.734-1.899-1.137-.39-1.945.321-1.542 1.561.604 1.854.208 3.375-.833 4.293-2.449 2.157-3.588 3.695-2.83 6.973.828 3.575 4.377 5.876 7.952 5.048l3.152-.681z" />
        <path fill="#65471B" d="M9.296 6.351c-.164-.088-.303-.224-.391-.399-.216-.428-.04-.927.393-1.112 4.266-1.831 7.699-.043 7.843.034.433.231.608.747.391 1.154-.216.405-.74.546-1.173.318-.123-.063-2.832-1.432-6.278.047-.257.109-.547.085-.785-.042zm12.135 3.75c-.156-.098-.286-.243-.362-.424-.187-.442.023-.927.468-1.084 4.381-1.536 7.685.48 7.823.567.415.26.555.787.312 1.178-.242.39-.776.495-1.191.238-.12-.072-2.727-1.621-6.267-.379-.266.091-.553.046-.783-.096z" />
      </svg>
      </div>
      <h1 id="not-found-title" class="not-found__title"><span class="not-found__code">404</span> 资源不存在</h1>
      <p class="not-found__description">生活总归带点荒谬</p>
      <p class="not-found__hint">页面可能已被移动，或是地址写错了。<br />不如换个方向，找一道题试试？</p>
      <nav class="not-found__actions" aria-label="离开错误页面">
        <router-link to="/problem" class="not-found__button not-found__button--primary" aria-label="找点乐子吧，前往题库">
          找点乐子吧 <span aria-hidden="true">→</span>
        </router-link>
        <router-link to="/" class="not-found__button">返回首页</router-link>
      </nav>
    </div>
  </main>
</template>

<style scoped>
.not-found {
  --nf-ink: #292722;
  --nf-muted: #7b776e;
  --nf-accent: #ffcb4c;
  --nf-soft: #fff8e5;
  min-height: 100vh;
  min-height: 100svh;
  display: grid;
  place-items: center;
  padding: 40px 20px 100px;
  background: #fff;
  color: var(--nf-ink);
  text-align: center;
}

.not-found__content { width: 100%; max-width: 560px; }
.not-found__scene { position: relative; display: grid; place-items: center; width: min(360px, 100%); height: 200px; margin: 0 auto 28px; }
.not-found__scene::before { content: ""; position: absolute; width: 160px; height: 160px; border-radius: 50%; background: var(--nf-soft); }
.not-found__trail { position: absolute; inset: 0; width: 100%; height: 100%; }
.not-found__symbol { position: absolute; font-family: Consolas, monospace; font-weight: 600; line-height: 1; }
.not-found__symbol--code { left: 14%; top: 20%; font-size: 22px; color: #8b949f; transform: rotate(-12deg); }
.not-found__symbol--question { right: 12%; top: 13%; display: grid; place-items: center; width: 40px; height: 40px; border: 1px solid #ecd69b; border-radius: 10px 10px 10px 2px; background: #fffdf7; color: #a27417; font-size: 26px; transform: rotate(12deg); }
.not-found__symbol--braces { left: 16%; bottom: 14%; color: #baa779; font-size: 20px; transform: rotate(8deg); }

.not-found__emoji {
  position: relative;
  display: block;
  width: 108px;
  height: 108px;
  animation: thinking-float 5s ease-in-out infinite;
}

.not-found__title {
  margin: 0;
  font-size: clamp(28px, 5vw, 40px);
  font-weight: 500;
  line-height: 1.4;
}

.not-found__description {
  margin: 12px 0 0;
  font-size: 18px;
  line-height: 1.6;
}

.not-found__code { font-family: Consolas, monospace; font-weight: 600; }
.not-found__hint { margin: 18px 0 0; color: var(--nf-muted); font-size: 14px; line-height: 1.9; }
.not-found__actions { display: flex; flex-wrap: wrap; justify-content: center; gap: 12px; margin-top: 30px; }

.not-found__button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  min-height: 44px;
  padding: 10px 20px;
  border: 1px solid #e5e2da;
  border-radius: 6px;
  background: #fff;
  color: inherit;
  font-size: 15px;
  line-height: 1.5;
  transition: color 150ms ease, border-color 150ms ease, background-color 150ms ease;
}

.not-found__button:hover {
  border-color: #dbbd70;
  background: var(--nf-soft);
}

.not-found__button--primary { border-color: var(--nf-accent); background: var(--nf-accent); font-weight: 500; }
.not-found__button--primary:hover { border-color: #f4bd34; background: #f4bd34; }

.not-found__button:focus-visible {
  outline: 2px solid #9b6a0b;
  outline-offset: 4px;
}

@media (max-width: 480px) {
  .not-found { padding-bottom: 64px; }
  .not-found__emoji { width: 96px; height: 96px; }
  .not-found__description { font-size: 16px; }
}

@keyframes thinking-float {
  0%, 100% { transform: translateY(0) rotate(-3deg); }
  50% { transform: translateY(-7px) rotate(3deg); }
}

@media (prefers-reduced-motion: reduce) {
  .not-found__emoji { animation: none; }
  .not-found__button { transition: none; }
}
</style>
