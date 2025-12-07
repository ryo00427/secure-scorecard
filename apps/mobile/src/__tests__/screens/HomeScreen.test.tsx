// =============================================================================
// HomeScreen テスト
// =============================================================================
// ホーム画面のレンダリングと基本的な機能をテストします。

import React from 'react';
import { render, screen, waitFor } from '@testing-library/react-native';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import HomeScreen from '../../screens/main/HomeScreen';
import { AuthProvider } from '../../context/AuthContext';

// -----------------------------------------------------------------------------
// Mocks - モック
// -----------------------------------------------------------------------------

// API モック
jest.mock('../../services/api', () => ({
  tasksApi: {
    getToday: jest.fn().mockResolvedValue({
      tasks: [
        {
          id: 1,
          title: 'テストタスク1',
          priority: 'high',
          status: 'pending',
        },
        {
          id: 2,
          title: 'テストタスク2',
          priority: 'medium',
          status: 'pending',
        },
      ],
    }),
  },
  cropsApi: {
    getAll: jest.fn().mockResolvedValue({
      crops: [
        {
          id: 1,
          name: 'トマト',
          variety: 'ミニトマト',
          status: 'growing',
        },
        {
          id: 2,
          name: 'キュウリ',
          variety: '夏すずみ',
          status: 'growing',
        },
      ],
    }),
  },
}));

// SecureStore モック
jest.mock('expo-secure-store', () => ({
  getItemAsync: jest.fn().mockResolvedValue(null),
  setItemAsync: jest.fn().mockResolvedValue(undefined),
  deleteItemAsync: jest.fn().mockResolvedValue(undefined),
}));

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

describe('HomeScreen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('ホーム画面がレンダリングされる', async () => {
    renderWithProviders(<HomeScreen />);

    // ウェルカムテキストが表示される
    await waitFor(() => {
      expect(screen.getByText('こんにちは、')).toBeTruthy();
    });
  });

  it('今日のタスクセクションが表示される', async () => {
    renderWithProviders(<HomeScreen />);

    await waitFor(() => {
      expect(screen.getByText('今日のタスク')).toBeTruthy();
    });
  });

  it('栽培中の作物セクションが表示される', async () => {
    renderWithProviders(<HomeScreen />);

    await waitFor(() => {
      expect(screen.getByText('栽培中の作物')).toBeTruthy();
    });
  });

  it('タスクデータが正しく表示される', async () => {
    renderWithProviders(<HomeScreen />);

    await waitFor(() => {
      expect(screen.getByText('テストタスク1')).toBeTruthy();
    });
  });

  it('作物データが正しく表示される', async () => {
    renderWithProviders(<HomeScreen />);

    await waitFor(() => {
      expect(screen.getByText('トマト')).toBeTruthy();
      expect(screen.getByText('ミニトマト')).toBeTruthy();
    });
  });
});
