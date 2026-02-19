import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import http from '../api/http';

interface AdminProfile {
  id: number;
  email: string;
  role: string;
}

export const useAdminStore = defineStore('admin-user', () => {
  const token = ref<string>(localStorage.getItem('admin_access_token') ?? '');
  const profile = ref<AdminProfile | null>(null);

  const isLoggedIn = computed(() => Boolean(token.value));

  function setToken(nextToken: string): void {
    token.value = nextToken;
    localStorage.setItem('admin_access_token', nextToken);
  }

  async function login(email: string, password: string): Promise<void> {
    const res = await http.post('/pub/login', { email, password });
    const user = res.data.data.user as AdminProfile;
    if (user.role !== 'admin') {
      throw new Error('仅管理员可登录后台');
    }
    setToken(res.data.data.access_token);
    profile.value = user;
  }

  function logout(): void {
    token.value = '';
    profile.value = null;
    localStorage.removeItem('admin_access_token');
  }

  return {
    token,
    profile,
    isLoggedIn,
    login,
    logout,
  };
});
