<template>
  <section class="grid">
    <article class="panel" v-if="instance">
      <h2>实例详情</h2>
      <p>名称：{{ instance.name }}</p>
      <p>状态：{{ instance.status }}</p>
      <p>IP：{{ instance.ip }}</p>
      <BaseButton @click="openConsole">打开 VNC 控制台</BaseButton>
    </article>

    <article class="panel">
      <h3>监控概览</h3>
      <p>CPU：{{ metrics.cpu }}%</p>
      <p>内存：{{ metrics.memory }}%</p>
      <p>磁盘：{{ metrics.disk }}%</p>
    </article>

    <article class="panel">
      <h3>快照管理</h3>
      <form style="display: flex; gap: 8px; margin-bottom: 8px;" @submit.prevent="createSnapshot">
        <BaseInput v-model="snapshotName" placeholder="快照名称" :error="snapshotError" />
        <BaseButton type="submit">创建</BaseButton>
      </form>
      <ul>
        <li v-for="snap in snapshots" :key="snap.id" style="display: flex; gap: 8px; align-items: center;">
          <span>{{ snap.name }}</span>
          <BaseButton @click="restoreSnapshot(snap.name)">恢复</BaseButton>
          <BaseButton @click="deleteSnapshot(snap.name)">删除</BaseButton>
        </li>
      </ul>
    </article>
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import http from '../../api/http';
import BaseInput from '../../components/ui/BaseInput.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import { required, runValidators } from '../../utils/validators';

const route = useRoute();
const id = Number(route.params.id);

const instance = ref<any>(null);
const snapshots = ref<any[]>([]);
const snapshotName = ref('');
const snapshotError = ref('');
const metrics = ref({ cpu: 18, memory: 42, disk: 61 });

async function load(): Promise<void> {
  const detail = await http.get(`/user/instances/${id}`);
  const snap = await http.get(`/user/instances/${id}/snapshots`);
  instance.value = detail.data.data;
  snapshots.value = snap.data.data ?? [];
}

async function openConsole(): Promise<void> {
  const res = await http.post(`/user/instances/${id}/console`);
  window.open(res.data.data.url, '_blank');
}

async function createSnapshot(): Promise<void> {
  snapshotError.value = runValidators(snapshotName.value, [required('快照名称')]);
  if (snapshotError.value) {
    return;
  }
  await http.post(`/user/instances/${id}/snapshots`, { name: snapshotName.value });
  snapshotName.value = '';
  await load();
}

async function restoreSnapshot(name: string): Promise<void> {
  await http.post(`/user/instances/${id}/snapshots/${name}/restore`);
}

async function deleteSnapshot(name: string): Promise<void> {
  await http.delete(`/user/instances/${id}/snapshots/${name}`);
  await load();
}

onMounted(load);
</script>
