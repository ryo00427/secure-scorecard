// =============================================================================
// WorkLogScreen - 作業ログ画面
// =============================================================================
// デザインファイル: design/stitch_ (2)/screen.png
// 作業記録を追加するフォームを提供します。

import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  ScrollView,
  SafeAreaView,
  StatusBar,
  Platform,
  KeyboardAvoidingView,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useNavigation, useRoute } from '@react-navigation/native';
import type { RouteProp } from '@react-navigation/native';
import { useQuery } from '@tanstack/react-query';
import { cropsApi } from '../../services/api';

// 作業タイプ
type WorkType = 'watering' | 'fertilizing' | 'pest_control' | 'harvesting' | 'other';

// ナビゲーションの型定義
type RootStackParamList = {
  WorkLog: { cropId?: number };
};

type RouteType = RouteProp<RootStackParamList, 'WorkLog'>;

// 作業タイプの定義
const WORK_TYPES: { key: WorkType; label: string; icon: string }[] = [
  { key: 'watering', label: '水やり', icon: 'water' },
  { key: 'fertilizing', label: '肥料', icon: 'flask' },
  { key: 'pest_control', label: '害虫駆除', icon: 'bug' },
  { key: 'harvesting', label: '収穫', icon: 'leaf' },
  { key: 'other', label: 'その他', icon: 'ellipsis-horizontal' },
];

export default function WorkLogScreen() {
  const navigation = useNavigation();
  const route = useRoute<RouteType>();
  const cropId = route.params?.cropId;

  // フォーム状態
  const [dateStr, setDateStr] = useState(
    new Date().toISOString().split('T')[0] // YYYY-MM-DD形式
  );
  const [selectedCropId, setSelectedCropId] = useState<number | null>(cropId || null);
  const [workType, setWorkType] = useState<WorkType>('watering');
  const [memo, setMemo] = useState('');
  const [showCropPicker, setShowCropPicker] = useState(false);

  // 作物一覧を取得
  const { data: cropsData } = useQuery({
    queryKey: ['crops'],
    queryFn: () => cropsApi.getAll(),
  });

  // APIは配列を直接返すので、cropsData自体が配列
  const crops = cropsData || [];
  const selectedCrop = crops.find((c) => c.id === selectedCropId);

  // 保存
  // 注: 作業ログAPIはバックエンドで実装予定
  const handleSave = () => {
    console.log('作業ログ保存:', {
      date: dateStr,
      cropId: selectedCropId,
      workType,
      memo,
    });
    navigation.goBack();
  };

  // 閉じる
  const handleClose = () => {
    navigation.goBack();
  };

  return (
    <SafeAreaView className="flex-1 bg-gray-100">
      <StatusBar barStyle="dark-content" />

      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        className="flex-1"
      >
        {/* ヘッダー */}
        <View className="flex-row items-center justify-between bg-gray-100 px-4 py-3">
          <TouchableOpacity onPress={handleClose} className="p-2">
            <Ionicons name="close" size={24} color="#6b7280" />
          </TouchableOpacity>
          <Text className="text-lg font-bold text-gray-600">作業ログ</Text>
          <TouchableOpacity
            onPress={handleSave}
            className="rounded-lg bg-emerald-500 px-4 py-2"
          >
            <Text className="font-bold text-white">保存</Text>
          </TouchableOpacity>
        </View>

        <ScrollView
          className="flex-1"
          showsVerticalScrollIndicator={false}
          keyboardShouldPersistTaps="handled"
        >
          {/* 日付と対象植物 */}
          <View className="mx-4 mt-4 overflow-hidden rounded-xl bg-gray-200">
            {/* 日付 */}
            <View className="flex-row items-center border-b border-gray-300 px-4 py-4">
              <View className="mr-3 h-10 w-10 items-center justify-center rounded-lg bg-emerald-500">
                <Ionicons name="calendar" size={20} color="white" />
              </View>
              <Text className="flex-1 text-gray-600">日付</Text>
              <TextInput
                value={dateStr}
                onChangeText={setDateStr}
                placeholder="2024-12-15"
                placeholderTextColor="#9ca3af"
                className="w-32 text-right text-gray-800"
                keyboardType="numbers-and-punctuation"
              />
            </View>

            {/* 対象植物 */}
            <TouchableOpacity
              onPress={() => setShowCropPicker(!showCropPicker)}
              className="flex-row items-center px-4 py-4"
            >
              <View className="mr-3 h-10 w-10 items-center justify-center rounded-lg bg-emerald-500">
                <Ionicons name="leaf" size={20} color="white" />
              </View>
              <Text className="flex-1 text-gray-600">対象の植物</Text>
              <Text className="text-gray-800">
                {selectedCrop ? selectedCrop.name : '選択してください'}
              </Text>
              <Ionicons name="chevron-forward" size={20} color="#9ca3af" />
            </TouchableOpacity>
          </View>

          {/* 植物選択ドロップダウン */}
          {showCropPicker && (
            <View className="mx-4 mt-2 rounded-xl bg-white shadow-sm">
              {crops.length > 0 ? (
                crops.map((crop) => (
                  <TouchableOpacity
                    key={crop.id}
                    onPress={() => {
                      setSelectedCropId(crop.id);
                      setShowCropPicker(false);
                    }}
                    className="flex-row items-center border-b border-gray-100 px-4 py-3"
                  >
                    <Text
                      className={`flex-1 ${
                        selectedCropId === crop.id
                          ? 'font-medium text-emerald-600'
                          : 'text-gray-800'
                      }`}
                    >
                      {crop.name}
                    </Text>
                    {selectedCropId === crop.id && (
                      <Ionicons name="checkmark" size={20} color="#22c55e" />
                    )}
                  </TouchableOpacity>
                ))
              ) : (
                <View className="px-4 py-3">
                  <Text className="text-gray-500">作物がありません</Text>
                </View>
              )}
            </View>
          )}

          {/* 作業の種類 */}
          <View className="px-4 py-4">
            <Text className="mb-3 font-medium text-gray-600">作業の種類</Text>
            <View className="flex-row flex-wrap">
              {WORK_TYPES.map((type) => (
                <TouchableOpacity
                  key={type.key}
                  onPress={() => setWorkType(type.key)}
                  className={`mb-2 mr-2 rounded-lg px-4 py-2 ${
                    workType === type.key ? 'bg-emerald-500' : 'bg-gray-200'
                  }`}
                >
                  <Text
                    className={`font-medium ${
                      workType === type.key ? 'text-white' : 'text-gray-500'
                    }`}
                  >
                    {type.label}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>

          {/* メモ */}
          <View className="px-4 py-2">
            <Text className="mb-3 font-medium text-gray-600">メモ</Text>
            <TextInput
              value={memo}
              onChangeText={setMemo}
              placeholder="作業の詳細や気づいたこと..."
              placeholderTextColor="#9ca3af"
              multiline
              numberOfLines={4}
              textAlignVertical="top"
              className="h-28 rounded-xl bg-white px-4 py-3 text-base text-gray-800"
            />
          </View>

          {/* 写真セクション */}
          <View className="px-4 py-4">
            <Text className="mb-3 font-medium text-gray-600">写真</Text>
            <TouchableOpacity className="h-24 w-24 items-center justify-center rounded-xl bg-white">
              <Ionicons name="camera" size={32} color="#9ca3af" />
              <Text className="mt-1 text-xs text-gray-400">追加</Text>
            </TouchableOpacity>
          </View>

          {/* 余白 */}
          <View className="h-24" />
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
