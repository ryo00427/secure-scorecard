// =============================================================================
// TasksScreen - タスク一覧画面
// =============================================================================
// タスクの一覧表示と管理を提供します。

import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
  Alert,
} from 'react-native';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Ionicons } from '@expo/vector-icons';
import { tasksApi } from '../../services/api';

// -----------------------------------------------------------------------------
// Types - 型定義
// -----------------------------------------------------------------------------

type FilterType = 'all' | 'today' | 'overdue';

// -----------------------------------------------------------------------------
// Component - コンポーネント
// -----------------------------------------------------------------------------

export default function TasksScreen() {
  const queryClient = useQueryClient();
  const [filter, setFilter] = useState<FilterType>('all');

  // タスク一覧を取得
  const { data: allTasks, isLoading, refetch } = useQuery({
    queryKey: ['tasks', filter],
    queryFn: () => {
      switch (filter) {
        case 'today':
          return tasksApi.getToday();
        case 'overdue':
          return tasksApi.getOverdue();
        default:
          return tasksApi.getAll();
      }
    },
  });

  // タスク完了ミューテーション
  const completeMutation = useMutation({
    mutationFn: (id: number) => tasksApi.complete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
    onError: (error) => {
      Alert.alert('エラー', error instanceof Error ? error.message : 'タスクの完了に失敗しました');
    },
  });

  const handleComplete = (id: number, title: string) => {
    Alert.alert(
      'タスク完了',
      `「${title}」を完了しますか？`,
      [
        { text: 'キャンセル', style: 'cancel' },
        {
          text: '完了',
          onPress: () => completeMutation.mutate(id),
        },
      ]
    );
  };

  const tasks = allTasks?.tasks || [];

  return (
    <View className="flex-1 bg-gray-50">
      {/* フィルタータブ */}
      <View className="flex-row border-b border-gray-200 bg-white px-4">
        {[
          { key: 'all', label: 'すべて' },
          { key: 'today', label: '今日' },
          { key: 'overdue', label: '期限切れ' },
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
        {tasks.length > 0 ? (
          tasks.map((task) => (
            <View
              key={task.id}
              className="mb-3 rounded-lg bg-white p-4 shadow-sm"
            >
              <View className="flex-row items-start">
                {/* 完了ボタン */}
                <TouchableOpacity
                  className={`mr-3 h-6 w-6 items-center justify-center rounded-full border-2 ${
                    task.status === 'completed'
                      ? 'border-green-500 bg-green-500'
                      : 'border-gray-300'
                  }`}
                  onPress={() => {
                    if (task.status !== 'completed') {
                      handleComplete(task.id, task.title);
                    }
                  }}
                  disabled={task.status === 'completed'}
                >
                  {task.status === 'completed' && (
                    <Ionicons name="checkmark" size={14} color="white" />
                  )}
                </TouchableOpacity>

                {/* タスク情報 */}
                <View className="flex-1">
                  <Text
                    className={`text-base ${
                      task.status === 'completed'
                        ? 'text-gray-400 line-through'
                        : 'text-gray-800'
                    }`}
                  >
                    {task.title}
                  </Text>
                  {task.description && (
                    <Text className="mt-1 text-sm text-gray-500">
                      {task.description}
                    </Text>
                  )}
                  <View className="mt-2 flex-row items-center">
                    {/* 優先度バッジ */}
                    <View
                      className={`mr-2 rounded-full px-2 py-0.5 ${
                        task.priority === 'high'
                          ? 'bg-red-100'
                          : task.priority === 'medium'
                          ? 'bg-yellow-100'
                          : 'bg-green-100'
                      }`}
                    >
                      <Text
                        className={`text-xs ${
                          task.priority === 'high'
                            ? 'text-red-700'
                            : task.priority === 'medium'
                            ? 'text-yellow-700'
                            : 'text-green-700'
                        }`}
                      >
                        {task.priority === 'high'
                          ? '高'
                          : task.priority === 'medium'
                          ? '中'
                          : '低'}
                      </Text>
                    </View>
                    {/* 期限 */}
                    <View className="flex-row items-center">
                      <Ionicons name="calendar-outline" size={14} color="#6b7280" />
                      <Text className="ml-1 text-xs text-gray-500">
                        {new Date(task.due_date).toLocaleDateString('ja-JP')}
                      </Text>
                    </View>
                  </View>
                </View>
              </View>
            </View>
          ))
        ) : (
          <View className="items-center justify-center py-12">
            <Ionicons name="checkbox-outline" size={48} color="#d1d5db" />
            <Text className="mt-4 text-gray-500">タスクがありません</Text>
          </View>
        )}
      </ScrollView>

      {/* 追加ボタン */}
      <TouchableOpacity
        className="absolute bottom-6 right-6 h-14 w-14 items-center justify-center rounded-full bg-primary-600 shadow-lg"
        onPress={() => Alert.alert('タスク作成', 'タスク作成画面は実装中です')}
      >
        <Ionicons name="add" size={28} color="white" />
      </TouchableOpacity>
    </View>
  );
}
