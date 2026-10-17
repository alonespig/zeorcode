<template>
  <div class="judging-animation" role="status" aria-live="polite" aria-label="正在评测">
    <svg class="judging-scene" viewBox="0 0 420 150" aria-hidden="true">
      <defs>
        <linearGradient id="owl-feather" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0" stop-color="#9b7555" />
          <stop offset="1" stop-color="#76523a" />
        </linearGradient>
        <linearGradient id="owl-belly" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stop-color="#f2dcc0" />
          <stop offset="1" stop-color="#dfbd92" />
        </linearGradient>
        <filter id="owl-shadow" x="-40%" y="-40%" width="180%" height="200%">
          <feDropShadow dx="0" dy="4" stdDeviation="4" flood-color="#24364d" flood-opacity=".14" />
        </filter>
      </defs>

      <path class="flight-line" d="M32 78 C85 58 120 62 151 78" />
      <g class="code-card code-card--one" transform="translate(50 55)">
        <rect width="40" height="24" rx="5" />
        <text x="20" y="16">&lt;/&gt;</text>
      </g>
      <g class="code-card code-card--two" transform="translate(102 69)">
        <rect width="34" height="22" rx="5" />
        <text x="17" y="15">01</text>
      </g>

      <g transform="translate(210 57)">
        <g class="owl" filter="url(#owl-shadow)">
          <path class="owl-ear" d="M-31-24L-20-39L-10-27M31-24L20-39L10-27" />
          <ellipse class="owl-body" cy="1" rx="34" ry="39" />
          <ellipse class="owl-belly" cy="9" rx="24" ry="27" />
          <g class="owl-eyes">
            <circle class="owl-eye-ring" cx="-13" cy="-8" r="11" />
            <circle class="owl-eye-ring" cx="13" cy="-8" r="11" />
            <circle class="owl-eye" cx="-13" cy="-8" r="4" />
            <circle class="owl-eye" cx="13" cy="-8" r="4" />
            <circle class="owl-eye-glint" cx="-11.5" cy="-9.5" r="1" />
            <circle class="owl-eye-glint" cx="14.5" cy="-9.5" r="1" />
          </g>
          <path class="owl-beak" d="M-4 1H4L0 8Z" />
          <path class="owl-wing owl-wing--left" d="M-29 4Q-47 17-31 29Q-21 24-17 11Z" />
          <path class="owl-wing owl-wing--right" d="M29 4Q47 17 31 29Q21 24 17 11Z" />
        </g>
      </g>

      <path class="branch" d="M137 94 Q174 101 211 95 Q248 89 285 96" />
      <path class="branch-tip" d="M269 93 Q286 83 300 87" />

      <g class="test-lamp test-lamp--one" transform="translate(171 97)">
        <path class="lamp-cord" d="M0 0V12" />
        <circle class="lamp-glow" cy="23" r="10" />
        <circle class="lamp-body" cy="23" r="7" />
        <text y="26">1</text>
      </g>
      <g class="test-lamp test-lamp--two" transform="translate(211 96)">
        <path class="lamp-cord" d="M0 0V17" />
        <circle class="lamp-glow" cy="28" r="10" />
        <circle class="lamp-body" cy="28" r="7" />
        <text y="31">2</text>
      </g>
      <g class="test-lamp test-lamp--three" transform="translate(251 94)">
        <path class="lamp-cord" d="M0 0V12" />
        <circle class="lamp-glow" cy="23" r="10" />
        <circle class="lamp-body" cy="23" r="7" />
        <text y="26">3</text>
      </g>

      <path class="moon" d="M337 30a14 14 0 1 0 11 23a17 17 0 1 1-11-23Z" />
      <circle class="star star--one" cx="315" cy="51" r="2" />
      <circle class="star star--two" cx="365" cy="67" r="1.6" />
      <circle class="star star--three" cx="302" cy="29" r="1.5" />
    </svg>
    <span class="sr-only">正在评测，测试点结果将在完成后显示。</span>
  </div>
</template>

<style scoped>
.judging-animation {
  min-height: 144px;
  display: grid;
  place-items: center;
  overflow: hidden;
  background:
    radial-gradient(circle at 51% 46%, rgba(64, 158, 255, .075), transparent 34%),
    #fff;
}

