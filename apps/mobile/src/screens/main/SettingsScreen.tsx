// =============================================================================
// SettingsScreen - 設定画面
// =============================================================================
// ユーザー設定と通知設定を管理します。

import React from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  Switch,
  Alert,
} from 'react-native';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../../context/AuthContext';
import { notificationApi, authApi } from '../../services/api';

export default function SettingsScreen() {
  const { user, logout } = useAuth();
  const queryClient = useQueryClient();

  // 通知設定を取得
  const { data: notificationSettings } = useQuery({
    queryKey: ['notificationSettings'],
    queryFn: () => notificationApi.getSettings(),
  });

  // 通知設定更新ミューテーション
  const updateSettingsMutation = useMutation({
    mutationFn: notificationApi.updateSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notificationSettings'] });
    },
    onError: (error) => {
      Alert.alert('エラー', error instanceof Error ? error.message : '設定の更新に失敗しました');
    },
  });

  // ログアウト処理
  const handleLogout = () => {
    Alert.alert(
      'ログアウト',
      'ログアウトしますか？',
      [
        { text: 'キャンセル', style: 'cancel' },
        {
          text: 'ログアウト',
          style: 'destructive',
          onPress: async () => {
            try {
              await authApi.logout();
            } catch {
              // ログアウトAPIが失敗してもローカルのセッションはクリア
            }
            await logout();
          },
        },
      ]
    );
  };

  // 設定項目のトグル
  const toggleSetting = (key: string, value: boolean) => {
    updateSettingsMutation.mutate({ [key]: value });
  };

  return (
    <ScrollView className="flex-1 bg-gray-50">
      {/* プロフィール */}
      <View className="bg-white p-4">
        <View className="flex-row items-center">
          <View className="h-16 w-16 items-center justify-center rounded-full bg-primary-100">
            <Ionicons name="person" size={32} color="#16a34a" />
          </View>
          <View className="ml-4 flex-1">
            <Text className="text-xl font-bold text-gray-800">
              {user?.displayName || 'ユーザー'}
            </Text>
            <Text className="text-gray-500">{user?.email}</Text>
          </View>
        </View>
      </View>

      {/* 通知設定 */}
      <View className="mt-6">
        <Text className="px-4 pb-2 text-sm font-medium uppercase text-gray-500">
          通知設定
        </Text>
        <View className="bg-white">
          <View className="flex-row items-center justify-between border-b border-gray-100 px-4 py-3">
            <View className="flex-row items-center">
              <Ionicons name="notifications-outline" size={22} color="#374151" />
              <Text className="ml-3 text-base text-gray-800">プッシュ通知</Text>
            </View>
            <Switch
              value={notificationSettings?.push_enabled ?? true}
              onValueChange={(value) => toggleSetting('push_enabled', value)}
              trackColor={{ false: '#d1d5db', true: '#86efac' }}
              thumbColor={notificationSettings?.push_enabled ? '#16a34a' : '#f4f4f5'}
            />
          </View>
          <View className="flex-row items-center justify-between border-b border-gray-100 px-4 py-3">
            <View className="flex-row items-center">
              <Ionicons name="mail-outline" size={22} color="#374151" />
              <Text className="ml-3 text-base text-gray-800">メール通知</Text>
            </View>
            <Switch
              value={notificationSettings?.email_enabled ?? true}
              onValueChange={(value) => toggleSetting('email_enabled', value)}
              trackColor={{ false: '#d1d5db', true: '#86efac' }}
              thumbColor={notificationSettings?.email_enabled ? '#16a34a' : '#f4f4f5'}
            />
          </View>
          <View className="flex-row items-center justify-between border-b border-gray-100 px-4 py-3">
            <View className="flex-row items-center">
              <Ionicons name="checkbox-outline" size={22} color="#374151" />
              <Text className="ml-3 text-base text-gray-800">タスクリマインダー</Text>
            </View>
            <Switch
              value={notificationSettings?.task_reminders ?? true}
              onValueChange={(value) => toggleSetting('task_reminders', value)}
              trackColor={{ false: '#d1d5db', true: '#86efac' }}
              thumbColor={notificationSettings?.task_reminders ? '#16a34a' : '#f4f4f5'}
            />
          </View>
          <View className="flex-row items-center justify-between px-4 py-3">
            <View className="flex-row items-center">
              <Ionicons name="leaf-outline" size={22} color="#374151" />
              <Text className="ml-3 text-base text-gray-800">収穫リマインダー</Text>
            </View>
            <Switch
              value={notificationSettings?.harvest_reminders ?? true}
              onValueChange={(value) => toggleSetting('harvest_reminders', value)}
              trackColor={{ false: '#d1d5db', true: '#86efac' }}
              thumbColor={notificationSettings?.harvest_reminders ? '#16a34a' : '#f4f4f5'}
            />
          </View>
        </View>
      </View>

      {/* アプリ情報 */}
      <View className="mt-6">
        <Text className="px-4 pb-2 text-sm font-medium uppercase text-gray-500">
          アプリ情報
        </Text>
        <View className="bg-white">
          <View className="flex-row items-center justify-between border-b border-gray-100 px-4 py-3">
            <View className="flex-row items-center">
              <Ionicons name="information-circle-outline" size={22} color="#374151" />
              <Text className="ml-3 text-base text-gray-800">バージョン</Text>
            </View>
            <Text className="text-gray-500">0.1.0</Text>
          </View>
          <TouchableOpacity
            className="flex-row items-center border-b border-gray-100 px-4 py-3"
            onPress={() => Alert.alert('利用規約', '利用規約ページは実装中です')}
          >
            <Ionicons name="document-text-outline" size={22} color="#374151" />
            <Text className="ml-3 flex-1 text-base text-gray-800">利用規約</Text>
            <Ionicons name="chevron-forward" size={20} color="#9ca3af" />
          </TouchableOpacity>
          <TouchableOpacity
            className="flex-row items-center px-4 py-3"
            onPress={() => Alert.alert('プライバシーポリシー', 'プライバシーポリシーページは実装中です')}
          >
            <Ionicons name="shield-checkmark-outline" size={22} color="#374151" />
            <Text className="ml-3 flex-1 text-base text-gray-800">プライバシーポリシー</Text>
            <Ionicons name="chevron-forward" size={20} color="#9ca3af" />
          </TouchableOpacity>
        </View>
      </View>

      {/* ログアウト */}
      <View className="mt-6 px-4 pb-12">
        <TouchableOpacity
          className="items-center rounded-lg bg-red-500 py-4"
          onPress={handleLogout}
        >
          <Text className="text-lg font-semibold text-white">ログアウト</Text>
        </TouchableOpacity>
      </View>
    </ScrollView>
  );
}
