<template>
  <div ref="chartRef" style="width: 100%; height: 340px"></div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import { ratingTier } from '@/constants/index'

const props = defineProps({
  history: { type: Array, default: () => [] },
})

const chartRef = ref(null)
let chart = null

// 段位区间 [from,to)（从低到高）+ 背景色带浅色。
// 用一组明显区分的浅色，避免纯蓝/紫淡到 10% 后互相发紫、分不清（Expert 蓝 vs CM 紫）。
const TIER_BANDS = [
  { from: 0, to: 1200, band: '#eef0f2' },    // Newbie 灰
  { from: 1200, to: 1400, band: '#e7f4e7' },  // Pupil 绿
  { from: 1400, to: 1600, band: '#dff2f0' },  // Specialist 青
  { from: 1600, to: 1900, band: '#e2ebff' },  // Expert 蓝
  { from: 1900, to: 2100, band: '#f4e3f7' },  // CM 紫
  { from: 2100, to: 2400, band: '#fdeeda' },  // Master 橙
  { from: 2400, to: 4000, band: '#fde3e3' },  // GM 红
]

const fmtDate = (ms) => {
  const d = new Date(ms)
  const p = (n) => String(n).padStart(2, '0')
  return `${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

const render = () => {
  if (!chart) return
  const hist = props.history || []
  const has = hist.length > 0
  const ratings = hist.map((h) => h.newRating)
  const lo = has ? Math.min(...ratings) : 1200
  const hi = has ? Math.max(...ratings) : 1700
  const yMin = Math.max(0, Math.floor((lo - 120) / 100) * 100)
  const yMax = Math.ceil((hi + 120) / 100) * 100

  // 只画落在 [yMin,yMax] 内的段位色带
  const bands = TIER_BANDS
    .map((b) => {
      const from = Math.max(b.from, yMin)
      const to = Math.min(b.to, yMax)
      return from < to
        ? [{ yAxis: from, itemStyle: { color: b.band } }, { yAxis: to }]
        : null
    })
    .filter(Boolean)

  const data = hist.map((h) => [h.time, h.newRating])

  chart.setOption({
    grid: { left: 44, right: 16, top: 16, bottom: 26 },
    tooltip: {
      trigger: 'axis',
      formatter: (ps) => {
        const h = hist[ps[0].dataIndex]
        if (!h) return ''
        const sign = h.delta >= 0 ? '+' + h.delta : String(h.delta)
        const c = h.delta >= 0 ? '#16a34a' : '#dc2626'
        return `${h.contestName}<br/>Rating: <b>${h.newRating}</b> <span style="color:${c}">(${sign})</span><br/>名次: ${h.rank}`
      },
    },
    xAxis: {
      type: 'time',
      axisLabel: { color: '#9aa1ab', formatter: fmtDate, hideOverlap: true },
      axisLine: { lineStyle: { color: '#d6dbe1' } },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value', min: yMin, max: yMax, interval: 200,
      axisLabel: { color: '#9aa1ab' },
      axisLine: { show: false },
      axisTick: { show: false },
      splitLine: { lineStyle: { color: '#eef1f4' } },
    },
    graphic: has ? [] : [{
      type: 'text', left: 'center', top: 'middle', silent: true,
      style: { text: '还没有参加 rated 比赛', fill: '#b0b7c0', fontSize: 13 },
    }],
    series: [{
      type: 'line',
      data,
      z: 3,
      symbolSize: 7,
      showSymbol: true,
      smooth: false,
      lineStyle: { color: '#9aa1ab', width: 2 },
      itemStyle: {
        color: (p) => ratingTier(Array.isArray(p.value) ? p.value[1] : p.value).color,
        borderColor: '#fff',
        borderWidth: 1.5,
      },
      markArea: bands.length ? { silent: true, data: bands } : undefined,
    }],
  }, { notMerge: true })
}

const initChart = () => {
  chart = echarts.init(chartRef.value)
  render()
}

const onResize = () => chart && chart.resize()

onMounted(() => {
  nextTick(initChart)
  window.addEventListener('resize', onResize)
})
watch(() => props.history, render, { deep: true })
onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  if (chart) chart.dispose()
})
</script>
