import { useQuery } from '@tanstack/react-query'
import api from '../lib/api'
import { Activity, ShieldAlert, ShieldCheck, ShieldQuestion, TrendingUp } from 'lucide-react'
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import type { DashboardStats } from '../types'

export default function Dashboard() {
  const { data: stats, isLoading } = useQuery<DashboardStats>({
    queryKey: ['dashboard-stats'],
    queryFn: () => api.get('/dashboard/stats').then(r => r.data),
    refetchInterval: 10000,
  })

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full" />
      </div>
    )
  }

  const statCards = [
    { label: 'Total Requests', value: stats?.total_requests || 0, icon: Activity, color: 'text-blue-400', bg: 'bg-blue-500/10' },
    { label: 'Allowed', value: stats?.allowed_requests || 0, icon: ShieldCheck, color: 'text-green-400', bg: 'bg-green-500/10' },
    { label: 'Blocked', value: stats?.blocked_requests || 0, icon: ShieldAlert, color: 'text-red-400', bg: 'bg-red-500/10' },
    { label: 'Challenged', value: stats?.challenged_requests || 0, icon: ShieldQuestion, color: 'text-amber-400', bg: 'bg-amber-500/10' },
  ]

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-white">Dashboard</h1>
        <p className="text-slate-400 mt-1">Overview of your web application firewall</p>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((card) => (
          <div key={card.label} className="bg-slate-900 border border-slate-800 rounded-xl p-5">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-slate-400">{card.label}</p>
                <p className="text-3xl font-bold text-white mt-1">{card.value.toLocaleString()}</p>
              </div>
              <div className={`w-12 h-12 ${card.bg} rounded-xl flex items-center justify-center`}>
                <card.icon className={`w-6 h-6 ${card.color}`} />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Traffic timeline */}
      <div className="bg-slate-900 border border-slate-800 rounded-xl p-6">
        <h2 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <TrendingUp className="w-5 h-5 text-blue-400" />
          Traffic Timeline (24h)
        </h2>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={stats?.traffic_timeline || []}>
            <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
            <XAxis dataKey="time" stroke="#64748b" fontSize={12} />
            <YAxis stroke="#64748b" fontSize={12} />
            <Tooltip
              contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px' }}
              labelStyle={{ color: '#e2e8f0' }}
            />
            <Line type="monotone" dataKey="total" stroke="#3b82f6" strokeWidth={2} dot={false} name="Total" />
            <Line type="monotone" dataKey="blocked" stroke="#ef4444" strokeWidth={2} dot={false} name="Blocked" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top blocked IPs */}
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4">Top Blocked IPs</h2>
          <div className="space-y-3">
            {(stats?.top_blocked_ips || []).length === 0 && (
              <p className="text-slate-500 text-sm">No blocked IPs yet</p>
            )}
            {(stats?.top_blocked_ips || []).map((item, i) => (
              <div key={item.ip} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className="text-sm text-slate-500 w-6">{i + 1}.</span>
                  <span className="text-sm font-mono text-slate-300">{item.ip}</span>
                </div>
                <span className="text-sm font-medium text-red-400">{item.count}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Top attack types */}
        <div className="bg-slate-900 border border-slate-800 rounded-xl p-6">
          <h2 className="text-lg font-semibold text-white mb-4">Top Attack Types</h2>
          <ResponsiveContainer width="100%" height={200}>
            <BarChart data={stats?.top_attack_types || []} layout="vertical">
              <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
              <XAxis type="number" stroke="#64748b" fontSize={12} />
              <YAxis type="category" dataKey="type" stroke="#64748b" fontSize={11} width={120} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1e293b', border: '1px solid #334155', borderRadius: '8px' }}
              />
              <Bar dataKey="count" fill="#3b82f6" radius={[0, 4, 4, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  )
}
