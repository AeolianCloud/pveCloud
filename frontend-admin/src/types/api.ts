export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface AdminUserItem {
  id: number;
  email: string;
  role: string;
  status: string;
  email_verified: boolean;
  balance: number;
  instance_count: number;
  created_at: string;
}

export interface AdminOrderItem {
  id: number;
  user_id: number;
  amount: number;
  status: string;
  billing_cycle: string;
  config_snapshot: string;
  created_at: string;
}

export interface AdminTicketItem {
  id: number;
  user_id: number;
  user_email: string;
  title: string;
  status: string;
  priority: string;
  created_at: string;
}
