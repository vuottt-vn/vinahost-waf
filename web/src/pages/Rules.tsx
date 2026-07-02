import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../lib/api'
import { useAuthStore } from '../store/auth'
import { Plus, Search, ToggleLeft, ToggleRight, Trash2, RefreshCw } from 'lucide-react'
import type { Rule } from '../types'

export default function Rules() {
  const queryClient = useQueryClient()
  const isAdmin = useAuthStore((s) => s.isAdmin)
  const [search, setSearch] = useState('')
  const [showModal, setShowModal] = useState(false)
  const [editingRule, setEditingRule] = useState<Rule | null>(null)
  const [form, setForm] = useState({ name: '', description: '', sec_rule: '', is_enabled: true, priority: 100 })

  const { data: rules = [], isLoading } = useQuery<Rule[]>({
    queryKey: ['rules'],
    queryFn: () => api.get('/rules', { params: { search } }).then(r => r.data),
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof form) => editingRule
      ? api.put(`/rules/${editingRule.id}`, data)
      : api.post('/rules', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rules'] })
      setShowModal(false)
      resetForm()
    },
  })

  const toggleMutation = useMutation({
    mutationFn: (id: number) => api.post(`/rules/${id}/toggle`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['rules'] }),
  })

  const deleteMutation = useMutation({
    mutationFn: (id: number) => api.delete(`/rules/${id}`),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['rules'] }),
  })

  const reloadMutation = useMutation({
    mutationFn: () => api.post('/rules/reload'),
  })

  const resetForm = () => {
    setForm({ name: '', description: '', sec_rule: '', is_enabled: true, priority: 100 })
    setEditingRule(null)
  }

  const openEdit = (rule: Rule) => {
    setEditingRule(rule)
    setForm({ name: rule.name, description: rule.description, sec_rule: rule.sec_rule, is_enabled: rule.is_enabled, priority: rule.priority })
    setShowModal(true)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">WAF Rules</h1>
          <p className="text-slate-400 mt-1">Manage security rules for the firewall</p>
        </div>
        <div className="flex gap-3">
          {isAdmin && (
            <>
              <button onClick={() => reloadMutation.mutate()} className="flex items-center gap-2 px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-lg text-sm transition-colors">
                <RefreshCw className="w-4 h-4" /> Reload Engine
              </button>
              <button onClick={() => { resetForm(); setShowModal(true) }} className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm transition-colors">
                <Plus className="w-4 h-4" /> Add Rule
              </button>
            </>
          )}
        </div>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-500" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search rules..."
          className="w-full bg-slate-900 border border-slate-800 rounded-lg pl-10 pr-4 py-2.5 text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
        />
      </div>

      <div className="space-y-3">
        {rules.map((rule) => (
          <div key={rule.id} className="bg-slate-900 border border-slate-800 rounded-xl p-5">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-3">
                  <h3 className="font-semibold text-white">{rule.name}</h3>
                  <span className={`px-2 py-0.5 text-xs rounded-full ${rule.is_enabled ? 'bg-green-500/20 text-green-400' : 'bg-slate-700 text-slate-400'}`}>
                    {rule.is_enabled ? 'Enabled' : 'Disabled'}
                  </span>
                  <span className="px-2 py-0.5 text-xs bg-slate-800 text-slate-400 rounded-full">Priority: {rule.priority}</span>
                </div>
                <p className="text-sm text-slate-400 mt-1">{rule.description}</p>
                <pre className="mt-3 p-3 bg-slate-950 rounded-lg text-xs text-green-400 font-mono overflow-x-auto max-h-32">{rule.sec_rule}</pre>
              </div>
              {isAdmin && (
                <div className="flex gap-2 ml-4">
                  <button onClick={() => toggleMutation.mutate(rule.id)} className="p-2 hover:bg-slate-800 rounded-lg transition-colors" title="Toggle">
                    {rule.is_enabled ? <ToggleRight className="w-5 h-5 text-green-400" /> : <ToggleLeft className="w-5 h-5 text-slate-500" />}
                  </button>
                  <button onClick={() => openEdit(rule)} className="p-2 hover:bg-slate-800 rounded-lg text-slate-400 transition-colors text-xs">Edit</button>
                  <button onClick={() => { if (confirm('Delete this rule?')) deleteMutation.mutate(rule.id) }} className="p-2 hover:bg-red-500/10 text-slate-500 hover:text-red-400 rounded-lg transition-colors">
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              )}
            </div>
          </div>
        ))}
        {rules.length === 0 && !isLoading && (
          <div className="text-center py-12 text-slate-500">No rules found</div>
        )}
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-slate-900 border border-slate-700 rounded-2xl w-full max-w-2xl p-6 space-y-4">
            <h2 className="text-xl font-bold text-white">{editingRule ? 'Edit Rule' : 'Create Rule'}</h2>
            <input value={form.name} onChange={e => setForm({...form, name: e.target.value})} placeholder="Rule name" className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white" />
            <input value={form.description} onChange={e => setForm({...form, description: e.target.value})} placeholder="Description" className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white" />
            <textarea value={form.sec_rule} onChange={e => setForm({...form, sec_rule: e.target.value})} placeholder={'SecRule REQUEST_URI "@rx ..." "id:10001,phase:1,deny,status:403"'} rows={6} className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white font-mono text-sm" />
            <div className="flex items-center gap-4">
              <input type="number" value={form.priority} onChange={e => setForm({...form, priority: +e.target.value})} className="w-32 bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white" placeholder="Priority" />
              <label className="flex items-center gap-2 text-slate-300 text-sm">
                <input type="checkbox" checked={form.is_enabled} onChange={e => setForm({...form, is_enabled: e.target.checked})} className="rounded" />
                Enabled
              </label>
            </div>
            <div className="flex justify-end gap-3 pt-2">
              <button onClick={() => { setShowModal(false); resetForm() }} className="px-4 py-2 text-slate-400 hover:text-white">Cancel</button>
              <button onClick={() => createMutation.mutate(form)} className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg">{editingRule ? 'Update' : 'Create'}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
