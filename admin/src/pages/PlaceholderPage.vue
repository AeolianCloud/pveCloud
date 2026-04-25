<script setup lang="ts">
import { computed } from 'vue'
import { CalendarDays, Download, Filter, Plus, Search } from 'lucide-vue-next'
import { useRoute } from 'vue-router'

const route = useRoute()
const title = computed(() => String(route.meta.title || '业务管理'))

const rows = computed(() => [
  [`${title.value}-001`, '上海智云科技有限公司', '上海节点1', '在线', '05-22 10:23'],
  [`${title.value}-002`, '杭州数海信息技术有限公司', '杭州节点2', '处理中', '05-22 10:18'],
  [`${title.value}-003`, '深圳云拓网络有限公司', '深圳节点1', '待确认', '05-22 10:05'],
  [`${title.value}-004`, '广州星河智能科技', '北京节点1', '在线', '05-22 09:58'],
  [`${title.value}-005`, '北京启航数据服务', '华南节点', '在线', '05-22 09:47'],
])
</script>

<template>
  <section class="workspace-page">
    <header class="workspace-toolbar">
      <div>
        <span>业务工作台</span>
        <h1>{{ title }}</h1>
      </div>
      <div class="toolbar-actions">
        <label class="toolbar-search">
          <Search :size="18" aria-hidden="true" />
          <input type="search" :placeholder="`搜索${title}`" />
        </label>
        <button type="button"><CalendarDays :size="18" aria-hidden="true" />近30天</button>
        <button type="button"><Filter :size="18" aria-hidden="true" />筛选</button>
        <button type="button"><Download :size="18" aria-hidden="true" />导出</button>
        <button class="primary-button" type="button"><Plus :size="18" aria-hidden="true" />新增</button>
      </div>
    </header>

    <div class="workspace-kpis">
      <article>
        <span>今日新增</span>
        <strong>36</strong>
        <em>较昨日 +12.4%</em>
      </article>
      <article>
        <span>处理中</span>
        <strong>128</strong>
        <em>平均响应 18min</em>
      </article>
      <article>
        <span>累计数量</span>
        <strong>2,569</strong>
        <em>达成率 98.6%</em>
      </article>
      <article>
        <span>待关注</span>
        <strong>7</strong>
        <em>需要人工处理</em>
      </article>
    </div>

    <article class="workspace-table-card">
      <header>
        <h2>{{ title }}列表</h2>
        <span>当前模块沿用后台统一表格、筛选和操作视觉。</span>
      </header>
      <div class="workspace-table-scroll">
        <table>
          <thead>
            <tr>
              <th>编号</th>
              <th>客户名称</th>
              <th>节点</th>
              <th>状态</th>
              <th>更新时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in rows" :key="row[0]">
              <td>{{ row[0] }}</td>
              <td>{{ row[1] }}</td>
              <td>{{ row[2] }}</td>
              <td><span>{{ row[3] }}</span></td>
              <td>{{ row[4] }}</td>
              <td><a href="#">查看</a></td>
            </tr>
          </tbody>
        </table>
      </div>
    </article>
  </section>
</template>

<style scoped>
.workspace-page {
  display: grid;
  gap: 18px;
}

.workspace-toolbar,
.workspace-table-card,
.workspace-kpis article {
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--panel);
  box-shadow: var(--shadow-soft);
}

.workspace-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 18px;
}

.workspace-toolbar span {
  color: var(--muted);
  font-size: 13px;
  font-weight: 800;
}

.workspace-toolbar h1 {
  margin: 6px 0 0;
  font-size: 24px;
  line-height: 1.2;
}

.toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 10px;
}

.toolbar-actions button,
.toolbar-search {
  min-height: 40px;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 0 13px;
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--muted-strong);
  background: var(--panel);
  cursor: pointer;
  font-weight: 750;
}

.toolbar-search {
  min-width: 260px;
  cursor: text;
}

.toolbar-search input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  color: var(--text);
  background: transparent;
}

.workspace-kpis {
  display: grid;
  grid-template-columns: repeat(4, minmax(160px, 1fr));
  gap: 14px;
}

.workspace-kpis article {
  display: grid;
  gap: 8px;
  padding: 18px;
}

.workspace-kpis span {
  color: var(--muted);
  font-weight: 800;
}

.workspace-kpis strong {
  font-size: 28px;
  line-height: 1;
}

.workspace-kpis em {
  color: var(--success);
  font-style: normal;
  font-weight: 750;
}

.workspace-table-card {
  overflow: hidden;
}

.workspace-table-card header {
  min-height: 66px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  padding: 0 18px;
  border-bottom: 1px solid var(--border);
}

.workspace-table-card h2 {
  margin: 0;
  font-size: 18px;
}

.workspace-table-card header span {
  color: var(--muted);
  font-size: 13px;
}

.workspace-table-scroll {
  overflow-x: auto;
  padding: 0 16px 16px;
}

table {
  width: 100%;
  min-width: 760px;
  border-collapse: collapse;
  color: var(--muted-strong);
}

th,
td {
  height: 52px;
  padding: 0 12px;
  border-bottom: 1px solid var(--border);
  text-align: left;
  white-space: nowrap;
}

th {
  height: 46px;
  color: #53627a;
  background: var(--panel-soft);
  font-size: 13px;
}

td span {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 9px;
  border-radius: 6px;
  color: var(--success);
  background: var(--success-soft);
  font-weight: 800;
}

td a {
  color: var(--primary);
  font-weight: 800;
  text-decoration: none;
}

@media (max-width: 980px) {
  .workspace-toolbar {
    align-items: flex-start;
    flex-direction: column;
  }

  .toolbar-actions {
    justify-content: flex-start;
  }

  .toolbar-search {
    min-width: min(100%, 320px);
  }

  .workspace-kpis {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 620px) {
  .workspace-kpis {
    grid-template-columns: 1fr;
  }
}
</style>
