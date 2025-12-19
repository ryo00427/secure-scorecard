// =============================================================================
// AddCropScreen - 植物追加画面
// =============================================================================
// デザインファイル: design/stitch_ (1)/screen.png
// 新しい作物を登録するフォームを提供します。

import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  ScrollView,
  SafeAreaView,
  StatusBar,
  Image,
  Platform,
  KeyboardAvoidingView,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useNavigation } from '@react-navigation/native';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { cropsApi } from '../../services/api';

// デフォルトの植物画像
const DEFAULT_PLANT_IMAGE = 'https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=600';

export default function AddCropScreen() {
  const navigation = useNavigation();
  const queryClient = useQueryClient();

  // フォーム状態
  const [name, setName] = useState<string>('');
  const [plantedDateStr, setPlantedDateStr] = useState<string>(
    new Date().toISOString().split('T')[0] ?? '' // YYYY-MM-DD形式
  );
  const [location, setLocation] = useState('');
  const [memo, setMemo] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  // 作物作成ミューテーション
  const createCropMutation = useMutation({
    mutationFn: (data: {
      name: string;
      variety: string;
      planted_date: string;
      expected_harvest_date: string;
      status: 'planning' | 'growing' | 'harvested';
    }) => cropsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['crops'] });
      navigation.goBack();
    },
    onError: (error: Error) => {
      setErrors({ submit: error.message });
    },
  });

  // バリデーション
  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!name.trim()) {
      newErrors.name = '植物の名前を入力してください';
    }

    // 日付形式チェック（YYYY-MM-DD）
    const dateRegex = /^\d{4}-\d{2}-\d{2}$/;
    if (!plantedDateStr.trim()) {
      newErrors.date = '植え付け日を入力してください';
    } else if (!dateRegex.test(plantedDateStr)) {
      newErrors.date = '日付はYYYY-MM-DD形式で入力してください';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 保存
  const handleSave = () => {
    if (!validate()) return;

    // 植え付け日をパース
    const plantedDate = new Date(plantedDateStr + 'T00:00:00Z');

    // 収穫予定日を植え付け日から60日後に設定（仮）
    const harvestDate = new Date(plantedDate);
    harvestDate.setDate(harvestDate.getDate() + 60);

    createCropMutation.mutate({
      name: name.trim(),
      variety: location.trim() || name.trim(),
      planted_date: plantedDate.toISOString(),
      expected_harvest_date: harvestDate.toISOString(),
      status: 'growing',
    });
  };

  // 戻る
  const handleBack = () => {
    navigation.goBack();
  };

  return (
    <SafeAreaView className="flex-1 bg-white">
      <StatusBar barStyle="dark-content" />

      <KeyboardAvoidingView
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        className="flex-1"
      >
        {/* ヘッダー */}
        <View className="flex-row items-center justify-between border-b border-gray-100 px-4 py-3">
          <TouchableOpacity onPress={handleBack} className="p-2">
            <Ionicons name="arrow-back" size={24} color="#1f2937" />
          </TouchableOpacity>
          <Text className="text-lg font-bold text-gray-800">植物の追加</Text>
          <View className="w-10" />
        </View>

        <ScrollView
          className="flex-1"
          showsVerticalScrollIndicator={false}
          keyboardShouldPersistTaps="handled"
        >
          {/* 画像プレースホルダー */}
          <View className="relative mx-4 mt-4 overflow-hidden rounded-xl">
            <Image
              source={{ uri: DEFAULT_PLANT_IMAGE }}
              className="h-48 w-full"
              resizeMode="cover"
            />
            <View className="absolute inset-0 items-center justify-center bg-black/30">
              <View className="items-center">
                <Ionicons name="camera" size={32} color="white" />
                <Text className="mt-2 font-medium text-white">写真を変更</Text>
              </View>
            </View>
          </View>

          {/* フォーム */}
          <View className="px-4 py-6">
            {/* 植物の名前 */}
            <View className="mb-6">
              <Text className="mb-2 text-sm font-medium text-gray-700">
                植物の名前 <Text className="text-red-500">*</Text>
              </Text>
              <TextInput
                value={name}
                onChangeText={setName}
                placeholder="例：ミニトマト"
                placeholderTextColor="#9ca3af"
                className={`rounded-lg border bg-white px-4 py-3 text-base text-gray-800 ${
                  errors.name ? 'border-red-500' : 'border-gray-200'
                }`}
              />
              {errors.name && <Text className="mt-1 text-sm text-red-500">{errors.name}</Text>}
            </View>

            {/* 植え付け日 */}
            <View className="mb-6">
              <Text className="mb-2 text-sm font-medium text-gray-700">
                植え付け日 <Text className="text-red-500">*</Text>
              </Text>
              <View className="flex-row items-center rounded-lg border border-gray-200 bg-white">
                <TextInput
                  value={plantedDateStr}
                  onChangeText={setPlantedDateStr}
                  placeholder="2024-01-01"
                  placeholderTextColor="#9ca3af"
                  className="flex-1 px-4 py-3 text-base text-gray-800"
                  keyboardType="numbers-and-punctuation"
                />
                <View className="pr-4">
                  <Ionicons name="calendar-outline" size={20} color="#6b7280" />
                </View>
              </View>
              {errors.date && <Text className="mt-1 text-sm text-red-500">{errors.date}</Text>}
            </View>

            {/* 場所 */}
            <View className="mb-6">
              <Text className="mb-2 text-sm font-medium text-gray-700">場所（オプション）</Text>
              <TextInput
                value={location}
                onChangeText={setLocation}
                placeholder="例：南のベランダ"
                placeholderTextColor="#9ca3af"
                className="rounded-lg border border-gray-200 bg-white px-4 py-3 text-base text-gray-800"
              />
            </View>

            {/* メモ */}
            <View className="mb-6">
              <Text className="mb-2 text-sm font-medium text-gray-700">メモ（オプション）</Text>
              <TextInput
                value={memo}
                onChangeText={setMemo}
                placeholder="水やりの頻度、成長の記録など"
                placeholderTextColor="#9ca3af"
                multiline
                numberOfLines={4}
                textAlignVertical="top"
                className="h-28 rounded-lg border border-gray-200 bg-white px-4 py-3 text-base text-gray-800"
              />
            </View>

            {/* エラーメッセージ */}
            {errors.submit && (
              <View className="mb-4 rounded-lg bg-red-50 p-3">
                <Text className="text-center text-red-600">{errors.submit}</Text>
              </View>
            )}
          </View>
        </ScrollView>

        {/* 保存ボタン */}
        <View className="border-t border-gray-100 px-4 py-4">
          <TouchableOpacity
            onPress={handleSave}
            disabled={createCropMutation.isPending}
            className={`items-center rounded-full py-4 ${
              createCropMutation.isPending ? 'bg-emerald-300' : 'bg-emerald-500'
            }`}
          >
            <Text className="text-lg font-bold text-white">
              {createCropMutation.isPending ? '保存中...' : '保存'}
            </Text>
          </TouchableOpacity>
        </View>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
