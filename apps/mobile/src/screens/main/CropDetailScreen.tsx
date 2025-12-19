// =============================================================================
// CropDetailScreen - 作物詳細画面
// =============================================================================
// デザインファイル: design/stitch_ (3)/screen.png
// 作物の詳細情報、成長曲線、作業履歴を表示します。

import React, { useState, type ComponentProps } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  Image,
  SafeAreaView,
  StatusBar,
  Dimensions,
  RefreshControl,
} from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { Ionicons } from '@expo/vector-icons';
import { useNavigation, useRoute } from '@react-navigation/native';
import type { RouteProp } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { cropsApi } from '../../services/api';

// 画面幅
const { width: SCREEN_WIDTH } = Dimensions.get('window');

// タブの種類
type TabType = 'overview' | 'log' | 'photos';

// ナビゲーションの型定義
type RootStackParamList = {
  CropDetail: { cropId: number };
  EditCrop: { cropId: number };
  WorkLog: { cropId: number };
};

type NavigationProp = NativeStackNavigationProp<RootStackParamList>;
type RouteType = RouteProp<RootStackParamList, 'CropDetail'>;

// デフォルトの作物画像
const DEFAULT_CROP_IMAGES: Record<string, string> = {
  'ミニトマト': 'https://images.unsplash.com/photo-1592841200221-a6898f307baa?w=600',
  'トマト': 'https://images.unsplash.com/photo-1592841200221-a6898f307baa?w=600',
  'バジル': 'https://images.unsplash.com/photo-1618375569909-3c8616cf7733?w=600',
  'キュウリ': 'https://images.unsplash.com/photo-1449300079323-02e209d9d3a6?w=600',
  'ナス': 'https://images.unsplash.com/photo-1615484477778-ca3b77940c25?w=600',
  'ピーマン': 'https://images.unsplash.com/photo-1563565375-f3fdfdbefa83?w=600',
};

// アクティビティアイコン
type IconName = ComponentProps<typeof Ionicons>['name'];
const ACTIVITY_ICONS: Record<string, { icon: IconName; color: string }> = {
  watering: { icon: 'water', color: '#3b82f6' },
  fertilizing: { icon: 'flask', color: '#f59e0b' },
  photo: { icon: 'camera', color: '#ec4899' },
  harvesting: { icon: 'leaf', color: '#22c55e' },
  pruning: { icon: 'cut', color: '#8b5cf6' },
};

// モック: 作業履歴データ
const mockActivities = [
  { id: 1, type: 'watering', label: '水やり', detail: '500ml', date: '昨日' },
  { id: 2, type: 'fertilizing', label: '施肥', detail: '液体肥料', date: '3日前' },
  { id: 3, type: 'photo', label: '写真追加', detail: '', date: '5日前' },
];

