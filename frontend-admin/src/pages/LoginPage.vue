<template>
  <main style="display: grid; place-items: center; min-height: 100vh;">
    <section class="panel" style="width: 360px;">
      <h1>管理员登录</h1>
      <form style="display: grid; gap: 12px;" @submit.prevent="submit">
        <BaseInput v-model="email" type="email" placeholder="管理员邮箱" label="邮箱" :error="errors.email" />
        <BaseInput v-model="password" type="password" placeholder="密码" label="密码" :error="errors.password" />
        <BaseButton type="submit">登录</BaseButton>
      </form>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useAdminStore } from '../stores/user';
import BaseInput from '../components/ui/BaseInput.vue';
import BaseButton from '../components/ui/BaseButton.vue';
import { email as emailRule, required, runValidators } from '../utils/validators';

const router = useRouter();
const store = useAdminStore();
const email = ref('');
const password = ref('');
const errors = ref({ email: '', password: '' });

function validate(): boolean {
  errors.value.email = runValidators(email.value, [required('邮箱'), emailRule()]);
  errors.value.password = runValidators(password.value, [required('密码')]);
  return !errors.value.email && !errors.value.password;
}

async function submit(): Promise<void> {
  if (!validate()) {
    return;
  }
  await store.login(email.value, password.value);
  router.push('/dashboard');
}
</script>
