import { useCallback, useEffect, useState } from 'react';
import api from '../../services/api';
import { getCurrentUserID } from '../../utils/auth';

const ROLE_OPTIONS = ['Student', 'Lecturer', 'Admin']; // whitelist sama dengan backend

const ROLE_BADGE = {
  Admin: 'bg-[#112320] text-[#F8F9FA]',
  Lecturer: 'bg-[#D97757]/10 text-[#C26244] border border-[#D97757]/30',
  Student: 'bg-[#F8F9FA] text-[#64748B] border border-[#E2E8F0]',
};

// Tabel User Management (Minggu 6 Day 5): data 100% dari backend
// GET /api/admin/users (DTO admin: id, name, email, role, banned, created_at).
// Setelah ban/unban/role sukses, state di-refresh dari response server —
// frontend bukan sumber kebenaran.
const UserManagement = () => {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [busyID, setBusyID] = useState(''); // row yang sedang diproses (anti double-click)
  const [draftRole, setDraftRole] = useState({}); // { [userID]: pilihan role terpilih }
  const currentUID = getCurrentUserID();

  // Muat ulang manual / pasca-aksi (dipanggil dari event handler, bukan efek):
  // state di-refresh dari response server agar row selalu mencerminkan DB.
  const fetchUsers = useCallback(async () => {
    try {
      setLoading(true);
      setError('');
      const res = await api.get('/admin/users');
      setUsers(res.data.data || []);
    } catch (err) {
      const status = err.response?.status;
      if (status === 401) {
        setError('Sesi berakhir. Muat ulang halaman untuk login kembali.');
      } else if (status === 403) {
        setError('Akses ditolak: endpoint ini khusus admin.');
      } else {
        setError(err.response?.data?.error || 'Gagal memuat daftar user.');
      }
    } finally {
      setLoading(false);
    }
  }, []);

  // Load awal: IIFE async dengan semua setState SETELAH await
  // (pola yang sama dengan AdminDashboard — tidak ada setState sinkron di efek).
  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        const res = await api.get('/admin/users');
        if (cancelled) return;
        setUsers(res.data.data || []);
      } catch (err) {
        if (cancelled) return;
        const status = err.response?.status;
        if (status === 401) {
          setError('Sesi berakhir. Muat ulang halaman untuk login kembali.');
        } else if (status === 403) {
          setError('Akses ditolak: endpoint ini khusus admin.');
        } else {
          setError(err.response?.data?.error || 'Gagal memuat daftar user.');
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();

    return () => {
      cancelled = true;
    };
  }, []);

  // --- Ban / Unban: panggil backend, update state HANYA jika sukses ---
  const handleSetBanned = async (user, banned) => {
    if (busyID) return; // cegah double-click / aksi paralel
    setBusyID(user.id);
    setError('');
    try {
      await api.post(`/admin/users/${user.id}/${banned ? 'ban' : 'unban'}`);
      await fetchUsers(); // refresh dari server (row mencerminkan DB)
    } catch (err) {
      const status = err.response?.status;
      const message =
        status === 400 || status === 404
          ? err.response?.data?.error // self-ban / target hilang
          : 'Gagal memperbarui status user.';
      setError(message);
      // State sebelumnya dipertahankan — tidak ada perubahan palsu.
    } finally {
      setBusyID('');
    }
  };

  // --- Change Role: simpan role terpilih lalu kirim ke backend ---
  const handleSetRole = async (user) => {
    const role = draftRole[user.id];
    if (!role || role === user.role || busyID) return;
    setBusyID(user.id);
    setError('');
    try {
      await api.post(`/admin/users/${user.id}/role`, { role });
      await fetchUsers();
      setDraftRole((prev) => {
        const next = { ...prev };
        delete next[user.id];
        return next;
      });
    } catch (err) {
      const status = err.response?.status;
      const message =
        status === 400 || status === 404
          ? err.response?.data?.error // invalid role / last-admin / 404
          : 'Gagal mengubah role user.';
      setError(message);
    } finally {
      setBusyID('');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24 bg-white rounded-2xl border border-[#E2E8F0]">
        <div className="w-10 h-10 border-4 border-[#D97757] border-t-transparent rounded-full animate-spin"></div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-2xl border border-[#E2E8F0] shadow-sm overflow-hidden">
      <div className="px-6 py-5 border-b border-[#E2E8F0] flex flex-wrap items-center justify-between gap-2">
        <h2 className="font-display text-xl font-semibold text-[#1E293B]">
          User Management <span className="text-[#94A3B8] text-base">({users.length})</span>
        </h2>
        <button
          type="button"
          onClick={fetchUsers}
          className="text-xs font-bold uppercase tracking-wider text-[#64748B] hover:text-[#D97757] px-3 py-1.5 rounded-full hover:bg-[#F8F9FA] transition-colors cursor-pointer"
        >
          ↻ Muat ulang
        </button>
      </div>

      {error && (
        <div className="mx-6 mt-4 p-4 bg-red-50 border-l-4 border-red-500 text-red-700 text-sm font-medium rounded">
          {error}
        </div>
      )}

      {users.length === 0 ? (
        // Empty state (data memang kosong, bukan karena fetch belum selesai —
        // cabang ini hanya tercapai setelah loading=false)
        <div className="text-center py-16">
          <p className="text-4xl mb-3">👥</p>
          <p className="font-display font-semibold text-[#1E293B]">Belum ada user terdaftar</p>
        </div>
      ) : (
        // Horizontal scroll container untuk layar sempit (627px):
        // tabel tetap usable tanpa merusak layout halaman.
        <div className="overflow-x-auto">
          <table className="w-full min-w-[640px] text-sm">
            <thead>
              <tr className="bg-[#F8F9FA] text-left text-[10px] font-bold uppercase tracking-wider text-[#94A3B8]">
                <th className="px-6 py-3">Nama</th>
                <th className="px-4 py-3">Email</th>
                <th className="px-4 py-3">Role</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-6 py-3 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-[#E2E8F0]">
              {users.map((user) => {
                const busy = busyID === user.id;
                const isSelf = user.id === currentUID;
                const selectedRole = draftRole[user.id] || user.role;
                const roleChanged = selectedRole !== user.role;

                return (
                  <tr key={user.id} className={busy ? 'opacity-60' : ''}>
                    <td className="px-6 py-4 font-bold text-[#1E293B] whitespace-nowrap">
                      {user.name}
                      {isSelf && (
                        <span className="ml-2 text-[10px] font-bold uppercase text-[#D97757]">
                          (kamu)
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-4 text-[#64748B] whitespace-nowrap">{user.email}</td>
                    <td className="px-4 py-4">
                      <span
                        className={`inline-block text-[10px] font-bold uppercase tracking-wider rounded-full px-2.5 py-1 ${ROLE_BADGE[user.role] || ROLE_BADGE.Student}`}
                      >
                        {user.role}
                      </span>
                    </td>
                    <td className="px-4 py-4">
                      <span
                        className={`inline-flex items-center gap-1.5 text-xs font-semibold ${
                          user.banned ? 'text-red-600' : 'text-emerald-600'
                        }`}
                      >
                        <span
                          className={`w-1.5 h-1.5 rounded-full ${user.banned ? 'bg-red-500' : 'bg-emerald-500'}`}
                        ></span>
                        {user.banned ? 'Diblokir' : 'Aktif'}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap items-center justify-end gap-2">
                        {/* Ban / Unban: destructive (ban) pakai confirm existing */}
                        {user.banned ? (
                          <button
                            type="button"
                            onClick={() => handleSetBanned(user, false)}
                            disabled={Boolean(busyID)}
                            className="text-xs font-bold text-emerald-700 border border-emerald-200 hover:bg-emerald-50 rounded-full px-3 py-1.5 transition-colors disabled:opacity-60 cursor-pointer"
                          >
                            {busy ? '…' : 'Unban'}
                          </button>
                        ) : (
                          <button
                            type="button"
                            onClick={() => {
                              if (window.confirm(`Blokir akun ${user.name}? User tidak akan bisa login/mengakses API.`)) {
                                handleSetBanned(user, true);
                              }
                            }}
                            disabled={Boolean(busyID)}
                            className="text-xs font-bold text-red-600 border border-red-200 hover:bg-red-50 rounded-full px-3 py-1.5 transition-colors disabled:opacity-60 cursor-pointer"
                          >
                            {busy ? '…' : 'Ban'}
                          </button>
                        )}

                        {/* Change role: hanya opsi whitelist, tombol apply
                            aktif hanya saat pilihan berubah dari role server */}
                        <select
                          value={selectedRole}
                          onChange={(e) =>
                            setDraftRole((prev) => ({ ...prev, [user.id]: e.target.value }))
                          }
                          disabled={Boolean(busyID)}
                          className="text-xs font-semibold text-[#1E293B] border border-[#E2E8F0] rounded-full px-2.5 py-1.5 bg-white focus:border-[#D97757] focus:outline-none disabled:opacity-60 cursor-pointer"
                        >
                          {ROLE_OPTIONS.map((role) => (
                            <option key={role} value={role}>
                              {role}
                            </option>
                          ))}
                        </select>
                        <button
                          type="button"
                          onClick={() => handleSetRole(user)}
                          disabled={Boolean(busyID) || !roleChanged}
                          className={`text-xs font-bold rounded-full px-3 py-1.5 transition-colors disabled:opacity-60 cursor-pointer ${
                            roleChanged
                              ? 'bg-[#D97757] text-white hover:bg-[#C26244]'
                              : 'bg-[#E2E8F0] text-[#94A3B8]'
                          }`}
                        >
                          {busy ? '…' : 'Terapkan'}
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default UserManagement;
