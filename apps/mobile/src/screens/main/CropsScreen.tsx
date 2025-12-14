// =============================================================================
// CropsScreen - 作物一覧画面（マイプラント）
// =============================================================================
// デザインファイル: design/stitch_/screen.png（ダッシュボードと同様のカード形式）
// 作物の一覧表示と管理を提供します。

import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
  SafeAreaView,
  StatusBar,
} from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { Ionicons } from '@expo/vector-icons';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { cropsApi } from '../../services/api';
import { CropCard } from '../../components';

// フィルタータイプ
type FilterType = 'all' | 'growing' | 'harvested';

// ナビゲーションの型定義
type RootStackParamList = {
  CropDetail: { cropId: number };
  AddCrop: undefined;
};

type NavigationProp = NativeStackNavigationProp<RootStackParamList>;

export default function CropsScreen() {
  const navigation = useNavigation<NavigationProp>();
  const [filter, setFilter] = useState<FilterType>('all');

  // 作物一覧を取得
  const { data: cropsData, isLoading, refetch } = useQuery({
    queryKey: ['crops'],
    queryFn: () => cropsApi.getAll(),
  });

  // APIは配列を直接返すので、cropsData自体が配列
  const allCrops = cropsData || [];
  const crops =
    filter === 'all'
      ? allCrops
      : allCrops.filter((c) => c.status === filter);

  // 作物詳細画面へ遷移
  const handleCropPress = (cropId: number) => {
    navigation.navigate('CropDetail', { cropId });
  };

  // 作物追加画面へ遷移
  const handleAddCrop = () => {
    navigation.navigate('AddCrop');
  };

  // フィルターオプション
  const filterOptions = [
    { key: 'all' as FilterType, label: 'すべて' },
    { key: 'growing' as FilterType, label: '栽培中' },
    { key: 'harvested' as FilterType, label: '収穫済' },
  ];

  return (
    <SafeAreaView className="flex-1 bg-gray-50">
      <StatusBar barStyle="dark-content" />

      {/* ヘッダー */}
      <View className="flex-row items-center justify-between bg-white px-4 py-3">
        <Text className="text-xl font-bold text-gray-800">マイプラント</Text>
        <TouchableOpacity onPress={handleAddCrop} className="p-2">
          <Ionicons name="add-circle-outline" size={28} color="#22c55e" />
        </TouchableOpacity>
      </View>

      {/* フィルタータブ */}
      <View className="flex-row border-b border-gray-200 bg-white px-4">
        {filterOptions.map((item) => (
          <TouchableOpacity
            key={item.key}
            className={`mr-6 py-3 ${
              filter === item.key ? 'border-b-2 border-emerald-600' : ''
            }`}
            onPress={() => setFilter(item.key)}
          >
            <Text
              className={`text-base ${
                filter === item.key
                  ? 'font-semibold text-emerald-600'
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
        showsVerticalScrollIndicator={false}
      >
        {crops.length > 0 ? (
          <View className="pb-24">
            {crops.map((crop) => (
              <CropCard
                key={crop.id}
                id={crop.id}
                name={crop.name}
                variety={crop.variety}
                plantedDate={crop.planted_date}
                expectedHarvestDate={crop.expected_harvest_date}
                status={crop.status}
                onPress={() => handleCropPress(crop.id)}
              />
            ))}
          </View>
        ) : (
          <View className="items-center justify-center py-16">
            <Ionicons name="leaf-outline" size={64} color="#d1d5db" />
            <Text className="mt-4 text-lg text-gray-500">
              {filter === 'all'
                ? '作物がありません'
                : filter === 'growing'
                ? '栽培中の作物がありません'
                : '収穫済みの作物がありません'}
            </Text>
            <TouchableOpacity
              onPress={handleAddCrop}
              className="mt-6 rounded-full bg-emerald-600 px-8 py-3"
            >
              <Text className="font-medium text-white">作物を追加する</Text>
            </TouchableOpacity>
          </View>
        )}
      </ScrollView>

      {/* FAB（作物追加ボタン） */}
      <TouchableOpacity
        onPress={handleAddCrop}
        className="absolute bottom-6 right-6 h-14 w-14 items-center justify-center rounded-full bg-gray-800 shadow-lg"
        activeOpacity={0.8}
      >
        <Ionicons name="add" size={28} color="white" />
      </TouchableOpacity>
    </SafeAreaView>
  );
}
