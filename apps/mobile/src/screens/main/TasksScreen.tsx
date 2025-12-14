// =============================================================================
// TasksScreen - タスク一覧画面
// =============================================================================
// タスクの一覧表示と管理を提供します。
// タスクの作成、完了、フィルタリング機能を提供します。

import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  RefreshControl,
  Alert,
  Modal,
  TextInput,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Ionicons } from '@expo/vector-icons';
import { tasksApi } from '../../services/api';

type FilterType = 'all' | 'today' | 'overdue';
type PriorityType = 'low' | 'medium' | 'high';

// タスク作成フォームの初期値
const initialFormState = {
  title: '',
  description: '',
  due_date: '',
  priority: 'medium' as PriorityType,
};

export default function TasksScreen() {
  const queryClient = useQueryClient();
  const [filter, setFilter] = useState<FilterType>('all');
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [formData, setFormData] = useState(initialFormState);

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

  // タスク作成ミューテーション
  const createMutation = useMutation({
    mutationFn: (data: typeof initialFormState) => tasksApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      setIsModalVisible(false);
      setFormData(initialFormState);
      Alert.alert('成功', 'タスクを作成しました');
    },
    onError: (error: unknown) => {
      console.error('タスク作成エラー:', error);
      let errorMessage = 'タスクの作成に失敗しました';
      if (error instanceof Error) {
        errorMessage = error.message;
      } else if (typeof error === 'object' && error !== null) {
        // ApiError などのカスタムエラーオブジェクトを処理
        const errorObj = error as { message?: string; error?: string };
        errorMessage = errorObj.message || errorObj.error || JSON.stringify(error);
      }
      Alert.alert('エラー', errorMessage);
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

  // タスク作成ハンドラー
  const handleCreate = () => {
    // バリデーション
    if (!formData.title.trim()) {
      Alert.alert('エラー', 'タイトルを入力してください');
      return;
    }
    if (!formData.due_date.trim()) {
      Alert.alert('エラー', '期限を入力してください（例: 2024-12-31）');
      return;
    }
    // 日付形式チェック（YYYY-MM-DD）
    const dateRegex = /^\d{4}-\d{2}-\d{2}$/;
    if (!dateRegex.test(formData.due_date)) {
      Alert.alert('エラー', '期限は YYYY-MM-DD 形式で入力してください（例: 2024-12-31）');
      return;
    }
    // バックエンドはRFC3339形式（time.Time）を期待するため変換
    const dueDateISO = new Date(formData.due_date + 'T00:00:00Z').toISOString();
    createMutation.mutate({
      ...formData,
      due_date: dueDateISO,
    });
  };

  // モーダルを開く
  const openModal = () => {
    // デフォルトで今日の日付を設定
    const today = new Date().toISOString().split('T')[0] || '';
    setFormData({ ...initialFormState, due_date: today });
    setIsModalVisible(true);
  };

  // モーダルを閉じる
  const closeModal = () => {
    setIsModalVisible(false);
    setFormData(initialFormState);
  };

  // APIは配列を直接返すので、allTasks自体が配列
  const tasks = allTasks || [];

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
        onPress={openModal}
      >
        <Ionicons name="add" size={28} color="white" />
      </TouchableOpacity>

      {/* タスク作成モーダル */}
      <Modal
        visible={isModalVisible}
        animationType="slide"
        transparent={true}
        onRequestClose={closeModal}
      >
        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          className="flex-1"
        >
          <View className="flex-1 justify-end bg-black/50">
            <View className="rounded-t-3xl bg-white p-6">
              {/* ヘッダー */}
              <View className="mb-6 flex-row items-center justify-between">
                <Text className="text-xl font-bold text-gray-800">
                  新しいタスク
                </Text>
                <TouchableOpacity onPress={closeModal}>
                  <Ionicons name="close" size={24} color="#6b7280" />
                </TouchableOpacity>
              </View>

              {/* タイトル入力 */}
              <View className="mb-4">
                <Text className="mb-2 text-sm font-medium text-gray-700">
                  タイトル *
                </Text>
                <TextInput
                  className="rounded-lg border border-gray-300 px-4 py-3 text-base"
                  placeholder="タスクのタイトル"
                  value={formData.title}
                  onChangeText={(text) =>
                    setFormData((prev) => ({ ...prev, title: text }))
                  }
                />
              </View>

              {/* 説明入力 */}
              <View className="mb-4">
                <Text className="mb-2 text-sm font-medium text-gray-700">
                  説明
                </Text>
                <TextInput
                  className="rounded-lg border border-gray-300 px-4 py-3 text-base"
                  placeholder="タスクの説明（任意）"
                  value={formData.description}
                  onChangeText={(text) =>
                    setFormData((prev) => ({ ...prev, description: text }))
                  }
                  multiline
                  numberOfLines={3}
                />
              </View>

              {/* 期限入力 */}
              <View className="mb-4">
                <Text className="mb-2 text-sm font-medium text-gray-700">
                  期限 * (YYYY-MM-DD)
                </Text>
                <TextInput
                  className="rounded-lg border border-gray-300 px-4 py-3 text-base"
                  placeholder="2024-12-31"
                  value={formData.due_date}
                  onChangeText={(text) =>
                    setFormData((prev) => ({ ...prev, due_date: text }))
                  }
                  keyboardType="numbers-and-punctuation"
                />
              </View>

              {/* 優先度選択 */}
              <View className="mb-6">
                <Text className="mb-2 text-sm font-medium text-gray-700">
                  優先度
                </Text>
                <View className="flex-row gap-2">
                  {[
                    { key: 'low', label: '低', color: 'bg-green-100 border-green-300', activeColor: 'bg-green-500' },
                    { key: 'medium', label: '中', color: 'bg-yellow-100 border-yellow-300', activeColor: 'bg-yellow-500' },
                    { key: 'high', label: '高', color: 'bg-red-100 border-red-300', activeColor: 'bg-red-500' },
                  ].map((item) => (
                    <TouchableOpacity
                      key={item.key}
                      className={`flex-1 items-center rounded-lg border py-3 ${
                        formData.priority === item.key
                          ? item.activeColor
                          : item.color
                      }`}
                      onPress={() =>
                        setFormData((prev) => ({
                          ...prev,
                          priority: item.key as PriorityType,
                        }))
                      }
                    >
                      <Text
                        className={`font-medium ${
                          formData.priority === item.key
                            ? 'text-white'
                            : 'text-gray-700'
                        }`}
                      >
                        {item.label}
                      </Text>
                    </TouchableOpacity>
                  ))}
                </View>
              </View>

              {/* 作成ボタン */}
              <TouchableOpacity
                className={`items-center rounded-lg py-4 ${
                  createMutation.isPending ? 'bg-gray-400' : 'bg-primary-600'
                }`}
                onPress={handleCreate}
                disabled={createMutation.isPending}
              >
                <Text className="text-base font-semibold text-white">
                  {createMutation.isPending ? '作成中...' : 'タスクを作成'}
                </Text>
              </TouchableOpacity>
            </View>
          </View>
        </KeyboardAvoidingView>
      </Modal>
    </View>
  );
}
