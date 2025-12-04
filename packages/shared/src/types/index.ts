/**
 * 共通型定義
 */

// ユーザー関連
export interface User {
  id: string;
  email: string;
  createdAt: Date;
  updatedAt: Date;
}

// 作物関連
export type CropStatus = 'Seedling' | 'Vegetative' | 'Flowering' | 'Fruiting' | 'Harvested';

export interface Crop {
  id: string;
  userId: string;
  name: string;
  variety?: string;
  plantedDate: Date;
  expectedHarvestDate?: Date;
  status: CropStatus;
  createdAt: Date;
  updatedAt: Date;
}

// 成長記録関連
export type GrowthStage = 'Seedling' | 'Vegetative' | 'Flowering' | 'Fruiting';

export interface GrowthRecord {
  id: string;
  cropId: string;
  recordDate: Date;
  growthStage: GrowthStage;
  notes?: string;
  imageUrl?: string;
  createdAt: Date;
}

// 収穫関連
export type HarvestQuality = 'Excellent' | 'Good' | 'Average' | 'Poor';

export interface Harvest {
  id: string;
  cropId: string;
  harvestDate: Date;
  quantity: number;
  unit: string;
  quality: HarvestQuality;
  createdAt: Date;
}

// 区画関連
export type PlotStatus = 'Available' | 'Occupied' | 'Fallow';
export type SoilType = 'Clay' | 'Sandy' | 'Loamy' | 'Silty' | 'Peaty' | 'Chalky';
export type SunlightCondition = 'FullSun' | 'PartialShade' | 'FullShade';

export interface Plot {
  id: string;
  userId: string;
  name: string;
  width: number;
  height: number;
  positionX?: number;
  positionY?: number;
  soilType?: SoilType;
  sunlight?: SunlightCondition;
  status: PlotStatus;
  createdAt: Date;
  updatedAt: Date;
}

// タスク関連
export type TaskStatus = 'Pending' | 'Completed' | 'Overdue';
export type TaskPriority = 'Low' | 'Medium' | 'High';
export type RecurrenceFrequency = 'Daily' | 'Weekly' | 'Monthly';

export interface TaskRecurrence {
  frequency: RecurrenceFrequency;
  interval: number;
  maxOccurrences?: number;
  endDate?: Date;
}

export interface Task {
  id: string;
  userId: string;
  name: string;
  description?: string;
  dueDate: Date;
  priority: TaskPriority;
  status: TaskStatus;
  recurrence?: TaskRecurrence;
  cropId?: string;
  plotId?: string;
  createdAt: Date;
  updatedAt: Date;
}

// API レスポンス関連
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: ApiError;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

// 通知関連
export type NotificationType = 'TaskReminder' | 'HarvestReminder' | 'GrowthRecord' | 'OverdueAlert';

export interface NotificationSettings {
  pushEnabled: boolean;
  emailEnabled: boolean;
  taskReminders: boolean;
  harvestReminders: boolean;
  growthRecordNotifications: boolean;
}
