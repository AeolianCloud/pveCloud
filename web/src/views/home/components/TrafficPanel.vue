<script setup lang="ts">
type TrafficBar = {
  rest: number
  mid: number
  peak: number
  delay: number
  duration: number
}

const trafficBars: TrafficBar[] = [
  { rest: 34, mid: 52, peak: 68, delay: -0.4, duration: 2.4 },
  { rest: 58, mid: 72, peak: 84, delay: -1.7, duration: 2.85 },
  { rest: 42, mid: 64, peak: 76, delay: -0.9, duration: 2.55 },
  { rest: 72, mid: 86, peak: 96, delay: -2.2, duration: 3.08 },
  { rest: 46, mid: 58, peak: 70, delay: -1.1, duration: 2.33 },
  { rest: 64, mid: 78, peak: 90, delay: -2.8, duration: 3.23 },
  { rest: 38, mid: 56, peak: 74, delay: -0.2, duration: 2.7 },
  { rest: 70, mid: 82, peak: 94, delay: -1.9, duration: 3 },
  { rest: 48, mid: 66, peak: 80, delay: -0.8, duration: 2.48 },
  { rest: 78, mid: 88, peak: 98, delay: -2.5, duration: 3.3 },
  { rest: 52, mid: 70, peak: 86, delay: -1.4, duration: 2.78 },
  { rest: 62, mid: 76, peak: 92, delay: -3.1, duration: 3.15 },
]
</script>

<template>
  <div class="traffic-panel mt-4 rounded-2xl border border-neutral-200 bg-white/95 px-4 py-3 text-xs font-black uppercase tracking-[0.18em] text-neutral-500 shadow-[0_12px_28px_rgb(17_17_17_/_7%)] backdrop-blur">
    <div class="traffic-header flex items-center justify-between gap-3">
      <span>Traffic</span>
      <span class="inline-flex items-center gap-2">
        <span class="traffic-status-dot"></span>
        Stable
      </span>
    </div>
    <div class="traffic-bars mt-3 grid h-14 grid-cols-12 items-end justify-items-center gap-1 overflow-hidden rounded-xl" aria-hidden="true">
      <span
        v-for="(bar, index) in trafficBars"
        :key="index"
        class="traffic-column"
        :style="{
          '--bar-rest': `${bar.rest}%`,
          '--bar-mid': `${bar.mid}%`,
          '--bar-peak': `${bar.peak}%`,
          '--traffic-delay': `${bar.delay}s`,
          '--traffic-duration': `${bar.duration}s`,
        }"
      >
        <span class="traffic-bar"></span>
      </span>
    </div>
  </div>
</template>

<style scoped>
.traffic-panel {
  overflow: hidden;
  animation: panel-enter 520ms cubic-bezier(0.2, 0.8, 0.2, 1) 460ms both;
}

.traffic-header {
  animation: panel-enter 460ms cubic-bezier(0.2, 0.8, 0.2, 1) 560ms both;
}

.traffic-status-dot {
  width: 0.4rem;
  height: 0.4rem;
  background: #111111;
  border-radius: 999px;
  animation: status-breathe 1.9s ease-in-out infinite;
}

.traffic-bars {
  position: relative;
  padding: 0.25rem 0;
  animation: chart-enter 520ms cubic-bezier(0.2, 0.8, 0.2, 1) 620ms both;
}

.traffic-bars::before {
  position: absolute;
  inset: 0;
  pointer-events: none;
  content: '';
  background:
    linear-gradient(to bottom, rgb(17 17 17 / 6%) 1px, transparent 1px) 0 0 / 100% 33.333%,
    linear-gradient(to right, transparent, rgb(17 17 17 / 5%), transparent);
}

.traffic-column {
  display: block;
  position: relative;
  width: 58%;
  height: 100%;
  overflow: hidden;
  background: rgb(17 17 17 / 5%);
  border-radius: 999px;
}

.traffic-bar {
  display: block;
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  height: var(--bar-rest, 50%);
  background: #111111;
  border-radius: 999px;
  transform-origin: bottom;
  animation: traffic-equalize var(--traffic-duration, 3.4s) cubic-bezier(0.45, 0, 0.25, 1) infinite;
  animation-delay: var(--traffic-delay, 0ms);
}

@keyframes panel-enter {
  from {
    opacity: 0;
    transform: translateY(12px) scale(0.98);
  }

  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes chart-enter {
  from {
    opacity: 0;
    transform: translateY(8px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes status-breathe {
  0%,
  100% {
    opacity: 0.45;
    transform: scale(0.86);
  }

  50% {
    opacity: 1;
    transform: scale(1.18);
  }
}

@keyframes traffic-equalize {
  0%,
  100% {
    height: var(--bar-rest, 50%);
    opacity: 0.72;
  }

  28% {
    height: var(--bar-peak, 88%);
    opacity: 1;
  }

  52% {
    height: var(--bar-mid, 68%);
    opacity: 0.86;
  }

  74% {
    height: calc((var(--bar-mid, 68%) + var(--bar-peak, 88%)) / 2);
    opacity: 0.94;
  }
}

@media (prefers-reduced-motion: reduce) {
  .traffic-panel,
  .traffic-header,
  .traffic-status-dot,
  .traffic-bars,
  .traffic-bar {
    animation: none !important;
  }

  .traffic-bar {
    height: var(--bar-mid, 68%);
  }
}
</style>
