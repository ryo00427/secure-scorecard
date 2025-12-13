// =============================================================================
// App.tsx - アプリケーションエントリーポイント
// =============================================================================
// 認証コンテキストとReact Queryのプロバイダーを設定し、
// ナビゲーションを初期化します。

import './global.css';

import React from 'react';
import { StatusBar } from 'expo-status-bar';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider } from './src/context/AuthContext';
import AppNavigator from './src/navigation/AppNavigator';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // デフォルトで5分間キャッシュを保持
      staleTime: 5 * 60 * 1000,
      // エラー時に3回リトライ
      retry: 3,
      // バックグラウンドからフォアグラウンドに戻った時に再フェッチ
      refetchOnWindowFocus: true,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <StatusBar style="auto" />
        <AppNavigator />
      </AuthProvider>
    </QueryClientProvider>
  );
}
