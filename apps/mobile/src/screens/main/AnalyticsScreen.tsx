// =============================================================================
// AnalyticsScreen - 分析画面
// =============================================================================
// 収穫量グラフ、作物比較、CSVエクスポート機能を提供します。
// react-native-chart-kit を使用してグラフを表示します。

import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
  Dimensions,
  Alert,
  Share,
  ActivityIndicator,
} from 'react-native';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Ionicons } from '@expo/vector-icons';
import { BarChart, PieChart } from 'react-native-chart-kit';
import { analyticsApi, HarvestSummary, ChartData } from '../../services/api';

// -----------------------------------------------------------------------------
// Constants - 定数
// -----------------------------------------------------------------------------

const screenWidth = Dimensions.get('window').width;

// グラフ共通設定
const chartConfig = {
  backgroundColor: '#ffffff',
  backgroundGradientFrom: '#ffffff',
  backgroundGradientTo: '#ffffff',
  decimalPlaces: 0,
  color: (opacity = 1) => `rgba(22, 163, 74, ${opacity})`, // primary-600
  labelColor: (opacity = 1) => `rgba(75, 85, 99, ${opacity})`, // gray-600
  style: {
    borderRadius: 16,
  },
  propsForBackgroundLines: {
    strokeDasharray: '', // 実線
    stroke: '#e5e7eb', // gray-200
  },
};

// 円グラフ用の色
const pieChartColors = [
  '#16a34a', // primary-600
  '#22c55e', // green-500
  '#4ade80', // green-400
  '#86efac', // green-300
  '#bbf7d0', // green-200
  '#dcfce7', // green-100
];

// -----------------------------------------------------------------------------
// Types - 型定義
// -----------------------------------------------------------------------------

type ChartType = 'monthly_harvest' | 'crop_comparison' | 'plot_productivity';
type ExportType = 'crops' | 'harvests' | 'tasks' | 'all';

// -----------------------------------------------------------------------------
// Component - コンポーネント
// -----------------------------------------------------------------------------

