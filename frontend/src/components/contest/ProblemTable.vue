<template>
  <table class="data-table">
    <colgroup>
      <col style="width: 10%;">
      <col style="width: 30%;">
      <col style="width: 15%;">
      <col style="width: 15%;">
      <col style="width: 15%;">
      <col style="width: 15%;">
    </colgroup>
    <thead>
      <tr>
        <th>#</th>
        <th class="left">标题</th>
        <th class="center">通过数</th>
        <th class="center">提交数</th>
        <th>通过率</th>
        <th class="center">状态</th>
      </tr>
    </thead>
    <tbody>
      <tr class="tracking-wide font-[Arial,'Noto_Sans_SC',sans-serif]"
       v-for="item in problemList" :key="item.label">
        <td>
          <div class="flex items-center gap-1 justify-center">
            <span class="iconfont icon-qiqiu1" :style="{
              color: item?.color || 'green',
              fontSize: '18px',
            }">
            </span>
            <span>
              {{ item.label }}
            </span>
          </div>
        </td>
        <td class="left">
          <span class="f-link" @click="handleClickProblem(item.label)">{{ item.name }}</span>
        </td>
        <td class="center">
          {{ item.acceptedCount }}
        </td>
        <td class="center">
          {{ item.totalCount }}
        </td>
        <td class="center">
          {{ item.totalCount === 0 ? '0' : Math.floor(100 * item.acceptedCount / item.totalCount) }} %
        </td>
        <td class="center">
          <span v-if="item.status === AcceptedCode" class="iconfont icon-duihao1 ac"></span>
          <span v-else-if="item.status === WrongAnswerCode" class="iconfont icon-chahao wa"></span>
        </td>
      </tr>
    </tbody>
  </table>
</template>


<script setup>
import { AcceptedCode, WrongAnswerCode } from '@/constants/index'

defineProps({
  problemList: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['click-problem', 'click-submit'])

const handleClickProblem = (label) => {
  emit('click-problem', label);
}

</script>

<style lang="scss" scoped>
// .problem-table {
//   width: 100%;
//   border-collapse: collapse;
//   table-layout: fixed;

//   .left {
//     text-align: left;
//   }

//   .center {
//     text-align: center;
//   }

//   .right {
//     text-align: right;
//   }

//   tr td,
//   tr th {
//     padding: 0px 4px;
//     text-align: center;
//   }

  .ac {
    color: #4CAF50;
  }

  .wa {
    color: #F44336;
  }


//   thead tr th {
//     background-color: #F5F5F5;
//     font-size: 16px;
//     color: #333;
//     font-weight: normal;
//     padding-top: 14px;
//     padding-bottom: 14px;
//   }

//   tbody tr {
//     font-size: 14px;
//     color: #333;
//     background-color: #fff;
//     transition: all 0.5s;

//     border-bottom: 1px solid #EEEEEE;

//     td {
//       box-sizing: border-box;
//       padding-top: 12px;
//       padding-bottom: 12px;
//     }
//   }


//   .link-problem span {
//     cursor: pointer;
//     color: #3490de;
//   }
// }
</style>
