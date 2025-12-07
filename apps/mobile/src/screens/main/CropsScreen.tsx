// =============================================================================
// CropsScreen - 作物一覧画面
// =============================================================================
// 作物の一覧表示と管理を提供します。

import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
  Alert,
} from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { Ionicons } from '@expo/vector-icons';
import { cropsApi } from '../../services/api';

// -----------------------------------------------------------------------------
// Types - 型定義
// -----------------------------------------------------------------------------

type FilterType = 'all' | 'growing' | 'harvested';

// -----------------------------------------------------------------------------
// Component - コンポーネント
// -----------------------------------------------------------------------------

export default function CropsScreen() {
  const [filter, setFilter] = useState<FilterType>('all');

  // 作物一覧を取得
  const { data: cropsData, isLoading, refetch } = useQuery({
    queryKey: ['crops'],
    queryFn: () => cropsApi.getAll(),
  });

  const allCrops = cropsData?.crops || [];
  const crops = filter === 'all'
    ? allCrops
    : allCrops.filter(c => c.status === filter);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'planning':
        return { bg: 'bg-blue-100', text: 'text-blue-700', label: '計画中' };
      case 'growing':
        return { bg: 'bg-green-100', text: 'text-green-700', label: '栽培中' };
      case 'harvested':
        return { bg: 'bg-orange-100', text: 'text-orange-700', label: '収穫済' };
      default:
        return { bg: 'bg-gray-100', text: 'text-gray-700', label: status };
    }
  };

  return (
    <View className="flex-1 bg-gray-50">
      {/* フィルタータブ */}
      <View className="flex-row border-b border-gray-200 bg-white px-4">
        {[
          { key: 'all', label: 'すべて' },
          { key: 'growing', label: '栽培中' },
          { key: 'harvested', label: '収穫済' },
        ].map((item) => (
          <TouchableOpacity
            key={item.key}
            className={`mr-4 py-3 ${
              filter === item.key ? 'border-b-2 border-primary-600' : ''
            }`}
            onPress={() => setFilter(item.key as FilterType)}
          >
            <Text
              className={`text-base ${
                filter === item.key
                  ? 'font-semibold text-primary-600'
                  : 'text-gray-500'
              }`}
            >
              {item.label}
            </Text>
          </TouchableOpacity>
        ))}
      </View>

      <ScrollView
        className="flex-1 px-4 pt-4"
        refreshControl={
          <RefreshControl refreshing={isLoading} onRefresh={refetch} />
        }
      >
        {crops.length > 0 ? (
          crops.map((crop) => {
            const badge = getStatusBadge(crop.status);
            return (
              <TouchableOpacity
                key={crop.id}
                className="mb-3 rounded-lg bg-white p-4 shadow-sm"
                onPress={() => Alert.alert('作物詳細', `${crop.name}の詳細画面は実装中です`)}
              >
                <View className="flex-row items-center">
                  {/* アイコン */}
                  <View className="mr-4 h-12 w-12 items-center justify-center rounded-full bg-green-100">
                    <Ionicons name="leaf" size={24} color="#16a34a" />
                  </View>

                  {/* 作物情報 */}
                  <View className="flex-1">
                    <Text className="text-lg font-medium text-gray-800">
                      {crop.name}
                    </Text>
                    <Text className="text-sm text-gray-500">{crop.variety}</Text>
                  </View>

                  {/* ステータスバッジ */}
                  <View className={`rounded-full px-3 py-1 ${badge.bg}`}>
                    <Text className={`text-xs font-medium ${badge.text}`}>
                      {badge.label}
                    </Text>
                  </View>
                </View>

                {/* 日付情報 */}
                <View className="mt-3 flex-row border-t border-gray-100 pt-3">
                  <View className="flex-1 flex-row items-center">
                    <Ionicons name="calendar-outline" size={14} color="#6b7280" />
                    <Text className="ml-1 text-xs text-gray-500">
                      植付: {new Date(crop.planted_date).toLocaleDateString('ja-JP')}
                    </Text>
                  </View>
                  <View className="flex-row items-center">
                    <Ionicons name="flag-outline" size={14} color="#6b7280" />
                    <Text className="ml-1 text-xs text-gray-500">
                      収穫予定: {new Date(crop.expected_harvest_date).toLocaleDateString('ja-JP')}
                    </Text>
                  </View>
                </View>
              </TouchableOpacity>
            );
          })
        ) : (
          <View className="items-center justify-center py-12">
            <Ionicons name="leaf-outline" size={48} color="#d1d5db" />
            <Text className="mt-4 text-gray-500">作物がありません</Text>
          </View>
        )}
      </ScrollView>

      {/* 追加ボタン */}
      <TouchableOpacity
        className="absolute bottom-6 right-6 h-14 w-14 items-center justify-center rounded-full bg-primary-600 shadow-lg"
        onPress={() => Alert.alert('作物登録', '作物登録画面は実装中です')}
      >
        <Ionicons name="add" size={28} color="white" />
      </TouchableOpacity>
    </View>
  );
}
