import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../lib/api'
import { Plus, Trash2 } from 'lucide-react'
import type { User } from '../types'

export default function Users() {
  const queryClient = useQueryClient()
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState({ username: '', email: '', password: '', role: 'user' as 'admin' | 'user' })

  const { data: users = [] } = useQuery<User[]>({
    queryKey: ['users'],
    queryFn: () => api.get('/users').then(r => r.data),
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof form) => api.post('/users', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setShowModal(false)
      setForm({ username: '', email: '', password: '', role: 'user' })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => api.delete(`/users/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  })

  const toggleActive = useMutation({
    mutationFn: ({ id, is_active }: { id: number; is_active: boolean }) =>
      api.put(`/users/${id}`, { is_active: !is_active }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['users'] }),
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Users</h1>
          <p className="text-slate-400 mt-1">Manage user accounts and permissions</p>
        </div>
        <button onClick={() => setShowModal(true)} className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm">
          <Plus className="w-4 h-4" /> Add User
        </button>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
        <table className="w-full">
          <thead className="bg-slate-800/50">
            <tr>
              <th className="text-left px-5 py-3 text-xs font-medium text-slate-400 uppercase">User</th>
              <th className="text-left px-5 py-3 text-xs font-medium text-slate-400 uppercase">Email</th>
              <th className="text-left px-5 py-3 text-xs font-medium text-slate-400 uppercase">Role</th>
              <th className="text-left px-5 py-3 text-xs font-medium text-slate-400 uppercase">Status</th>
              <th className="text-left px-5 py-3 text-xs font-medium text-slate-400 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {users.map((user) => (
              <tr key={user.id} className="hover:bg-slate-800/30">
                <td className="px-5 py-4">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-slate-700 rounded-full flex items-center justify-center text-sm font-medium text-white">
                      {user.username.charAt(0).toUpperCase()}
                    </div>
                    <span className="text-sm font-medium text-white">{user.username}</span>
                  </div>
                </td>
                <td className="px-5 py-4 text-sm text-slate-400">{user.email}</td>
                <td className="px-5 py-4">
                  <span className={`px-2.5 py-1 text-xs rounded-full ${user.role === 'admin' ? 'bg-purple-500/20 text-purple-400' : 'bg-slate-700 text-slate-300'}`}>
                    {user.role}
                  </span>
                </td>
                <td className="px-5 py-4">
                  <button onClick={() => toggleActive.mutate({ id: user.id, is_active: user.is_active })}
                    className={`px-2.5 py-1 text-xs rounded-full cursor-pointer ${user.is_active ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                    {user.is_active ? 'Active' : 'Disabled'}
                  </button>
                </td>
                <td className="px-5 py-4">
                  <button onClick={() => { if (confirm('Delete this user?')) deleteMutation.mutate(user.id) }}
                    className="p-2 text-slate-500 hover:text-red-400 transition-colors">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-slate-900 border border-slate-700 rounded-2xl w-full max-w-md p-6 space-y-4">
            <h2 className="text-xl font-bold text-white">Create User</h2>
            <input value={form.username} onChange={e => setForm({...form, username: e.target.value})} placeholder="Username" className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white" />
            <input value={form.email} onChange={e => setForm({...form, email: e.target.value})} placeholder="Email" type="email" className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white" />
            <input value={form.password} onChange={e => setForm({...form, password: e.target.value})} placeholder="Password" type="password" className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white" />
            <select value={form.role} onChange={e => setForm({...form, role: e.target.value as 'admin' | 'user'})} className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white">
              <option value="user">User</option>
              <option value="admin">Admin</option>
            </select>
            <div className="flex justify-end gap-3 pt-2">
              <button onClick={() => setShowModal(false)} className="px-4 py-2 text-slate-400 hover:text-white">Cancel</button>
              <button onClick={() => createMutation.mutate(form)} className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg">Create</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
