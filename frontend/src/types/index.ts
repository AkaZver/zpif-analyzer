export interface Fund {
  id: number;
  name: string;
  isin: string;
  ticker: string;
  management_company: string;
  real_estate_segment: string;
  qualified_required: boolean;
  has_market_maker: boolean;
  fund_start_date: string | null;
  fund_end_date: string | null;
  investfunds_url: string;
  created_at: string;
  updated_at: string;
}

export interface FundFinancials {
  id: number;
  fund_id: number;
  snapshot_date: string;
  
  // Цены и стоимость
  unit_price_rub: number;
  nav_per_unit_rub: number;
  nav_total_mln_rub: number;
  discount_to_nav_pct: number;
  
  // Ключевые метрики
  cap_rate_pct: number;
  p_nav: number;
  p_affo: number;
  noi_yield_pct: number;
  
  // Выплаты
  annual_payout_rub: number;
  payout_yield_pct: number;
  payout_yield_after_tax_pct: number;
  total_return_pct: number;
  payout_frequency: string;
  payout_stability: string;
  rent_indexation_pct: number;
  
  // Долг и операции
  debt_to_nav_ratio: number;
  management_fee_pct: number;
  trading_volume_mln_rub: number;
  number_of_properties: number;
  main_tenants: string;
  
  // Прогнозы
  irr_forecast_pct: number;
  
  created_at: string;
  updated_at: string;
}

export interface FundDocument {
  id: number;
  fund_id: number;
  file_name: string;
  file_path: string;
  document_type: string;
  content_hash: string;
  source: 'manual' | 'auto';
  source_url: string;
  upload_date: string;
  status: 'pending' | 'downloaded' | 'analyzed' | 'error';
  created_at: string;
  updated_at: string;
}

export interface LLMAnalysis {
  id: number;
  fund_id: number;
  document_id: number | null;
  model_used: string;
  raw_response: string;
  analysis_summary: string;
  risk_assessment: string;
  pros_cons: string;
  extracted_metrics: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface LLMSettings {
  id: number;
  api_key_encrypted: string;
  base_url: string;
  model_name: string;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: number;
  username: string;
  email: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}
