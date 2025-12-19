// =============================================================================
// API Service - API クライアント
// =============================================================================
// バックエンド API との通信を担当します。
// 認証トークンの自動付与とエラーハンドリングを提供します。

import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

// API のベース URL（環境変数または固定値）
const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

// トークン保存キー
const TOKEN_KEY = 'auth_token';

// =============================================================================
// プラットフォーム対応ストレージ
// =============================================================================
// Web: localStorage を使用
// Native (iOS/Android): SecureStore を使用

const storage = {
  async getItem(key: string): Promise<string | null> {
    if (Platform.OS === 'web' && typeof window !== 'undefined') {
      return window.localStorage.getItem(key);
    }
    return SecureStore.getItemAsync(key);
  },
  async setItem(key: string, value: string): Promise<void> {
    if (Platform.OS === 'web' && typeof window !== 'undefined') {
      window.localStorage.setItem(key, value);
      return;
    }
    await SecureStore.setItemAsync(key, value);
  },
  async removeItem(key: string): Promise<void> {
    if (Platform.OS === 'web' && typeof window !== 'undefined') {
      window.localStorage.removeItem(key);
      return;
    }
    await SecureStore.deleteItemAsync(key);
  },
};

// API レスポンスの基本型
interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

// API エラーの型
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

// 共通のリクエストヘッダーを取得
async function getHeaders(): Promise<Record<string, string>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  // 認証トークンを取得して追加
  const token = await storage.getItem(TOKEN_KEY);
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  return headers;
}

// レスポンスを処理
async function handleResponse<T>(response: Response): Promise<T> {
  const contentType = response.headers.get('content-type');
  const isJson = contentType?.includes('application/json');

  if (!response.ok) {
    let errorMessage = 'リクエストに失敗しました';
    let errorCode: string | undefined;

    if (isJson) {
      try {
        const errorData = await response.json();
        // バックエンドのエラー形式: { "error": { "code": "...", "message": "...", "details": [...] } }
        if (errorData.error && typeof errorData.error === 'object') {
          const appError = errorData.error;
          errorMessage = appError.message || errorMessage;
          errorCode = appError.code;
          // バリデーションエラーの詳細がある場合は追加
          if (appError.details && Array.isArray(appError.details)) {
            const detailMessages = appError.details
              .map((d: { field?: string; message?: string }) => d.message || d.field)
              .filter(Boolean)
              .join(', ');
            if (detailMessages) {
              errorMessage = `${errorMessage}: ${detailMessages}`;
            }
          }
        } else if (typeof errorData.message === 'string') {
          // 別の形式: { "message": "..." }
          errorMessage = errorData.message;
        } else if (typeof errorData.error === 'string') {
          // 別の形式: { "error": "..." }
          errorMessage = errorData.error;
        }
      } catch {
        // JSON パースに失敗した場合はデフォルトメッセージを使用
      }
    }

    throw new ApiError(errorMessage, response.status, errorCode);
  }

  if (isJson) {
    return response.json();
  }

  return {} as T;
}

// GET リクエスト
export async function get<T>(endpoint: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'GET',
    headers: await getHeaders(),
  });
  return handleResponse<T>(response);
}

// POST リクエスト
export async function post<T>(endpoint: string, body?: unknown): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'POST',
    headers: await getHeaders(),
    body: body ? JSON.stringify(body) : undefined,
  });
  return handleResponse<T>(response);
}

// PUT リクエスト
export async function put<T>(endpoint: string, body?: unknown): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'PUT',
    headers: await getHeaders(),
    body: body ? JSON.stringify(body) : undefined,
  });
  return handleResponse<T>(response);
}

// DELETE リクエスト
export async function del<T>(endpoint: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    method: 'DELETE',
    headers: await getHeaders(),
  });
  return handleResponse<T>(response);
}

// API オブジェクト（通知サービスなど外部で使用）
export const api = {
  get,
  post,
  put,
  delete: del,
};

interface LoginRequest {
  email: string;
  password: string;
}

interface RegisterRequest {
  email: string;
  password: string;
  display_name: string;
}

interface AuthResponse {
  token: string;
  user: {
    id: number;
    email: string;
    display_name: string;
  };
}

export const authApi = {
  // ログイン
  login: (data: LoginRequest) => post<AuthResponse>('/auth/login', data),

  // ユーザー登録
  register: (data: RegisterRequest) => post<AuthResponse>('/auth/register', data),

  // ログアウト
  logout: () => post<{ message: string }>('/auth/logout'),

  // 現在のユーザー情報を取得
  me: () => get<{ user: AuthResponse['user'] }>('/auth/me'),
};

