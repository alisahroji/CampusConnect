import axios from 'axios';

// Buat instance axios dengan base URL backend Golang kita
const api = axios.create({
  baseURL: 'http://localhost:8080/api', 
});

// Interceptor: Otomatis menyisipkan token ke setiap request
api.interceptors.request.use(
  (config) => {
    // Kita akan menyimpan token di localStorage browser
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export default api;