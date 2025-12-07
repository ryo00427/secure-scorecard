// =============================================================================
// LoginScreen - ログイン画面
// =============================================================================
// ユーザーのログイン処理を提供します。

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

type LoginScreenNavigationProp = NativeStackNavigationProp<AuthStackParamList, 'Login'>;

interface Props {
  navigation: LoginScreenNavigationProp;
}

// -----------------------------------------------------------------------------
// Component - コンポーネント
// -----------------------------------------------------------------------------

export default function LoginScreen({ navigation }: Props) {
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  // ログイン処理
  const handleLogin = async () => {
    // バリデーション
    if (!email.trim()) {
      Alert.alert('エラー', 'メールアドレスを入力してください');
      return;
    }
    if (!password) {
      Alert.alert('エラー', 'パスワードを入力してください');
      return;
    }

    setIsLoading(true);
    try {
      const response = await authApi.login({ email: email.trim(), password });
      await login(response.token, {
        id: response.user.id,
        email: response.user.email,
        displayName: response.user.display_name,
      });
    } catch (error) {
      const message = error instanceof Error ? error.message : 'ログインに失敗しました';
      Alert.alert('ログインエラー', message);
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
        <View className="flex-1 justify-center px-8">
          {/* ヘッダー */}
          <View className="mb-12 items-center">
            <Text className="text-3xl font-bold text-primary-600">家庭菜園管理</Text>
            <Text className="mt-2 text-gray-500">Home Garden Management</Text>
          </View>

          {/* フォーム */}
          <View className="space-y-4">
            {/* メールアドレス */}
            <View>
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
                パスワード
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

            {/* ログインボタン */}
            <TouchableOpacity
              className={`mt-6 rounded-lg py-4 ${
                isLoading ? 'bg-gray-400' : 'bg-primary-600'
              }`}
              onPress={handleLogin}
              disabled={isLoading}
            >
              {isLoading ? (
                <ActivityIndicator color="white" />
              ) : (
                <Text className="text-center text-lg font-semibold text-white">
                  ログイン
                </Text>
              )}
            </TouchableOpacity>
          </View>

          {/* 登録リンク */}
          <View className="mt-8 flex-row items-center justify-center">
            <Text className="text-gray-600">アカウントをお持ちでない方は</Text>
            <TouchableOpacity
              onPress={() => navigation.navigate('Register')}
              disabled={isLoading}
            >
              <Text className="ml-1 font-semibold text-primary-600">新規登録</Text>
            </TouchableOpacity>
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
