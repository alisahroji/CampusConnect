// Util auth kecil untuk membaca identitas user dari JWT yang tersimpan
// di localStorage (sumber token yang sama dengan services/api.js).

const decodeJWTPayload = (token) => {
  const base64Url = token.split('.')[1];
  if (!base64Url) return null;
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
  const binary = atob(base64);
  const bytes = Uint8Array.from(binary, (c) => c.charCodeAt(0));
  return JSON.parse(new TextDecoder().decode(bytes));
};

export const isLoggedIn = () => Boolean(localStorage.getItem('access_token'));

// Mengembalikan ID user (claim "sub") dari access token, atau '' bila
// anonymous / token kedaluwarsa / token tidak dapat dibaca.
// Dipakai untuk cek ownership di UI — token mati diperlakukan sebagai guest
// agar UI konsisten dengan keputusan backend.
export const getCurrentUserID = () => {
  const token = localStorage.getItem('access_token');
  if (!token) return '';
  try {
    const payload = decodeJWTPayload(token);
    if (!payload?.sub) return '';
    if (payload.exp && Date.now() / 1000 > payload.exp) return '';
    return payload.sub;
  } catch {
    return '';
  }
};

// Menghapus sesi login (dipakai tombol Logout di AppNavbar): token akses dan
// refresh dihapus dari localStorage. Destinasi redirect ditentukan pemanggil.
export const logout = () => {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
};