interface Task {
  id: number;
  title: string;
  description: string;
  due_date: string;
  priority: 'low' | 'medium' | 'high';
  status: 'pending' | 'completed';
}

export const tasksApi = {
  // タスク一覧を取得（バックエンドは配列を直接返す）
  getAll: () => get<Task[]>('/tasks'),

  // 今日のタスクを取得（バックエンドは配列を直接返す）
  getToday: () => get<Task[]>('/tasks/today'),

  // 期限切れタスクを取得（バックエンドは配列を直接返す）
  getOverdue: () => get<Task[]>('/tasks/overdue'),

  // タスクを作成（バックエンドはオブジェクトを直接返す）
  create: (data: Omit<Task, 'id' | 'status'>) => post<Task>('/tasks', data),

  // タスクを完了（バックエンドはオブジェクトを直接返す）
  complete: (id: number) => post<Task>(`/tasks/${id}/complete`),

  // タスクを削除
  delete: (id: number) => del<{ message: string }>(`/tasks/${id}`),
};

interface Crop {
  id: number;
  name: string;
  variety: string;
  planted_date: string;
  expected_harvest_date: string;
  status: 'planning' | 'growing' | 'harvested';
}

export const cropsApi = {
  // 作物一覧を取得（バックエンドは配列を直接返す）
  getAll: () => get<Crop[]>('/crops'),

  // 作物詳細を取得（バックエンドはオブジェクトを直接返す）
  getById: (id: number) => get<Crop>(`/crops/${id}`),

  // 作物を作成（バックエンドはオブジェクトを直接返す）
  create: (data: Omit<Crop, 'id'>) => post<Crop>('/crops', data),

  // 作物を更新（バックエンドはオブジェクトを直接返す）
  update: (id: number, data: Partial<Crop>) => put<Crop>(`/crops/${id}`, data),

  // 作物を削除
  delete: (id: number) => del<{ message: string }>(`/crops/${id}`),
};

export const notificationApi = {
  // デバイストークンを登録
  registerDevice: (token: string, platform: 'ios' | 'android' | 'web') =>
    post<{ id: number; message: string }>('/notifications/device-token', { token, platform }),

  // デバイストークンを削除
  deleteDevice: (platform?: string) =>
    del<{ message: string }>(`/notifications/device-token${platform ? `?platform=${platform}` : ''}`),

  // 通知設定を取得
  getSettings: () =>
    get<{
      push_enabled: boolean;
      email_enabled: boolean;
      task_reminders: boolean;
      harvest_reminders: boolean;
      growth_record_notifications: boolean;
    }>('/users/settings/notifications'),

  // 通知設定を更新
  updateSettings: (settings: {
    push_enabled?: boolean;
    email_enabled?: boolean;
    task_reminders?: boolean;
    harvest_reminders?: boolean;
    growth_record_notifications?: boolean;
  }) => put<{ message: string }>('/users/settings/notifications', settings),
};

// 収穫サマリー
interface HarvestSummary {
  crop_id: number;
  crop_name: string;
  total_quantity: number;
  average_growth_days: number;
  harvest_count: number;
}

// グラフデータポイント
interface ChartDataPoint {
  label: string;
  value: number;
}

// グラフデータ
interface ChartData {
  chart_type: string;
  title: string;
  data: ChartDataPoint[];
}

// CSVエクスポートレスポンス
interface ExportResponse {
  download_url: string;
  expires_at: string;
}

export const analyticsApi = {
  // 収穫サマリーを取得
  getHarvestSummary: (params?: {
    start_date?: string;
    end_date?: string;
    crop_id?: number;
  }) => {
    const queryParams = new URLSearchParams();
    if (params?.start_date) queryParams.append('start_date', params.start_date);
    if (params?.end_date) queryParams.append('end_date', params.end_date);
    if (params?.crop_id) queryParams.append('crop_id', params.crop_id.toString());
    const query = queryParams.toString();
    return get<{ summaries: HarvestSummary[] }>(
      `/analytics/harvest-summary${query ? `?${query}` : ''}`
    );
  },

  // グラフデータを取得
  getChartData: (chartType: 'monthly_harvest' | 'crop_comparison' | 'plot_productivity') =>
    get<ChartData>(`/analytics/charts/${chartType}`),

  // CSVエクスポート
  exportCSV: (dataType: 'crops' | 'harvests' | 'tasks' | 'all') =>
    get<ExportResponse>(`/analytics/export/${dataType}`),
};

// 型エクスポート
export type { HarvestSummary, ChartData, ChartDataPoint, ExportResponse };
