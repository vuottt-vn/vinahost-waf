import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import api from '../lib/api'
import { useAuthStore } from '../store/auth'
import { Save } from 'lucide-react'

export default function Settings() {
  const isAdmin = useAuthStore((s) => s.isAdmin)
  const [settings, setSettings] = useState({
    mode: 'on',
    challenge_enabled: true,
    challenge_threshold: 5,
    challenge_difficulty: 3,
    request_body_limit: 131072,
    response_body_limit: 131072,
  })
  const [saved, setSaved] = useState(false)

  const { isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api.get('/settings').then(r => {
      setSettings(r.data)
      return r.data
    }),
  })

  const saveMutation = useMutation({
    mutationFn: () => api.put('/settings', settings),
    onSuccess: () => {
      setSaved(true)
      setTimeout(() => setSaved(false), 3000)
    },
  })

  if (isLoading) return <div className="flex items-center justify-center h-64"><div className="animate-spin w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full" /></div>

  return (
    <div className="space-y-6 max-w-2xl">
      <div>
        <h1 className="text-2xl font-bold text-white">WAF Settings</h1>
        <p className="text-slate-400 mt-1">Configure the web application firewall</p>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-xl p-6 space-y-6">
        {/* WAF Mode */}
        <div>
          <label className="block text-sm font-medium text-slate-300 mb-2">WAF Mode</label>
          <select
            value={settings.mode}
            onChange={e => setSettings({...settings, mode: e.target.value})}
            disabled={!isAdmin}
            className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2.5 text-white disabled:opacity-50"
          >
            <option value="on">Active (Block & Detect)</option>
            <option value="detection_only">Detection Only (Log, don't block)</option>
            <option value="off">Disabled</option>
          </select>
        </div>

        {/* Challenge settings */}
        <div className="border-t border-slate-800 pt-6">
          <h3 className="font-semibold text-white mb-4">JS Challenge Settings</h3>
          <div className="space-y-4">
            <label className="flex items-center gap-3">
              <input
                type="checkbox"
                checked={settings.challenge_enabled}
                onChange={e => setSettings({...settings, challenge_enabled: e.target.checked})}
                disabled={!isAdmin}
                className="rounded"
              />
              <span className="text-sm text-slate-300">Enable JS Challenge for suspicious requests</span>
            </label>

            <div>
              <label className="block text-sm text-slate-400 mb-1">Challenge Threshold (anomaly score)</label>
              <input
                type="number"
                min={1}
                max={20}
                value={settings.challenge_threshold}
                onChange={e => setSettings({...settings, challenge_threshold: +e.target.value})}
                disabled={!isAdmin}
                className="w-32 bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white disabled:opacity-50"
              />
            </div>

            <div>
              <label className="block text-sm text-slate-400 mb-1">Challenge Difficulty (1-5)</label>
              <input
                type="range"
                min={1}
                max={5}
                value={settings.challenge_difficulty}
                onChange={e => setSettings({...settings, challenge_difficulty: +e.target.value})}
                disabled={!isAdmin}
                className="w-48"
              />
              <span className="ml-3 text-sm text-slate-300">{settings.challenge_difficulty}</span>
            </div>
          </div>
        </div>

        {/* Body limits */}
        <div className="border-t border-slate-800 pt-6">
          <h3 className="font-semibold text-white mb-4">Body Size Limits</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Request Body (bytes)</label>
              <input
                type="number"
                value={settings.request_body_limit}
                onChange={e => setSettings({...settings, request_body_limit: +e.target.value})}
                disabled={!isAdmin}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white disabled:opacity-50"
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400 mb-1">Response Body (bytes)</label>
              <input
                type="number"
                value={settings.response_body_limit}
                onChange={e => setSettings({...settings, response_body_limit: +e.target.value})}
                disabled={!isAdmin}
                className="w-full bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-white disabled:opacity-50"
              />
            </div>
          </div>
        </div>

        {isAdmin && (
          <div className="border-t border-slate-800 pt-6 flex items-center gap-4">
            <button
              onClick={() => saveMutation.mutate()}
              className="flex items-center gap-2 px-6 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
            >
              <Save className="w-4 h-4" /> Save Settings
            </button>
            {saved && <span className="text-sm text-green-400">Settings saved successfully!</span>}
          </div>
        )}
      </div>
    </div>
  )
}
