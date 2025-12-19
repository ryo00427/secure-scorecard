// =============================================================================
// CalendarScreen - カレンダー画面
// =============================================================================
// デザインファイル: design/stitch_ (4)/screen.png
// 月間カレンダーとタスク/イベントを表示します。

import React, { useState, useMemo } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  SafeAreaView,
  StatusBar,
  ScrollView,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { useNavigation } from '@react-navigation/native';
import type { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { tasksApi } from '../../services/api';

// ナビゲーションの型定義
type RootStackParamList = {
  AddTask: { date?: string };
  TaskDetail: { taskId: number };
};

type NavigationProp = NativeStackNavigationProp<RootStackParamList>;

// 曜日ラベル
const WEEKDAYS = ['日', '月', '火', '水', '木', '金', '土'];

// イベントの色
const EVENT_COLORS = {
  watering: '#3b82f6', // 青
  fertilizing: '#f59e0b', // オレンジ
  harvesting: '#22c55e', // 緑
  default: '#6b7280', // グレー
};

// 日付のイベントデータ（モック）
const mockEvents: Record<string, string[]> = {
  '2': ['watering'],
  '5': ['fertilizing'],
  '9': ['watering'],
  '13': ['harvesting'],
  '15': ['watering', 'fertilizing'],
};

export default function CalendarScreen() {
  const navigation = useNavigation<NavigationProp>();
  const [currentDate, setCurrentDate] = useState(new Date());
  const [selectedDate, setSelectedDate] = useState<number | null>(new Date().getDate());

  // タスク一覧を取得（将来的にカレンダーに表示）
  const { data: _tasksData } = useQuery({
    queryKey: ['tasks'],
    queryFn: () => tasksApi.getAll(),
  });

  // 現在の年月
  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();

  // 月の最初の日と最後の日
  const firstDayOfMonth = new Date(year, month, 1);
  const lastDayOfMonth = new Date(year, month + 1, 0);

  // 月の日数
  const daysInMonth = lastDayOfMonth.getDate();

  // 月の最初の曜日（0=日曜日）
  const firstDayWeekday = firstDayOfMonth.getDay();

  // カレンダーグリッドを生成
  const calendarDays = useMemo(() => {
    const days: (number | null)[] = [];

    // 月初めの空白
    for (let i = 0; i < firstDayWeekday; i++) {
      days.push(null);
    }

    // 日付
    for (let i = 1; i <= daysInMonth; i++) {
      days.push(i);
    }

    return days;
  }, [firstDayWeekday, daysInMonth]);

  // 前月へ
  const handlePrevMonth = () => {
    setCurrentDate(new Date(year, month - 1, 1));
    setSelectedDate(null);
  };

  // 次月へ
  const handleNextMonth = () => {
    setCurrentDate(new Date(year, month + 1, 1));
    setSelectedDate(null);
  };

  // 日付選択
  const handleSelectDate = (day: number) => {
    setSelectedDate(day);
  };

  // タスク追加
  const handleAddTask = () => {
    const dateStr = selectedDate
      ? `${year}-${String(month + 1).padStart(2, '0')}-${String(selectedDate).padStart(2, '0')}`
      : undefined;
    navigation.navigate('AddTask', { date: dateStr });
  };

  // 今日かどうか
  const isToday = (day: number) => {
    const today = new Date();
    return (
      day === today.getDate() &&
      month === today.getMonth() &&
      year === today.getFullYear()
    );
  };

  // 曜日の色
  const getWeekdayColor = (index: number) => {
    if (index === 0) return 'text-red-500'; // 日曜
    if (index === 6) return 'text-blue-500'; // 土曜
    return 'text-gray-700';
  };

  // 日付の色
  const getDayColor = (day: number, _dayIndex: number) => {
    const weekday = (firstDayWeekday + day - 1) % 7;
    if (weekday === 0) return 'text-red-500'; // 日曜
    if (weekday === 6) return 'text-blue-500'; // 土曜
    return 'text-gray-700';
  };

  return (
    <SafeAreaView className="flex-1 bg-white">
      <StatusBar barStyle="dark-content" />

      {/* ヘッダー */}
      <View className="flex-row items-center justify-between px-4 py-4">
        <TouchableOpacity onPress={handlePrevMonth} className="p-2">
          <Ionicons name="chevron-back" size={24} color="#1f2937" />
        </TouchableOpacity>
        <Text className="text-xl font-bold text-gray-800">
          {year}年 {month + 1}月
        </Text>
        <TouchableOpacity onPress={handleNextMonth} className="p-2">
          <Ionicons name="chevron-forward" size={24} color="#1f2937" />
        </TouchableOpacity>
      </View>

      {/* 曜日ヘッダー */}
      <View className="flex-row border-b border-gray-100 px-2 pb-2">
        {WEEKDAYS.map((day, index) => (
          <View key={day} className="flex-1 items-center">
            <Text className={`text-sm font-medium ${getWeekdayColor(index)}`}>
              {day}
            </Text>
          </View>
        ))}
      </View>

      {/* カレンダーグリッド */}
      <ScrollView className="flex-1" showsVerticalScrollIndicator={false}>
        <View className="flex-row flex-wrap px-2 py-2">
          {calendarDays.map((day, index) => {
            if (day === null) {
              return <View key={`empty-${index}`} className="h-16 w-[14.28%]" />;
            }

            const events = mockEvents[String(day)] || [];
            const isSelected = selectedDate === day;

            return (
              <TouchableOpacity
                key={day}
                onPress={() => handleSelectDate(day)}
                className="h-16 w-[14.28%] items-center py-1"
              >
                <View
                  className={`h-10 w-10 items-center justify-center rounded-full ${
                    isSelected
                      ? 'bg-emerald-500'
                      : isToday(day)
                      ? 'bg-gray-100'
                      : ''
                  }`}
                >
                  <Text
                    className={`text-base font-medium ${
                      isSelected
                        ? 'text-white'
                        : getDayColor(day, index)
                    }`}
                  >
                    {day}
                  </Text>
                </View>

                {/* イベントドット */}
                {events.length > 0 && (
                  <View className="mt-1 flex-row">
                    {events.slice(0, 2).map((event, eventIndex) => (
                      <View
                        key={eventIndex}
                        className="mx-0.5 h-1.5 w-1.5 rounded-full"
                        style={{
                          backgroundColor:
                            EVENT_COLORS[event as keyof typeof EVENT_COLORS] ||
                            EVENT_COLORS.default,
                        }}
                      />
                    ))}
                  </View>
                )}
              </TouchableOpacity>
            );
          })}
        </View>

        {/* 選択された日のタスク */}
        {selectedDate && (
          <View className="border-t border-gray-100 px-4 py-4">
            <Text className="mb-3 text-lg font-bold text-gray-800">
              {month + 1}月{selectedDate}日のタスク
            </Text>

            {(mockEvents[String(selectedDate)] ?? []).length > 0 ? (
              (mockEvents[String(selectedDate)] ?? []).map((event, index) => (
                <View
                  key={index}
                  className="mb-2 flex-row items-center rounded-lg bg-gray-50 p-3"
                >
                  <View
                    className="mr-3 h-3 w-3 rounded-full"
                    style={{
                      backgroundColor:
                        EVENT_COLORS[event as keyof typeof EVENT_COLORS] ||
                        EVENT_COLORS.default,
                    }}
                  />
                  <Text className="flex-1 text-gray-700">
                    {event === 'watering'
                      ? '水やり'
                      : event === 'fertilizing'
                      ? '施肥'
                      : event === 'harvesting'
                      ? '収穫'
                      : event}
                  </Text>
                </View>
              ))
            ) : (
              <View className="items-center py-8">
                <Text className="text-gray-500">この日のタスクはありません</Text>
              </View>
            )}
          </View>
        )}

        {/* 余白 */}
        <View className="h-24" />
      </ScrollView>

      {/* FAB（タスク追加ボタン） */}
      <TouchableOpacity
        onPress={handleAddTask}
        className="absolute bottom-6 right-6 h-14 w-14 items-center justify-center rounded-full bg-emerald-600 shadow-lg"
        activeOpacity={0.8}
      >
        <Ionicons name="add" size={28} color="white" />
      </TouchableOpacity>
    </SafeAreaView>
  );
}
