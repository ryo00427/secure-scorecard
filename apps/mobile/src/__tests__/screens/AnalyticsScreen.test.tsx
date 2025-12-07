// =============================================================================
// AnalyticsScreen テスト
// =============================================================================
// 分析画面のレンダリングと機能をテストします。

import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react-native';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import AnalyticsScreen from '../../screens/main/AnalyticsScreen';

// -----------------------------------------------------------------------------
// Mocks - モック
// -----------------------------------------------------------------------------

// API モック
const mockGetHarvestSummary = jest.fn().mockResolvedValue({
  summaries: [
    {
      crop_id: 1,
      crop_name: 'トマト',
      total_quantity: 50,
      average_growth_days: 90,
      harvest_count: 10,
    },
    {
      crop_id: 2,
      crop_name: 'キュウリ',
      total_quantity: 30,
      average_growth_days: 60,
      harvest_count: 15,
    },
  ],
});

const mockGetChartData = jest.fn().mockResolvedValue({
  chart_type: 'monthly_harvest',
  title: '月別収穫量',
  data: [
    { label: '1月', value: 10 },
    { label: '2月', value: 15 },
    { label: '3月', value: 20 },
  ],
});

const mockExportCSV = jest.fn().mockResolvedValue({
  download_url: 'https://example.com/export.csv',
  expires_at: '2024-01-15T12:00:00Z',
});

jest.mock('../../services/api', () => ({
  analyticsApi: {
    getHarvestSummary: () => mockGetHarvestSummary(),
    getChartData: (type: string) => mockGetChartData(type),
    exportCSV: (type: string) => mockExportCSV(type),
  },
}));

// react-native-chart-kit モック
jest.mock('react-native-chart-kit', () => ({
  BarChart: 'BarChart',
  PieChart: 'PieChart',
}));

// Share API モック
jest.mock('react-native', () => {
  const RN = jest.requireActual('react-native');
  return {
    ...RN,
    Share: {
      share: jest.fn().mockResolvedValue({ action: 'sharedAction' }),
    },
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
    <QueryClientProvider client={queryClient}>{component}</QueryClientProvider>
  );
};

// -----------------------------------------------------------------------------
// Tests - テスト
// -----------------------------------------------------------------------------

describe('AnalyticsScreen', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('分析画面がレンダリングされる', async () => {
    renderWithProviders(<AnalyticsScreen />);

    await waitFor(() => {
      expect(screen.getByText('分析')).toBeTruthy();
    });
  });

  it('収穫サマリーセクションが表示される', async () => {
    renderWithProviders(<AnalyticsScreen />);

    await waitFor(() => {
      expect(screen.getByText('収穫サマリー')).toBeTruthy();
    });
  });

  it('グラフセクションが表示される', async () => {
    renderWithProviders(<AnalyticsScreen />);

    await waitFor(() => {
      expect(screen.getByText('グラフ')).toBeTruthy();
    });
  });

  it('データエクスポートセクションが表示される', async () => {
    renderWithProviders(<AnalyticsScreen />);

    await waitFor(() => {
      expect(screen.getByText('データエクスポート')).toBeTruthy();
    });
  });

  it('総収穫量が表示される', async () => {
    renderWithProviders(<AnalyticsScreen />);

    await waitFor(() => {
      // 50 + 30 = 80 kg
      expect(screen.getByText('80 kg')).toBeTruthy();
    });
  });

  it('収穫回数が表示される', async () => {
    renderWithProviders(<AnalyticsScreen />);

    await waitFor(() => {
      // 10 + 15 = 25 回
      expect(screen.getByText('25 回')).toBeTruthy();
    });
  });

  it('グラフタイプ切り替えボタンが表示される', async () => {
    renderWithProviders(<AnalyticsScreen />);

    await waitFor(() => {
      expect(screen.getByText('月別収穫量')).toBeTruthy();
      expect(screen.getByText('作物比較')).toBeTruthy();
      expect(screen.getByText('区画生産性')).toBeTruthy();
    });
  });

  it('エクスポートボタンが表示される', async () => {
    renderWithProviders(<AnalyticsScreen />);

    await waitFor(() => {
      expect(screen.getByText('作物データ')).toBeTruthy();
      expect(screen.getByText('収穫データ')).toBeTruthy();
      expect(screen.getByText('タスクデータ')).toBeTruthy();
      expect(screen.getByText('全データ')).toBeTruthy();
    });
  });
});
