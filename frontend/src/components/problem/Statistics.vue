<template>
  <div class="statistics-container">
    <div ref="chartRef" style="width: 100%; height: 250px;"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue';
import * as echarts from 'echarts';

const props = defineProps({
  // 数据：[{ name: 'xxx', value: 数字 }]
  data: {
    type: Array,
    required: true,
    default: () => []
  },
  title: {
    type: String,
    default: '提交统计'
  },
  // 是否环形图
  isRing: {
    type: Boolean,
    default: false
  }
});

const chartRef = ref(null);
let chartInstance = null;

// 初始化图表
const initChart = () => {
  chartInstance = echarts.init(chartRef.value);
  renderChart();
};

// 渲染/更新图表
const renderChart = () => {
  const { data, title, isRing } = props;

  // 没有任何提交：不画空饼圈，居中提示
  const total = data.reduce((sum, item) => sum + (item.value || 0), 0);
  if (total === 0) {
    chartInstance.setOption({
      title: { text: title, left: 'center', textStyle: { fontSize: 13, color: '#333', fontWeight: 'bold' } },
      graphic: [{ type: 'text', left: 'center', top: 'middle', silent: true,
        style: { text: '暂无提交', fill: '#b0b7c0', fontSize: 13 } }],
    }, true);
    return;
  }

  const option = {
    color: [
      '#32CD32',
      '#FF0000',
    ],
    title: {
      text: title,
      left: 'center',
      textStyle: {
        fontSize: 13,
        color: '#333',
        fontWeight: 'bold'
      }
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'horizontal',
      bottom: 8,
      itemWidth: 12,
      itemHeight: 12,
      textStyle: {
        fontSize: 11,
        color: '#666'
      },
      data: data.map(item => item.name)
    },
    series: [
      {
        type: 'pie',
        radius: isRing ? ['38%', '62%'] : '58%', // 环形/实心
        center: ['50%', '46%'],                  // 上移一点，给底部图例留空间
        data: data,
        // 百分比放饼块内部，避免外部长标签被容器裁断；名称看下方图例
        label: {
          show: true,
          position: 'inside',
          // 值为 0 的扇区不显示标签，避免 "0%" 飘进相邻扇区里
          formatter: (p) => (p.value > 0 ? p.percent + '%' : ''),
          fontSize: 11,
          color: '#fff'
        },
        labelLine: {
          show: false
        },
      },
    ]
  };
  chartInstance.setOption(option, true);
};

// 监听数据变化
watch(() => props.data, () => {
  renderChart();
}, { deep: true });

// 自适应
const handleResize = () => {
  chartInstance && chartInstance.resize();
};

onMounted(() => {
  initChart();
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  chartInstance && chartInstance.dispose();
  window.removeEventListener('resize', handleResize);
});
</script>


<style scoped lang="scss">
.statistics-container {
  width: 100%;
}
</style>
