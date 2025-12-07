// =============================================================================
// AuthContext - 認証状態管理コンテキスト
// =============================================================================
// JWT トークンの保存・取得・削除と認証状態の管理を提供します。
// SecureStore を使用してトークンを安全に保存します。

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import * as SecureStore from 'expo-secure-store';

// -----------------------------------------------------------------------------
// Types - 型定義
// -----------------------------------------------------------------------------

// ユーザー情報
interface User {
  id: number;
  email: string;
  displayName: string;
}

// 認証コンテキストの型
interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  login: (token: string, user: User) => Promise<void>;
  logout: () => Promise<void>;
  updateUser: (user: User) => void;
}

// -----------------------------------------------------------------------------
// Constants - 定数
// -----------------------------------------------------------------------------

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'auth_user';

// -----------------------------------------------------------------------------
// Context - コンテキスト
// -----------------------------------------------------------------------------

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// -----------------------------------------------------------------------------
// Provider - プロバイダーコンポーネント
// -----------------------------------------------------------------------------

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // アプリ起動時に保存済みの認証情報を読み込む
  useEffect(() => {
    loadStoredAuth();
  }, []);

  // SecureStore から認証情報を読み込む
  const loadStoredAuth = async () => {
    try {
      const storedToken = await SecureStore.getItemAsync(TOKEN_KEY);
      const storedUser = await SecureStore.getItemAsync(USER_KEY);

      if (storedToken && storedUser) {
        setToken(storedToken);
        setUser(JSON.parse(storedUser));
      }
    } catch (error) {
      console.error('認証情報の読み込みに失敗しました:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // ログイン処理
  const login = async (newToken: string, newUser: User) => {
    try {
      // SecureStore に保存
      await SecureStore.setItemAsync(TOKEN_KEY, newToken);
      await SecureStore.setItemAsync(USER_KEY, JSON.stringify(newUser));

      // 状態を更新
      setToken(newToken);
      setUser(newUser);
    } catch (error) {
      console.error('ログイン情報の保存に失敗しました:', error);
      throw error;
    }
  };

  // ログアウト処理
  const logout = async () => {
    try {
      // SecureStore から削除
      await SecureStore.deleteItemAsync(TOKEN_KEY);
      await SecureStore.deleteItemAsync(USER_KEY);

      // 状態をクリア
      setToken(null);
      setUser(null);
    } catch (error) {
      console.error('ログアウトに失敗しました:', error);
      throw error;
    }
  };

  // ユーザー情報を更新
  const updateUser = (newUser: User) => {
    setUser(newUser);
    SecureStore.setItemAsync(USER_KEY, JSON.stringify(newUser)).catch(console.error);
  };

  const value: AuthContextType = {
    user,
    token,
    isLoading,
    isAuthenticated: !!token && !!user,
    login,
    logout,
    updateUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// -----------------------------------------------------------------------------
// Hook - カスタムフック
// -----------------------------------------------------------------------------

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
