/**
 * 共通ユーティリティ関数
 */

/**
 * 日付をISO文字列形式（YYYY-MM-DD）に変換
 */
export function formatDateToISO(date: Date): string {
  return date.toISOString().split('T')[0] ?? '';
}

/**
 * 2つの日付の間の日数を計算
 */
export function daysBetween(startDate: Date, endDate: Date): number {
  const oneDay = 24 * 60 * 60 * 1000;
  return Math.round(Math.abs((endDate.getTime() - startDate.getTime()) / oneDay));
}

/**
 * 指定日数後の日付を取得
 */
export function addDays(date: Date, days: number): Date {
  const result = new Date(date);
  result.setDate(result.getDate() + days);
  return result;
}

/**
 * 日付が今日かどうかを判定
 */
export function isToday(date: Date): boolean {
  const today = new Date();
  return (
    date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear()
  );
}

/**
 * 日付が過去かどうかを判定
 */
export function isPast(date: Date): boolean {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const compareDate = new Date(date);
  compareDate.setHours(0, 0, 0, 0);
  return compareDate < today;
}

/**
 * メールアドレスのバリデーション（RFC 5322簡易版）
 */
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * パスワード強度チェック（最低8文字）
 */
export function isStrongPassword(password: string): boolean {
  return password.length >= 8;
}

/**
 * UUIDv4 生成
 */
export function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

/**
 * 配列をチャンクに分割
 */
export function chunk<T>(array: T[], size: number): T[][] {
  const chunks: T[][] = [];
  for (let i = 0; i < array.length; i += size) {
    chunks.push(array.slice(i, i + size));
  }
  return chunks;
}

/**
 * オブジェクトからnull/undefinedのプロパティを除去
 */
export function removeNullish<T extends Record<string, unknown>>(obj: T): Partial<T> {
  return Object.fromEntries(
    Object.entries(obj).filter(([_, v]) => v != null)
  ) as Partial<T>;
}
