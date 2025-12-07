// =============================================================================
// RegisterScreen - ユーザー登録画面
// =============================================================================
// 新規ユーザー登録処理を提供します。

import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
} from 'react-native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useAuth } from '../../context/AuthContext';
import { authApi } from '../../services/api';
import { AuthStackParamList } from '../../navigation/AppNavigator';

// -----------------------------------------------------------------------------
// Types - 型定義
// -----------------------------------------------------------------------------

type RegisterScreenNavigationProp = NativeStackNavigationProp<AuthStackParamList, 'Register'>;

interface Props {
  navigation: RegisterScreenNavigationProp;
}

// -----------------------------------------------------------------------------
// Component - コンポーネント
// -----------------------------------------------------------------------------

export default function RegisterScreen({ navigation }: Props) {
  const { login } = useAuth();
  const [displayName, setDisplayName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  // 登録処理
  const handleRegister = async () => {
    // バリデーション
    if (!displayName.trim()) {
      Alert.alert('エラー', '表示名を入力してください');
      return;
    }
    if (!email.trim()) {
      Alert.alert('エラー', 'メールアドレスを入力してください');
      return;
    }
    if (!password) {
      Alert.alert('エラー', 'パスワードを入力してください');
      return;
    }
    if (password.length < 8) {
      Alert.alert('エラー', 'パスワードは8文字以上で入力してください');
      return;
    }
    if (password !== confirmPassword) {
      Alert.alert('エラー', 'パスワードが一致しません');
      return;
    }

    setIsLoading(true);
    try {
      const response = await authApi.register({
        email: email.trim(),
        password,
        display_name: displayName.trim(),
      });
      await login(response.token, {
        id: response.user.id,
        email: response.user.email,
        displayName: response.user.display_name,
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : '登録に失敗しました';
      Alert.alert('登録エラー', message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      className="flex-1 bg-white"
    >
      <ScrollView
        contentContainerStyle={{ flexGrow: 1 }}
        keyboardShouldPersistTaps="handled"
      >
        <View className="flex-1 justify-center px-8 py-12">
          {/* ヘッダー */}
          <View className="mb-8 items-center">
            <Text className="text-3xl font-bold text-primary-600">新規登録</Text>
            <Text className="mt-2 text-gray-500">アカウントを作成</Text>
          </View>

          {/* フォーム */}
          <View className="space-y-4">
            {/* 表示名 */}
            <View>
              <Text className="mb-2 text-sm font-medium text-gray-700">
                表示名
              </Text>
              <TextInput
                className="rounded-lg border border-gray-300 px-4 py-3 text-base"
                placeholder="山田 太郎"
                value={displayName}
                onChangeText={setDisplayName}
                autoCorrect={false}
                editable={!isLoading}
              />
            </View>

            {/* メールアドレス */}
            <View className="mt-4">
              <Text className="mb-2 text-sm font-medium text-gray-700">
                メールアドレス
              </Text>
              <TextInput
                className="rounded-lg border border-gray-300 px-4 py-3 text-base"
                placeholder="example@email.com"
                value={email}
                onChangeText={setEmail}
                keyboardType="email-address"
                autoCapitalize="none"
                autoCorrect={false}
                editable={!isLoading}
              />
            </View>

            {/* パスワード */}
            <View className="mt-4">
              <Text className="mb-2 text-sm font-medium text-gray-700">
                パスワード（8文字以上）
              </Text>
              <TextInput
                className="rounded-lg border border-gray-300 px-4 py-3 text-base"
                placeholder="パスワード"
                value={password}
                onChangeText={setPassword}
                secureTextEntry
                editable={!isLoading}
              />
            </View>

            {/* パスワード確認 */}
            <View className="mt-4">
              <Text className="mb-2 text-sm font-medium text-gray-700">
                パスワード（確認）
              </Text>
              <TextInput
                className="rounded-lg border border-gray-300 px-4 py-3 text-base"
                placeholder="パスワード（確認）"
                value={confirmPassword}
                onChangeText={setConfirmPassword}
                secureTextEntry
                editable={!isLoading}
              />
            </View>

            {/* 登録ボタン */}
            <TouchableOpacity
              className={`mt-6 rounded-lg py-4 ${
                isLoading ? 'bg-gray-400' : 'bg-primary-600'
              }`}
              onPress={handleRegister}
              disabled={isLoading}
            >
              {isLoading ? (
                <ActivityIndicator color="white" />
              ) : (
                <Text className="text-center text-lg font-semibold text-white">
                  登録する
                </Text>
              )}
            </TouchableOpacity>
          </View>

          {/* ログインリンク */}
          <View className="mt-8 flex-row items-center justify-center">
            <Text className="text-gray-600">すでにアカウントをお持ちの方は</Text>
            <TouchableOpacity
              onPress={() => navigation.navigate('Login')}
              disabled={isLoading}
            >
              <Text className="ml-1 font-semibold text-primary-600">ログイン</Text>
            </TouchableOpacity>
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
