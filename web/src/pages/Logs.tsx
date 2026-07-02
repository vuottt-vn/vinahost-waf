import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import api from '../lib/api'
import { Search, ChevronLeft, ChevronRight } from 'lucide-react'
import type { AuditLog, PaginatedResponse } from '../types'

export default function Logs() {
  const [page, setPage] = useState(1)
  const [action, setAction] = useState('')
  const [clientIp, setClientIp] = useState('')
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null)

  const { data, isLoading } = useQuery<PaginatedResponse<AuditLog>>({
    queryKey: ['logs', page, action, clientIp],
    queryFn: () => api.get('/logs', { params: { page, page_size: 25, action, client_ip: clientIp } }).then(r => r.data),
  })

  const logs = data?.data || []
  const totalPages = data?.pages || 1

  const actionBadge = (action: string) => {
    const styles: Record<string, string> = {
      allow: 'bg-green-500/20 text-green-400',
      block: 'bg-red-500/20 text-red-400',
      challenge: 'bg-amber-500/20 text-amber-400',
    }
    return styles[action] || 'bg-slate-700 text-slate-300'
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-white">Audit Logs</h1>
        <p className="text-slate-400 mt-1">View firewall request logs</p>
      </div>

      <div className="flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
          <input value={clientIp} onChange={e => { setClientIp(e.target.value); setPage(1) }} placeholder="Filter by IP..." className="w-full bg-slate-900 border border-slate-800 rounded-lg pl-10 pr-4 py-2 text-sm text-white" />
        </div>
        <select value={action} onChange={e => { setAction(e.target.value); setPage(1) }} className="bg-slate-900 border border-slate-800 rounded-lg px-4 py-2 text-sm text-white">
          <option value="">All Actions</option>
          <option value="allow">Allow</option>
          <option value="block">Block</option>
          <option value="challenge">Challenge</option>
        </select>
      </div>

      <div className="bg-slate-900 border border-slate-800 rounded-xl overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-slate-800/50">
              <tr>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-400 uppercase">Time</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-400 uppercase">Client IP</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-400 uppercase">Method</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-400 uppercase">URI</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-400 uppercase">Status</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-400 uppercase">Action</th>
                <th className="text-left px-4 py-3 text-xs font-medium text-slate-400 uppercase">Score</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {logs.map((log) => (
                <tr key={log.id} onClick={() => setSelectedLog(log)} className="hover:bg-slate-800/30 cursor-pointer">
                  <td className="px-4 py-3 text-xs text-slate-400 whitespace-nowrap">{new Date(log.created_at).toLocaleString()}</td>
                  <td className="px-4 py-3 text-sm font-mono text-slate-300">{log.client_ip}</td>
                  <td className="px-4 py-3"><span className="px-2 py-0.5 text-xs bg-slate-800 text-slate-300 rounded">{log.method}</span></td>
                  <td className="px-4 py-3 text-sm text-slate-400 max-w-xs truncate">{log.request_uri}</td>
                  <td className="px-4 py-3 text-sm text-slate-300">{log.status_code}</td>
                  <td className="px-4 py-3"><span className={`px-2 py-0.5 text-xs rounded-full ${actionBadge(log.action)}`}>{log.action}</span></td>
                  <td className="px-4 py-3 text-sm text-slate-400">{log.anomaly_score}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        <div className="flex items-center justify-between px-4 py-3 border-t border-slate-800">
          <p className="text-sm text-slate-400">{data?.total || 0} total logs</p>
          <div className="flex items-center gap-2">
            <button onClick={() => setPage(Math.max(1, page - 1))} disabled={page <= 1} className="p-2 text-slate-400 hover:text-white disabled:opacity-30"><ChevronLeft className="w-4 h-4" /></button>
            <span className="text-sm text-slate-400">Page {page} of {totalPages}</span>
            <button onClick={() => setPage(Math.min(totalPages, page + 1))} disabled={page >= totalPages} className="p-2 text-slate-400 hover:text-white disabled:opacity-30"><ChevronRight className="w-4 h-4" /></button>
          </div>
        </div>
      </div>

      {/* Log detail modal */}
      {selectedLog && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4" onClick={() => setSelectedLog(null)}>
          <div className="bg-slate-900 border border-slate-700 rounded-2xl w-full max-w-2xl p-6 space-y-4 max-h-[80vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
            <h2 className="text-xl font-bold text-white">Log Detail</h2>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div><span className="text-slate-500">Transaction ID:</span><p className="text-slate-300 font-mono">{selectedLog.transaction_id}</p></div>
              <div><span className="text-slate-500">Time:</span><p className="text-slate-300">{new Date(selectedLog.created_at).toLocaleString()}</p></div>
              <div><span className="text-slate-500">Client IP:</span><p className="text-slate-300 font-mono">{selectedLog.client_ip}</p></div>
              <div><span className="text-slate-500">Server:</span><p className="text-slate-300">{selectedLog.server_ip}</p></div>
              <div><span className="text-slate-500">Method:</span><p className="text-slate-300">{selectedLog.method}</p></div>
              <div><span className="text-slate-500">Status:</span><p className="text-slate-300">{selectedLog.status_code}</p></div>
              <div className="col-span-2"><span className="text-slate-500">URI:</span><p className="text-slate-300 break-all">{selectedLog.request_uri}</p></div>
              <div><span className="text-slate-500">Action:</span><p className={`font-medium ${selectedLog.action === 'block' ? 'text-red-400' : selectedLog.action === 'challenge' ? 'text-amber-400' : 'text-green-400'}`}>{selectedLog.action}</p></div>
              <div><span className="text-slate-500">Anomaly Score:</span><p className="text-slate-300">{selectedLog.anomaly_score}</p></div>
              <div className="col-span-2"><span className="text-slate-500">User Agent:</span><p className="text-slate-400 text-xs">{selectedLog.user_agent}</p></div>
              {selectedLog.matched_rules && (
                <div className="col-span-2">
                  <span className="text-slate-500">Matched Rules:</span>
                  <pre className="mt-1 p-3 bg-slate-950 rounded-lg text-xs text-red-400 font-mono whitespace-pre-wrap">{selectedLog.matched_rules}</pre>
                </div>
              )}
            </div>
            <div className="flex justify-end">
              <button onClick={() => setSelectedLog(null)} className="px-4 py-2 text-slate-400 hover:text-white">Close</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
