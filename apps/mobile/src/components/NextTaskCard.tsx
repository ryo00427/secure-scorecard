// =============================================================================
// NextTaskCard - 次の作業カードコンポーネント
// =============================================================================
// ダッシュボードに表示される「次の作業」の注目カードです。
// 緑背景に作物画像と完了ボタンを表示します。

import React from 'react';
import { View, Text, TouchableOpacity, Image } from 'react-native';
import { Ionicons } from '@expo/vector-icons';

// 作業タイプに応じたデフォルト画像
const TASK_IMAGES: Record<string, string> = {
  watering: 'https://images.unsplash.com/photo-1592841200221-a6898f307baa?w=300',
  fertilizing: 'https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=300',
  harvesting: 'https://images.unsplash.com/photo-1592841200221-a6898f307baa?w=300',
  default: 'https://images.unsplash.com/photo-1592841200221-a6898f307baa?w=300',
};

interface NextTaskCardProps {
  // タスクのタイトル
  title: string;
  // タスクの説明
  description?: string;
  // 対象の作物名
  cropName?: string;
  // 作物の画像URL
  cropImageUrl?: string;
  // 完了ボタン押下時のコールバック
  onComplete?: () => void;
  // カードタップ時のコールバック
  onPress?: () => void;
}

export default function NextTaskCard({
  title,
  description,
  cropImageUrl,
  onComplete,
  onPress,
}: NextTaskCardProps) {
  return (
    <TouchableOpacity
      onPress={onPress}
      activeOpacity={0.9}
      className="overflow-hidden rounded-xl"
    >
      <View className="flex-row bg-emerald-700">
        {/* 左側: タスク情報 */}
        <View className="flex-1 justify-center p-5">
          <Text className="text-lg font-bold text-white">{title}</Text>
          {description && (
            <Text className="mt-1 text-sm text-white/80">{description}</Text>
          )}

          {/* 完了ボタン */}
          <TouchableOpacity
            onPress={onComplete}
            className="mt-4 flex-row items-center self-start rounded-full bg-emerald-600 px-4 py-2"
            activeOpacity={0.7}
          >
            <Text className="mr-2 font-medium text-white">完了</Text>
            <Ionicons name="checkmark" size={18} color="white" />
          </TouchableOpacity>
        </View>

        {/* 右側: 作物画像 */}
        <View className="w-32">
          <Image
            source={{ uri: cropImageUrl || TASK_IMAGES.default }}
            className="h-full w-full"
            resizeMode="cover"
          />
        </View>
      </View>
    </TouchableOpacity>
  );
}
