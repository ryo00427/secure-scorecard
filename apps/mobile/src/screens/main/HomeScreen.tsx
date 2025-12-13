// =============================================================================
// HomeScreen - ホーム画面
// =============================================================================
// ダッシュボード表示とクイックアクションを提供します。

import React from 'react';
import { View, Text, ScrollView, TouchableOpacity, RefreshControl } from 'react-native';
import { useQuery } from '@tanstack/react-query';
import { Ionicons } from '@expo/vector-icons';
import { useAuth } from '../../context/AuthContext';
import { tasksApi, cropsApi } from '../../services/api';

export default function HomeScreen() {
  const { user } = useAuth();

  // 今日のタスクを取得
  const { data: todayTasks, isLoading: tasksLoading, refetch: refetchTasks } = useQuery({
    queryKey: ['tasks', 'today'],
    queryFn: () => tasksApi.getToday(),
  });

  // 作物一覧を取得
  const { data: crops, isLoading: cropsLoading, refetch: refetchCrops } = useQuery({
    queryKey: ['crops'],
    queryFn: () => cropsApi.getAll(),
  });

  const isLoading = tasksLoading || cropsLoading;

  const onRefresh = () => {
    refetchTasks();
    refetchCrops();
  };

  // 栽培中の作物数
  const growingCrops = crops?.crops?.filter(c => c.status === 'growing').length || 0;

  // 今日のタスク数
  const todayTaskCount = todayTasks?.tasks?.length || 0;

  return (
    <ScrollView
      className="flex-1 bg-gray-50"
      refreshControl={
        <RefreshControl refreshing={isLoading} onRefresh={onRefresh} />
      }
    >
      {/* ウェルカムヘッダー */}
      <View className="bg-primary-600 px-6 pb-8 pt-4">
        <Text className="text-lg text-white/80">こんにちは、</Text>
        <Text className="text-2xl font-bold text-white">
          {user?.displayName || 'ゲスト'}さん
        </Text>
      </View>

      {/* サマリーカード */}
      <View className="-mt-4 flex-row justify-between px-4">
        <View className="mr-2 flex-1 rounded-xl bg-white p-4 shadow-sm">
          <View className="flex-row items-center">
            <View className="rounded-full bg-blue-100 p-2">
              <Ionicons name="checkbox-outline" size={24} color="#3b82f6" />
            </View>
            <View className="ml-3">
              <Text className="text-2xl font-bold text-gray-800">{todayTaskCount}</Text>
              <Text className="text-sm text-gray-500">今日のタスク</Text>
            </View>
          </View>
        </View>

        <View className="ml-2 flex-1 rounded-xl bg-white p-4 shadow-sm">
          <View className="flex-row items-center">
            <View className="rounded-full bg-green-100 p-2">
              <Ionicons name="leaf-outline" size={24} color="#16a34a" />
            </View>
            <View className="ml-3">
              <Text className="text-2xl font-bold text-gray-800">{growingCrops}</Text>
              <Text className="text-sm text-gray-500">栽培中</Text>
            </View>
          </View>
        </View>
      </View>

      {/* 今日のタスク */}
      <View className="mt-6 px-4">
        <View className="flex-row items-center justify-between">
          <Text className="text-lg font-bold text-gray-800">今日のタスク</Text>
          <TouchableOpacity>
            <Text className="text-primary-600">すべて見る</Text>
          </TouchableOpacity>
        </View>

        <View className="mt-3">
          {todayTasks?.tasks && todayTasks.tasks.length > 0 ? (
            todayTasks.tasks.slice(0, 3).map((task) => (
              <View
                key={task.id}
                className="mb-2 flex-row items-center rounded-lg bg-white p-4 shadow-sm"
              >
                <View
                  className={`h-3 w-3 rounded-full ${
                    task.priority === 'high'
                      ? 'bg-red-500'
                      : task.priority === 'medium'
                      ? 'bg-yellow-500'
                      : 'bg-green-500'
                  }`}
                />
                <Text className="ml-3 flex-1 text-gray-800">{task.title}</Text>
                <Ionicons name="chevron-forward" size={20} color="#9ca3af" />
              </View>
            ))
          ) : (
            <View className="rounded-lg bg-white p-6 shadow-sm">
              <Text className="text-center text-gray-500">
                今日のタスクはありません
              </Text>
            </View>
          )}
        </View>
      </View>

      {/* 栽培中の作物 */}
      <View className="mt-6 px-4 pb-6">
        <View className="flex-row items-center justify-between">
          <Text className="text-lg font-bold text-gray-800">栽培中の作物</Text>
          <TouchableOpacity>
            <Text className="text-primary-600">すべて見る</Text>
          </TouchableOpacity>
        </View>

        <View className="mt-3">
          {crops?.crops && crops.crops.filter(c => c.status === 'growing').length > 0 ? (
            crops.crops
              .filter(c => c.status === 'growing')
              .slice(0, 3)
              .map((crop) => (
                <View
                  key={crop.id}
                  className="mb-2 flex-row items-center rounded-lg bg-white p-4 shadow-sm"
                >
                  <View className="rounded-full bg-green-100 p-2">
                    <Ionicons name="leaf" size={20} color="#16a34a" />
                  </View>
                  <View className="ml-3 flex-1">
                    <Text className="font-medium text-gray-800">{crop.name}</Text>
                    <Text className="text-sm text-gray-500">{crop.variety}</Text>
                  </View>
                  <Ionicons name="chevron-forward" size={20} color="#9ca3af" />
                </View>
              ))
          ) : (
            <View className="rounded-lg bg-white p-6 shadow-sm">
              <Text className="text-center text-gray-500">
                栽培中の作物はありません
              </Text>
            </View>
          )}
        </View>
      </View>
    </ScrollView>
  );
}
