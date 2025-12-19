// =============================================================================
// AuthContext - 認証状態管理コンテキスト
// =============================================================================
// JWT トークンの保存・取得・削除と認証状態の管理を提供します。
// Native: SecureStore を使用してトークンを安全に保存
// Web: localStorage を使用

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

// プラットフォーム対応ストレージ
// Web: window.localStorage を使用
// Native: SecureStore を使用
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

const TOKEN_KEY = 'auth_token';
const USER_KEY = 'auth_user';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

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

  // ストレージから認証情報を読み込む
  const loadStoredAuth = async () => {
    try {
      const storedToken = await storage.getItem(TOKEN_KEY);
      const storedUser = await storage.getItem(USER_KEY);

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
      // ストレージに保存
      await storage.setItem(TOKEN_KEY, newToken);
      await storage.setItem(USER_KEY, JSON.stringify(newUser));

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
      // ストレージから削除
      await storage.removeItem(TOKEN_KEY);
      await storage.removeItem(USER_KEY);

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
    storage.setItem(USER_KEY, JSON.stringify(newUser)).catch(console.error);
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

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