export default function AnalyticsScreen() {
  const [selectedChart, setSelectedChart] = useState<ChartType>('monthly_harvest');

  // 収穫サマリーを取得
  const {
    data: summaryData,
    isLoading: summaryLoading,
    refetch: refetchSummary,
  } = useQuery({
    queryKey: ['analytics', 'harvest-summary'],
    queryFn: () => analyticsApi.getHarvestSummary(),
  });

  // グラフデータを取得
  const {
    data: chartData,
    isLoading: chartLoading,
    refetch: refetchChart,
  } = useQuery({
    queryKey: ['analytics', 'charts', selectedChart],
    queryFn: () => analyticsApi.getChartData(selectedChart),
  });

  // CSVエクスポート
  const exportMutation = useMutation({
    mutationFn: (dataType: ExportType) => analyticsApi.exportCSV(dataType),
    onSuccess: async (data) => {
      try {
        await Share.share({
          url: data.download_url,
          title: 'データエクスポート',
          message: `ダウンロードリンク: ${data.download_url}`,
        });
      } catch (error) {
        Alert.alert('エラー', 'データの共有に失敗しました');
      }
    },
    onError: (error) => {
      Alert.alert('エラー', 'エクスポートに失敗しました');
      console.error('Export error:', error);
    },
  });

  const isLoading = summaryLoading || chartLoading;

  const onRefresh = () => {
    refetchSummary();
    refetchChart();
  };

  // ---------------------------------------------------------------------------
  // Render Helpers - レンダリングヘルパー
  // ---------------------------------------------------------------------------

  // グラフタイプ選択ボタン
  const renderChartTypeButton = (type: ChartType, label: string, icon: string) => (
    <TouchableOpacity
      key={type}
      onPress={() => setSelectedChart(type)}
      className={`mr-2 flex-row items-center rounded-full px-4 py-2 ${
        selectedChart === type ? 'bg-primary-600' : 'bg-gray-200'
      }`}
    >
      <Ionicons
        name={icon as keyof typeof Ionicons.glyphMap}
        size={16}
        color={selectedChart === type ? '#ffffff' : '#4b5563'}
      />
      <Text
        className={`ml-2 text-sm font-medium ${
          selectedChart === type ? 'text-white' : 'text-gray-600'
        }`}
      >
        {label}
      </Text>
    </TouchableOpacity>
  );

  // 棒グラフをレンダリング
  const renderBarChart = (data: ChartData) => {
    if (!data.data || data.data.length === 0) {
      return (
        <View className="items-center justify-center py-8">
          <Text className="text-gray-500">データがありません</Text>
        </View>
      );
    }

    const barData = {
      labels: data.data.map((d) => d.label.substring(0, 6)),
      datasets: [
        {
          data: data.data.map((d) => d.value),
        },
      ],
    };

    return (
      <BarChart
        data={barData}
        width={screenWidth - 48}
        height={220}
        chartConfig={chartConfig}
        style={{
          marginVertical: 8,
          borderRadius: 16,
        }}
        yAxisLabel=""
        yAxisSuffix="kg"
        fromZero
      />
    );
  };

  // 円グラフをレンダリング
  const renderPieChart = (summaries: HarvestSummary[]) => {
    if (!summaries || summaries.length === 0) {
      return (
        <View className="items-center justify-center py-8">
          <Text className="text-gray-500">データがありません</Text>
        </View>
      );
    }

    const pieData = summaries.slice(0, 6).map((summary, index) => ({
      name: summary.crop_name,
      quantity: summary.total_quantity,
      color: pieChartColors[index % pieChartColors.length],
      legendFontColor: '#4b5563',
      legendFontSize: 12,
    }));

    return (
      <PieChart
        data={pieData}
        width={screenWidth - 48}
        height={220}
        chartConfig={chartConfig}
        accessor="quantity"
        backgroundColor="transparent"
        paddingLeft="15"
        absolute
      />
    );
  };

  // エクスポートボタン
  const renderExportButton = (type: ExportType, label: string) => (
    <TouchableOpacity
      key={type}
      onPress={() => exportMutation.mutate(type)}
      disabled={exportMutation.isPending}
      className={`mr-2 mb-2 rounded-lg bg-white px-4 py-3 shadow-sm ${
        exportMutation.isPending ? 'opacity-50' : ''
      }`}
    >
      <View className="flex-row items-center">
        <Ionicons name="download-outline" size={20} color="#16a34a" />
        <Text className="ml-2 font-medium text-gray-800">{label}</Text>
      </View>
    </TouchableOpacity>
  );

  // ---------------------------------------------------------------------------
  // Main Render - メインレンダリング
  // ---------------------------------------------------------------------------

  return (
    <ScrollView
      className="flex-1 bg-gray-50"
      refreshControl={
        <RefreshControl refreshing={isLoading} onRefresh={onRefresh} />
      }
    >
      {/* ヘッダー */}
      <View className="bg-primary-600 px-6 pb-6 pt-4">
        <Text className="text-2xl font-bold text-white">分析</Text>
        <Text className="mt-1 text-white/80">収穫データと統計を確認できます</Text>
      </View>

      {/* サマリーカード */}
      <View className="-mt-4 px-4">
        <View className="rounded-xl bg-white p-4 shadow-sm">
          <Text className="mb-3 text-lg font-bold text-gray-800">
            収穫サマリー
          </Text>

          {summaryLoading ? (
            <ActivityIndicator color="#16a34a" />
          ) : summaryData?.summaries && summaryData.summaries.length > 0 ? (
            <>
              {/* 総収穫量 */}
              <View className="mb-4 flex-row items-center justify-between">
                <View>
                  <Text className="text-sm text-gray-500">総収穫量</Text>
                  <Text className="text-2xl font-bold text-primary-600">
                    {summaryData.summaries.reduce(
                      (sum, s) => sum + s.total_quantity,
                      0
                    )}{' '}
                    kg
                  </Text>
                </View>
                <View>
                  <Text className="text-sm text-gray-500">収穫回数</Text>
                  <Text className="text-2xl font-bold text-gray-800">
                    {summaryData.summaries.reduce(
                      (sum, s) => sum + s.harvest_count,
                      0
                    )}{' '}
                    回
                  </Text>
                </View>
              </View>

              {/* 作物別円グラフ */}
              {renderPieChart(summaryData.summaries)}
            </>
          ) : (
            <View className="items-center justify-center py-8">
              <Ionicons name="analytics-outline" size={48} color="#9ca3af" />
              <Text className="mt-2 text-gray-500">
                収穫データがまだありません
              </Text>
            </View>
          )}
        </View>
      </View>

      {/* グラフセクション */}
      <View className="mt-6 px-4">
        <Text className="mb-3 text-lg font-bold text-gray-800">グラフ</Text>

        {/* グラフタイプ選択 */}
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          className="mb-4"
        >
          {renderChartTypeButton('monthly_harvest', '月別収穫量', 'bar-chart-outline')}
          {renderChartTypeButton('crop_comparison', '作物比較', 'pie-chart-outline')}
          {renderChartTypeButton('plot_productivity', '区画生産性', 'stats-chart-outline')}
        </ScrollView>

        {/* グラフ表示 */}
        <View className="rounded-xl bg-white p-4 shadow-sm">
          <Text className="mb-2 font-medium text-gray-800">
            {selectedChart === 'monthly_harvest' && '月別収穫量'}
            {selectedChart === 'crop_comparison' && '作物別収穫量比較'}
            {selectedChart === 'plot_productivity' && '区画別生産性'}
          </Text>

          {chartLoading ? (
            <View className="items-center justify-center py-8">
              <ActivityIndicator color="#16a34a" />
            </View>
          ) : chartData ? (
            renderBarChart(chartData)
          ) : (
            <View className="items-center justify-center py-8">
              <Text className="text-gray-500">グラフデータがありません</Text>
            </View>
          )}
        </View>
      </View>

      {/* CSVエクスポートセクション */}
      <View className="mt-6 px-4 pb-6">
        <Text className="mb-3 text-lg font-bold text-gray-800">
          データエクスポート
        </Text>
        <Text className="mb-3 text-sm text-gray-500">
          CSVファイルとしてデータをエクスポートできます
        </Text>

        <View className="flex-row flex-wrap">
          {renderExportButton('crops', '作物データ')}
          {renderExportButton('harvests', '収穫データ')}
          {renderExportButton('tasks', 'タスクデータ')}
          {renderExportButton('all', '全データ')}
        </View>

        {exportMutation.isPending && (
          <View className="mt-3 flex-row items-center">
            <ActivityIndicator size="small" color="#16a34a" />
            <Text className="ml-2 text-sm text-gray-500">
              エクスポート中...
            </Text>
          </View>
        )}
      </View>
    </ScrollView>
  );
}
