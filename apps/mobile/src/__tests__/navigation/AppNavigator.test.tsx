// =============================================================================
// AppNavigator テスト
// =============================================================================
// アプリナビゲーションのレンダリングと画面遷移をテストします。

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react-native';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AppNavigator from '../../navigation/AppNavigator';
import { AuthProvider } from '../../context/AuthContext';

// -----------------------------------------------------------------------------
// Mocks - モック
// -----------------------------------------------------------------------------

// SecureStore モック - 未認証状態
const mockGetItemAsync = jest.fn().mockResolvedValue(null);
const mockSetItemAsync = jest.fn().mockResolvedValue(undefined);
const mockDeleteItemAsync = jest.fn().mockResolvedValue(undefined);

jest.mock('expo-secure-store', () => ({
  getItemAsync: () => mockGetItemAsync(),
  setItemAsync: (...args: unknown[]) => mockSetItemAsync(...args),
  deleteItemAsync: (...args: unknown[]) => mockDeleteItemAsync(...args),
}));

// API モック
jest.mock('../../services/api', () => ({
  tasksApi: {
    getToday: jest.fn().mockResolvedValue({ tasks: [] }),
  },
  cropsApi: {
    getAll: jest.fn().mockResolvedValue({ crops: [] }),
  },
}));

// react-native-screens モック
jest.mock('react-native-screens', () => {
  const actual = jest.requireActual('react-native-screens');
  return {
    ...actual,
    enableScreens: jest.fn(),
  };
});

// react-native-safe-area-context モック
jest.mock('react-native-safe-area-context', () => {
  const React = require('react');
  const { View } = require('react-native');
  return {
    SafeAreaProvider: ({ children }: { children: React.ReactNode }) => (
      <View>{children}</View>
    ),
    SafeAreaView: ({ children }: { children: React.ReactNode }) => (
      <View>{children}</View>
    ),
    useSafeAreaInsets: () => ({ top: 0, right: 0, bottom: 0, left: 0 }),
  };
});

// -----------------------------------------------------------------------------
// Test Utilities - テストユーティリティ
// -----------------------------------------------------------------------------

const createTestQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
    },
  });

const renderWithProviders = (component: React.ReactElement) => {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <AuthProvider>{component}</AuthProvider>
    </QueryClientProvider>
  );
};

// -----------------------------------------------------------------------------
// Tests - テスト
// -----------------------------------------------------------------------------

describe('AppNavigator', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('未認証時はログイン画面が表示される', async () => {
    // 未認証状態をモック
    mockGetItemAsync.mockResolvedValue(null);

    renderWithProviders(<AppNavigator />);

    // ローディングが完了するまで待機
    await waitFor(
      () => {
        // ログイン画面のテキストが表示されることを確認
        expect(screen.getByText('ログイン')).toBeTruthy();
      },
      { timeout: 5000 }
    );
  });

  it('認証済みの場合はホーム画面が表示される', async () => {
    // 認証済み状態をモック
    mockGetItemAsync
      .mockResolvedValueOnce('test-token')
      .mockResolvedValueOnce(
        JSON.stringify({
          id: 1,
          email: 'test@example.com',
          displayName: 'テストユーザー',
        })
      );

    renderWithProviders(<AppNavigator />);

    // ローディングが完了してホーム画面が表示されるまで待機
    await waitFor(
      () => {
        expect(screen.getByText('ホーム')).toBeTruthy();
      },
      { timeout: 5000 }
    );
  });

  it('ナビゲーションタブが表示される', async () => {
    // 認証済み状態をモック
    mockGetItemAsync
      .mockResolvedValueOnce('test-token')
      .mockResolvedValueOnce(
        JSON.stringify({
          id: 1,
          email: 'test@example.com',
          displayName: 'テストユーザー',
        })
      );

    renderWithProviders(<AppNavigator />);

    await waitFor(
      () => {
        expect(screen.getByText('ホーム')).toBeTruthy();
        expect(screen.getByText('タスク')).toBeTruthy();
        expect(screen.getByText('作物')).toBeTruthy();
        expect(screen.getByText('分析')).toBeTruthy();
        expect(screen.getByText('設定')).toBeTruthy();
      },
      { timeout: 5000 }
    );
  });
});
