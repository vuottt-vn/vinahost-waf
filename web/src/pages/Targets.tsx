import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../lib/api'
import { useAuthStore } from '../store/auth'
import { Plus, Trash2, ExternalLink } from 'lucide-react'
import type { ProxyTarget } from '../types'

export default function Targets() {
  const queryClient = useQueryClient()
  const isAdmin = useAuthStore((s) => s.isAdmin)
  const [showModal, setShowModal] = useState(false)
  const [form, setForm] = useState({ name: '', upstream_url: '' })

  const { data: targets = [] } = useQuery<ProxyTarget[]>({
    queryKey: ['targets'],
    queryFn: () => api.get('/targets').then(r => r.data),
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof form) => api.post('/targets', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['targets'] })
      setShowModal(false)
      setForm({ name: '', upstream_url: '' })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => api.delete(`/targets/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['targets'] }),
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Proxy Targets</h1>
          <p className="text-slate-400 mt-1">Manage upstream servers that the WAF protects</p>
        </div>
        {isAdmin && (
          <button onClick={() => setShowModal(true)} className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm">
            <Plus className="w-4 h-4" /> Add Target
          </button>
        )}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {targets.map((target) => (
          <div key={target.id} className="bg-slate-900 border border-slate-800 rounded-xl p-5">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-semibold text-white">{target.name}</h3>
                <a href={target.upstream_url} target="_blank" rel="noreferrer" className="text-sm text-blue-400 hover:text-blue-300 flex items-center gap-1 mt-1">
                  {target.upstream_url} <ExternalLink className="w-3 h-3" />
                </a>
              </div>
              <span className={`px-2 py-0.5 text-xs rounded-full ${target.is_enabled ? 'bg-green-500/20 text-green-400' : 'bg-slate-700 text-slate-400'}`}>
                {target.is_enabled ? 'Active' : 'Disabled'}
              </span>
            </div>
            {isAdmin && (
              <div className="mt-4 pt-3 border-t border-slate-800 flex justify-end">
                <button onClick={() => { if (confirm('Delete this target?')) deleteMutation.mutate(target.id) }} className="p-2 text-slate-500 hover:text-red-400 transition-colors">
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
            )}
          </div>
        ))}
        {targets.length === 0 && (
          <div className="col-span-3 text-center py-12 text-slate-500">
            No proxy targets configured. Add one to start protecting your applications.
          </div>
        )}
      </div>

      {showModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-slate-900 border border-slate-700 rounded-2xl w-full max-w-md p-6 space-y-4">
            <h2 className="text-xl font-bold text-white">Add Proxy Target</h2>
            <input value={form.name} onChange={e => setForm({...form, name: e.target.value})} placeholder="Target name (e.g., my-app)" className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white" />
            <input value={form.upstream_url} onChange={e => setForm({...form, upstream_url: e.target.value})} placeholder="Upstream URL (e.g., http://localhost:3001)" className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white" />
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
