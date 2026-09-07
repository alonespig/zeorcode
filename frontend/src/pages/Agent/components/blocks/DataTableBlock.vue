<template>
  <section class="table-card">
    <div class="table-heading">
      <h3>{{ block.title }}</h3>
      <span>{{ block.rows?.length || 0 }} 项</span>
    </div>
    <div class="table-scroll">
      <table>
        <thead>
          <tr>
            <th
              v-for="column in block.columns || []"
              :key="column.key"
              :style="columnStyle(column)"
            >{{ column.label }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, index) in block.rows || []" :key="row.id || index">
            <td
              v-for="column in block.columns || []"
              :key="column.key"
              :style="{ textAlign: column.align || 'left' }"
            >
              <template v-if="column.formatter === 'difficulty'">
                <span :class="['difficulty', `level-${row[column.key]}`]">{{ difficultyText(row[column.key]) }}</span>
              </template>
              <template v-else-if="column.formatter === 'tags'">
                <span class="tag-list">
                  <span v-for="tag in row[column.key] || []" :key="tag">{{ tag }}</span>
                </span>
              </template>
              <RouterLink v-else-if="column.key === 'name'" :to="`/problem/${row.id}`" class="problem-link">
                {{ row[column.key] }}
              </RouterLink>
              <template v-else>{{ row[column.key] }}</template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup>
defineProps({ block: { type: Object, required: true } })
const difficultyText = (value) => ({ 1: '入门', 2: '普及', 3: '提高' }[value] || '未知')
const columnStyle = (column) => ({ width: column.width ? `${column.width}px` : undefined, textAlign: column.align || 'left' })
</script>

<style scoped>
.table-card { margin-top: 14px; overflow: hidden; border: 1px solid #e0e6ee; border-radius: 6px; background: #fff; }
.table-heading { height: 47px; display: flex; align-items: center; justify-content: space-between; padding: 0 16px; background: #f8faff; border-bottom: 1px solid #e8edf4; }
.table-heading h3 { margin: 0; color: #273348; font-size: 14px; }
.table-heading span { color: #929cab; font-size: 12px; }
.table-scroll { overflow-x: auto; }
table { width: 100%; min-width: 720px; border-collapse: collapse; table-layout: auto; }
th { height: 39px; padding: 0 12px; color: #738096; background: #fbfcfe; border-bottom: 1px solid #e9edf3; font-size: 12px; font-weight: 600; white-space: nowrap; }
td { height: 48px; padding: 7px 12px; color: #455167; border-bottom: 1px solid #edf0f5; font-size: 13px; }
tbody tr:last-child td { border-bottom: 0; }
tbody tr:hover td { background: #f8faff; }
.problem-link { color: #276bdb; font-weight: 600; }
.problem-link:hover { text-decoration: underline; }
.difficulty { display: inline-block; min-width: 42px; padding: 3px 7px; border-radius: 4px; font-size: 11px; text-align: center; }
.level-1 { color: #267451; background: #eaf7f0; }
.level-2 { color: #9a6500; background: #fff5dc; }
.level-3 { color: #b64d50; background: #fff0f0; }
.tag-list { display: flex; flex-wrap: wrap; gap: 4px; }
.tag-list span { padding: 2px 6px; color: #5d6b83; background: #f0f3f8; border-radius: 3px; font-size: 10px; white-space: nowrap; }
</style>
