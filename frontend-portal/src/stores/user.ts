import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import http from '../api/http';

interface UserInfo {
  id: number;
  email: string;
  role: string;
}

export const useUserStore = defineStore('user', () => {
  const accessToken = ref<string>(localStorage.getItem('portal_access_token') ?? '');
  const refreshToken = ref<string>(localStorage.getItem('portal_refresh_token') ?? '');
  const profile = ref<UserInfo | null>(null);

  const isLoggedIn = computed(() => Boolean(accessToken.value));

  function setTokens(access: string, refresh: string): void {
    accessToken.value = access;
    refreshToken.value = refresh;
    localStorage.setItem('portal_access_token', access);
    localStorage.setItem('portal_refresh_token', refresh);
  }

  async function login(email: string, password: string): Promise<void> {
    const res = await http.post('/pub/login', { email, password });
    setTokens(res.data.data.access_token, res.data.data.refresh_token);
    profile.value = res.data.data.user;
  }

  async function register(email: string, password: string): Promise<void> {
    await http.post('/pub/register', { email, password });
  }

  function logout(): void {
    accessToken.value = '';
    refreshToken.value = '';
    profile.value = null;
    localStorage.removeItem('portal_access_token');
    localStorage.removeItem('portal_refresh_token');
  }

  return {
    accessToken,
    refreshToken,
    profile,
    isLoggedIn,
    login,
    register,
    logout,
    setTokens,
  };
});