.judging-scene { width: min(420px, 88%); height: 150px; overflow: visible; }
.flight-line { fill: none; stroke: #d9e6f4; stroke-width: 2; stroke-linecap: round; stroke-dasharray: 6 7; animation: flight-flow 1.5s linear infinite; }
.code-card { color: #617a98; animation: card-drift 2.8s ease-in-out infinite; }
.code-card--two { animation-delay: -1.4s; }
.code-card rect { fill: #f6f9fc; stroke: #d6e1ed; }
.code-card text { fill: currentColor; font-family: "Cascadia Code", Consolas, monospace; font-size: 10px; text-anchor: middle; }
.owl { transform-box: fill-box; transform-origin: center bottom; animation: owl-watch 2.4s ease-in-out infinite; }
.owl-ear, .owl-body { fill: url(#owl-feather); }
.owl-belly { fill: url(#owl-belly); }
.owl-eye-ring { fill: #fff; }
.owl-eyes { transform-box: fill-box; transform-origin: center; animation: owl-blink 3.6s ease-in-out infinite; }
.owl-eye { fill: #263446; }
.owl-eye-glint { fill: #fff; }
.owl-beak { fill: #f29a3d; }
.owl-wing { fill: #704b34; transform-box: fill-box; }
.owl-wing--left { transform-origin: right top; animation: wing-left 1.2s ease-in-out infinite alternate; }
.owl-wing--right { transform-origin: left top; animation: wing-right 1.2s ease-in-out infinite alternate; }
.branch, .branch-tip { fill: none; stroke: #876247; stroke-width: 5; stroke-linecap: round; }
.branch-tip { stroke-width: 3; }
.lamp-cord { fill: none; stroke: #b8c5d3; stroke-width: 1.5; }
.test-lamp text { fill: #fff; font-family: Consolas, monospace; font-size: 8px; text-anchor: middle; }
.lamp-body { fill: #d8e2ed; }
.lamp-glow { fill: #3e92ec; opacity: 0; transform-box: fill-box; transform-origin: center; }
.test-lamp .lamp-body, .test-lamp .lamp-glow { animation: lamp-light 2.7s ease-in-out infinite; }
.test-lamp .lamp-glow { animation-name: lamp-glow; }
.test-lamp--two .lamp-body, .test-lamp--two .lamp-glow { animation-delay: .55s; }
.test-lamp--three .lamp-body, .test-lamp--three .lamp-glow { animation-delay: 1.1s; }
.moon { fill: #ffd369; opacity: .9; }
.star { fill: #78a9df; animation: star-twinkle 2s ease-in-out infinite; }
.star--two { animation-delay: -.65s; }
.star--three { animation-delay: -1.3s; }
.sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip: rect(0, 0, 0, 0); white-space: nowrap; border: 0; }

@keyframes flight-flow { to { stroke-dashoffset: -39; } }
@keyframes card-drift { 0%, 100% { opacity: .45; transform: translateY(3px); } 50% { opacity: 1; transform: translateY(-3px); } }
@keyframes owl-watch { 0%, 100% { transform: translateY(2px) rotate(-1deg); } 50% { transform: translateY(-3px) rotate(1deg); } }
@keyframes owl-blink { 0%, 43%, 48%, 100% { transform: scaleY(1); } 45%, 46% { transform: scaleY(.12); } }
@keyframes wing-left { from { transform: rotate(-3deg); } to { transform: rotate(6deg); } }
@keyframes wing-right { from { transform: rotate(3deg); } to { transform: rotate(-6deg); } }
@keyframes lamp-light { 0%, 24%, 76%, 100% { fill: #d8e2ed; } 36%, 64% { fill: #2faf76; } }
@keyframes lamp-glow { 0%, 26%, 76%, 100% { opacity: 0; transform: scale(.75); } 40%, 62% { opacity: .18; transform: scale(1.55); } }
@keyframes star-twinkle { 0%, 100% { opacity: .25; transform: scale(.75); } 50% { opacity: 1; transform: scale(1); } }

@media (max-width: 600px) {
  .judging-animation { min-height: 126px; }
  .judging-scene { width: 96%; height: 132px; }
}

@media (prefers-reduced-motion: reduce) {
  .flight-line, .code-card, .owl, .owl-eyes, .owl-wing, .lamp-body, .lamp-glow, .star { animation: none; }
  .lamp-body { fill: #2faf76; }
  .lamp-glow { display: none; }
}
</style>
