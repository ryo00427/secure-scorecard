// =============================================================================
// ProgressBar - 進捗バーコンポーネント
// =============================================================================
// 作物の成長進捗を視覚的に表示します。

import React from 'react';
import { View } from 'react-native';

interface ProgressBarProps {
  // 進捗率（0-100）
  progress: number;
  // バーの色（デフォルト: 緑）
  color?: string;
  // バーの高さ
  height?: number;
  // 背景色
  backgroundColor?: string;
}

export default function ProgressBar({
  progress,
  color = '#22c55e',
  height = 6,
  backgroundColor = '#e5e7eb',
}: ProgressBarProps) {
  // 進捗率を0-100の範囲に制限
  const clampedProgress = Math.min(100, Math.max(0, progress));

  return (
    <View
      style={{
        height,
        backgroundColor,
        borderRadius: height / 2,
        overflow: 'hidden',
      }}
    >
      <View
        style={{
          height: '100%',
          width: `${clampedProgress}%`,
          backgroundColor: color,
          borderRadius: height / 2,
        }}
      />
    </View>
  );
}
