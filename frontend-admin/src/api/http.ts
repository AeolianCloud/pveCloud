import axios, { AxiosError } from 'axios';

const http = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
});

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError<{ message?: string }>) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('admin_access_token');
      window.location.href = '/login';
      return Promise.reject(error);
    }
    const message = error.response?.data?.message ?? '请求失败，请稍后重试';
    window.alert(message);
    return Promise.reject(error);
  },
);

export default http;
