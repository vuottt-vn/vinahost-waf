export interface User {
  id: number
  username: string
  email: string
  role: 'admin' | 'user'
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Rule {
  id: number
  name: string
  description: string
  sec_rule: string
  is_enabled: boolean
  priority: number
  created_by: number
  creator?: User
  created_at: string
  updated_at: string
}

export interface AuditLog {
  id: number
  transaction_id: string
  client_ip: string
  server_ip: string
  request_uri: string
  method: string
  status_code: number
  action: 'allow' | 'block' | 'challenge'
  matched_rules: string
  anomaly_score: number
  user_agent: string
  created_at: string
}

export interface ProxyTarget {
  id: number
  name: string
  upstream_url: string
  is_enabled: boolean
  created_by: number
  created_at: string
  updated_at: string
}

export interface DashboardStats {
  total_requests: number
  blocked_requests: number
  challenged_requests: number
  allowed_requests: number
  top_attack_types: { type: string; count: number }[]
  top_blocked_ips: { ip: string; count: number }[]
  traffic_timeline: { time: string; total: number; blocked: number }[]
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: User
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  pages: number
}
