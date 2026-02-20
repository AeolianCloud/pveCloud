export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface UserProfile {
  id: number;
  email: string;
  role: string;
}

export interface WalletInfo {
  id: number;
  user_id: number;
  balance: number;
  frozen_balance: number;
}

export interface UserOrder {
  id: number;
  user_id: number;
  product_id: number;
  amount: number;
  billing_cycle: string;
  status: string;
  config_snapshot: string;
  created_at: string;
}

export interface TaskStatus {
  id: number;
  status: string;
  progress: number;
  message: string;
}
