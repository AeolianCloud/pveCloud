<template>
  <main class="container" style="padding: 24px 0;">
    <section class="panel" style="max-width: 420px; margin: 0 auto;">
      <h1>用户登录</h1>
      <form class="grid" @submit.prevent="submit">
        <BaseInput v-model="email" type="email" placeholder="邮箱" label="邮箱" :error="errors.email" />
        <BaseInput v-model="password" type="password" placeholder="密码" label="密码" :error="errors.password" />
        <BaseButton type="submit">登录</BaseButton>
      </form>
      <p style="margin-top: 12px;">还没有账号？<RouterLink to="/register">去注册</RouterLink></p>
    </section>
  </main>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '../../stores/user';
import BaseInput from '../../components/ui/BaseInput.vue';
import BaseButton from '../../components/ui/BaseButton.vue';
import { email as emailRule, required, runValidators } from '../../utils/validators';

const router = useRouter();
const userStore = useUserStore();
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
  await userStore.login(email.value, password.value);
  router.push('/console');
}
</script>