export default function CropDetailScreen() {
  const navigation = useNavigation<NavigationProp>();
  const route = useRoute<RouteType>();
  const { cropId } = route.params;
  const [activeTab, setActiveTab] = useState<TabType>('overview');

  // 作物詳細を取得
  const { data: cropData, isLoading, refetch } = useQuery({
    queryKey: ['crop', cropId],
    queryFn: () => cropsApi.getById(cropId),
  });

  // APIはオブジェクトを直接返すので、cropData自体がCrop
  const crop = cropData;

  // 戻る
  const handleBack = () => {
    navigation.goBack();
  };

  // 編集
  const handleEdit = () => {
    navigation.navigate('EditCrop', { cropId });
  };

  // 作業ログ追加
  const handleAddLog = () => {
    navigation.navigate('WorkLog', { cropId });
  };

  // 経過日数と進捗を計算
  const calculateProgress = () => {
    if (!crop) return { daysSincePlanting: 0, progress: 0 };

    const planted = new Date(crop.planted_date);
    const harvest = new Date(crop.expected_harvest_date);
    const now = new Date();

    const daysSincePlanting = Math.floor(
      (now.getTime() - planted.getTime()) / (1000 * 60 * 60 * 24)
    );
    const totalDays = Math.floor(
      (harvest.getTime() - planted.getTime()) / (1000 * 60 * 60 * 24)
    );
    const progress = totalDays > 0 ? (daysSincePlanting / totalDays) * 100 : 0;

    return { daysSincePlanting, progress };
  };

  const { daysSincePlanting, progress: _progress } = calculateProgress();

  // 画像URLを取得
  const getImageUrl = () => {
    if (!crop) return 'https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=600';
    return DEFAULT_CROP_IMAGES[crop.name] || 'https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=600';
  };

  if (isLoading || !crop) {
    return (
      <SafeAreaView className="flex-1 items-center justify-center bg-gray-50">
        <Text className="text-gray-500">読み込み中...</Text>
      </SafeAreaView>
    );
  }

  // タブコンテンツ
  const renderTabContent = () => {
    switch (activeTab) {
      case 'overview':
        return (
          <View>
            {/* 詳細情報グリッド */}
            <View className="mx-4 mt-4 rounded-xl bg-white p-4">
              <View className="flex-row">
                <View className="flex-1 border-r border-gray-100 pr-4">
                  <Text className="text-xs text-gray-500">品種</Text>
                  <Text className="mt-1 font-medium text-gray-800">{crop.variety || crop.name}</Text>
                </View>
                <View className="flex-1 pl-4">
                  <Text className="text-xs text-gray-500">植え付け日</Text>
                  <Text className="mt-1 font-medium text-gray-800">
                    {new Date(crop.planted_date).toLocaleDateString('ja-JP')}
                  </Text>
                </View>
              </View>
              <View className="mt-4 flex-row border-t border-gray-100 pt-4">
                <View className="flex-1 border-r border-gray-100 pr-4">
                  <Text className="text-xs text-gray-500">高さ</Text>
                  <Text className="mt-1 font-medium text-gray-800">30cm</Text>
                </View>
                <View className="flex-1 pl-4">
                  <Text className="text-xs text-gray-500">植え付けからの日数</Text>
                  <Text className="mt-1 font-medium text-gray-800">{daysSincePlanting}日</Text>
                </View>
              </View>
            </View>

            {/* 最近のアクティビティ */}
            <View className="mx-4 mt-4 rounded-xl bg-white p-4">
              <View className="flex-row items-center justify-between">
                <Text className="font-semibold text-gray-800">最近のアクティビティ</Text>
                <TouchableOpacity>
                  <Text className="text-sm font-medium text-emerald-600">すべて見る</Text>
                </TouchableOpacity>
              </View>
              <View className="mt-3">
                {mockActivities.map((activity) => {
                  const iconInfo = ACTIVITY_ICONS[activity.type] || { icon: 'ellipse' as IconName, color: '#6b7280' };
                  return (
                    <View key={activity.id} className="flex-row items-center py-3">
                      <View
                        className="mr-3 h-10 w-10 items-center justify-center rounded-full"
                        style={{ backgroundColor: `${iconInfo.color}20` }}
                      >
                        <Ionicons
                          name={iconInfo.icon}
                          size={20}
                          color={iconInfo.color}
                        />
                      </View>
                      <View className="flex-1">
                        <Text className="font-medium text-gray-800">{activity.label}</Text>
                        {activity.detail && (
                          <Text className="text-sm text-gray-500">{activity.detail}</Text>
                        )}
                      </View>
                      <Text className="text-sm text-gray-400">{activity.date}</Text>
                    </View>
                  );
                })}
              </View>
            </View>
          </View>
        );

      case 'log':
        return (
          <View className="items-center justify-center py-16">
            <Ionicons name="document-text-outline" size={64} color="#d1d5db" />
            <Text className="mt-4 text-gray-500">作業ログはまだありません</Text>
            <TouchableOpacity
              onPress={handleAddLog}
              className="mt-4 rounded-full bg-emerald-600 px-6 py-2"
            >
              <Text className="font-medium text-white">作業を記録する</Text>
            </TouchableOpacity>
          </View>
        );

      case 'photos':
        return (
          <View className="items-center justify-center py-16">
            <Ionicons name="images-outline" size={64} color="#d1d5db" />
            <Text className="mt-4 text-gray-500">写真はまだありません</Text>
            <TouchableOpacity className="mt-4 rounded-full bg-emerald-600 px-6 py-2">
              <Text className="font-medium text-white">写真を追加する</Text>
            </TouchableOpacity>
          </View>
        );

      default:
        return null;
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-gray-50">
      <StatusBar barStyle="light-content" />

      <ScrollView
        className="flex-1"
        refreshControl={
          <RefreshControl refreshing={isLoading} onRefresh={refetch} />
        }
        showsVerticalScrollIndicator={false}
      >
        {/* ヘッダー画像 */}
        <View className="relative">
          <Image
            source={{ uri: getImageUrl() }}
            style={{ width: SCREEN_WIDTH, height: 280 }}
            resizeMode="cover"
          />
          {/* グラデーションオーバーレイ */}
          <View className="absolute inset-0 bg-black/30" />

          {/* ナビゲーションボタン */}
          <View className="absolute left-0 right-0 top-0 flex-row items-center justify-between p-4">
            <TouchableOpacity
              onPress={handleBack}
              className="rounded-full bg-white/20 p-2"
            >
              <Ionicons name="chevron-back" size={24} color="white" />
            </TouchableOpacity>
            <TouchableOpacity
              onPress={handleEdit}
              className="rounded-full bg-white/20 px-4 py-2"
            >
              <Text className="font-medium text-white">編集</Text>
            </TouchableOpacity>
          </View>

          {/* 作物名 */}
          <View className="absolute bottom-4 left-4">
            <Text className="text-2xl font-bold text-white">{crop.name}</Text>
          </View>
        </View>

        {/* タブ */}
        <View className="flex-row border-b border-gray-200 bg-white">
          {[
            { key: 'overview' as TabType, label: '概要' },
            { key: 'log' as TabType, label: 'ログ' },
            { key: 'photos' as TabType, label: '写真' },
          ].map((tab) => (
            <TouchableOpacity
              key={tab.key}
              onPress={() => setActiveTab(tab.key)}
              className={`flex-1 items-center py-3 ${
                activeTab === tab.key ? 'border-b-2 border-emerald-600' : ''
              }`}
            >
              <Text
                className={`font-medium ${
                  activeTab === tab.key ? 'text-emerald-600' : 'text-gray-500'
                }`}
              >
                {tab.label}
              </Text>
            </TouchableOpacity>
          ))}
        </View>

        {/* タブコンテンツ */}
        <View className="pb-24">{renderTabContent()}</View>
      </ScrollView>

      {/* FAB（作業ログ追加ボタン） */}
      <TouchableOpacity
        onPress={handleAddLog}
        className="absolute bottom-6 right-6 h-14 w-14 items-center justify-center rounded-full bg-emerald-600 shadow-lg"
        activeOpacity={0.8}
      >
        <Ionicons name="add" size={28} color="white" />
      </TouchableOpacity>
    </SafeAreaView>
  );
}
