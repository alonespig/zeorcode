<template>
  <div ref="chartRef" class="chart"></div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref, watch, nextTick } from 'vue'
import echarts from '@/utils/echarts'

// 接收数据（核心）
const props = defineProps({
  dates: {
    type: Array,
    default: () => []
  },
  counts: {
    type: Array,
    default: () => []
  },
})

const chartRef = ref(null)
let chart = null
let resizeObserver = null

//  渲染图表
const renderChart = () => {
  if (!chart) return
  chart.setOption({
    backgroundColor: 'transparent',

    grid: {
      left: 10,
      right: 10,
      top: 30,
      bottom: 10
    },

    tooltip: {
      trigger: 'axis',
      backgroundColor: '#fff',
      borderColor: '#eee',
      borderWidth: 1,
      textStyle: {
        color: '#333'
      }
    },

    xAxis: {
      type: 'category',
      data: props.dates,

      axisLine: { show: false },
      axisTick: { show: false },

      axisLabel: {
        color: '#999',
        fontSize: 12
      }
    },

    yAxis: {
      type: 'value',

      minInterval: 1, // 过题数是整数，刻度步长至少 1，不出现 0.2/0.4 这种小数
      max: ({ max }) => Math.max(max, 4), // 没刷题(全 0)时也撑到 0~4，不至于只剩「0 1」

      axisLine: { show: false },
      axisTick: { show: false },

      axisLabel: {
        color: '#aaa',
        fontSize: 12
      },

      splitLine: {
        lineStyle: {
          color: '#f0f0f0'
        }
      }
    },

    series: [
      {
        type: 'line',
        data: props.counts,

        smooth: true,

        // ✅ 关键：细线
        lineStyle: {
          width: 2,
          color: '#5470C6'
        },

        // ✅ 点变小
        symbol: 'circle',
        symbolSize: 5,

        itemStyle: {
          color: '#5470C6'
        },

        // ✅ 去掉大面积阴影（之前丑的原因之一）
        areaStyle: undefined,

        // ✅ hover 高亮
        emphasis: {
          focus: 'series'
        }
      }
    ]
  })
}

onMounted(() => {
  chart = echarts.init(chartRef.value)

  renderChart()

  //  用 ResizeObserver（比 window.resize 更准）
  resizeObserver = new ResizeObserver(() => {
    chart.resize()
  })

  resizeObserver.observe(chartRef.value)
})

//  数据变化自动更新
watch(
  () => [props.dates, props.counts],
  async () => {
    await nextTick()
    renderChart()
  },
  { deep: true }
)

onBeforeUnmount(() => {
  resizeObserver && resizeObserver.disconnect()
  chart && chart.dispose()
})
</script>

<style scoped>
.chart {
  width: 100%;
  height: 300px;
}
</style>
