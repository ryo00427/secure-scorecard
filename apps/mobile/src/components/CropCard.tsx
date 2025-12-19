// =============================================================================
// CropCard - 作物カードコンポーネント
// =============================================================================
// ダッシュボードに表示される作物のカードです。
// 画像、名前、経過日数、進捗バー、収穫予定を表示します。

import React from 'react';
import { View, Text, TouchableOpacity, Image } from 'react-native';
import ProgressBar from './ProgressBar';

// 作物の状態に応じた色
const STATUS_COLORS = {
  planning: '#3b82f6', // 青
  growing: '#22c55e', // 緑
  harvested: '#f59e0b', // オレンジ
} as const;

// デフォルトの作物画像
const DEFAULT_CROP_IMAGES: Record<string, string> = {
  'ミニトマト': 'https://images.unsplash.com/photo-1592841200221-a6898f307baa?w=200',
  'トマト': 'https://images.unsplash.com/photo-1592841200221-a6898f307baa?w=200',
  'バジル': 'https://images.unsplash.com/photo-1618375569909-3c8616cf7733?w=200',
  'キュウリ': 'https://images.unsplash.com/photo-1449300079323-02e209d9d3a6?w=200',
  'ナス': 'https://images.unsplash.com/photo-1615484477778-ca3b77940c25?w=200',
  'ピーマン': 'https://images.unsplash.com/photo-1563565375-f3fdfdbefa83?w=200',
};

interface CropCardProps {
  // 作物ID
  id: number;
  // 作物名
  name: string;
  // 品種
  variety?: string;
  // 植え付け日
  plantedDate: string;
  // 収穫予定日
  expectedHarvestDate: string;
  // 状態
  status: 'planning' | 'growing' | 'harvested';
  // 画像URL（オプション）
  imageUrl?: string;
  // タップ時のコールバック
  onPress?: () => void;
}

export default function CropCard({
  name,
  plantedDate,
  expectedHarvestDate,
  status,
  imageUrl,
  onPress,
}: CropCardProps) {
  // 経過日数を計算
  const planted = new Date(plantedDate);
  const now = new Date();
  const daysSincePlanting = Math.floor(
    (now.getTime() - planted.getTime()) / (1000 * 60 * 60 * 24)
  );

  // 収穫までの日数を計算
  const harvest = new Date(expectedHarvestDate);
  const daysUntilHarvest = Math.floor(
    (harvest.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
  );

  // 進捗率を計算（植え付けから収穫予定までの割合）
  const totalDays = Math.floor(
    (harvest.getTime() - planted.getTime()) / (1000 * 60 * 60 * 24)
  );
  const progress = totalDays > 0 ? (daysSincePlanting / totalDays) * 100 : 0;

  // 収穫までのテキストを生成
  const getHarvestText = () => {
    if (status === 'harvested') {
      return '収穫済み';
    }
    if (daysUntilHarvest <= 0) {
      return 'いつでも収穫できます';
    }
    if (daysUntilHarvest <= 7) {
      return `収穫まであと${daysUntilHarvest}日`;
    }
    const weeks = Math.ceil(daysUntilHarvest / 7);
    return `収穫まであと約${weeks}週間`;
  };

  // 画像URLを取得（指定がなければデフォルト画像）
  const getImageUrl = () => {
    if (imageUrl) return imageUrl;
    return DEFAULT_CROP_IMAGES[name] || 'https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=200';
  };

  return (
    <TouchableOpacity
      onPress={onPress}
      className="mb-3 rounded-xl bg-white p-4 shadow-sm"
      activeOpacity={0.7}
    >
      <View className="flex-row items-center">
        {/* 作物画像 */}
        <View className="mr-4 h-20 w-20 overflow-hidden rounded-lg bg-gray-100">
          <Image
            source={{ uri: getImageUrl() }}
            className="h-full w-full"
            resizeMode="cover"
          />
        </View>

        {/* 作物情報 */}
        <View className="flex-1">
          {/* 名前とステータスインジケーター */}
          <View className="flex-row items-center justify-between">
            <Text className="text-lg font-semibold text-gray-800">{name}</Text>
            <View
              className="h-3 w-3 rounded-full"
              style={{ backgroundColor: STATUS_COLORS[status] }}
            />
          </View>

          {/* 経過日数 */}
          <Text className="mt-1 text-sm text-gray-500">
            種まきから{daysSincePlanting}日
          </Text>

          {/* 進捗バー */}
          <View className="mt-2">
            <ProgressBar
              progress={progress}
              color={STATUS_COLORS[status]}
              height={6}
            />
          </View>

          {/* 収穫予定 */}
          <Text className="mt-2 text-xs text-gray-400">{getHarvestText()}</Text>
        </View>
      </View>
    </TouchableOpacity>
  );
}
