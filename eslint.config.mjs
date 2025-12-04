import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import globals from 'globals';

/** @type {import('eslint').Linter.Config[]} */
export default [
  // グローバル設定
  {
    ignores: [
      '**/node_modules/**',
      '**/dist/**',
      '**/build/**',
      '**/coverage/**',
      '**/.turbo/**',
      '**/.expo/**',
      '**/metro.config.js',
      '**/babel.config.js',
      '**/tailwind.config.js',
      '**/jest.config.js',
      '**/jest.setup.js',
      'apps/backend/**', // Go バックエンドは除外
    ],
  },

  // JavaScript/TypeScript 推奨設定
  js.configs.recommended,
  ...tseslint.configs.recommended,

  // TypeScript ファイル設定
  {
    files: ['**/*.ts', '**/*.tsx'],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: {
        ...globals.node,
        ...globals.es2022,
      },
      parserOptions: {
        ecmaFeatures: {
          jsx: true,
        },
      },
    },
    rules: {
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
      ],
      '@typescript-eslint/explicit-function-return-type': 'off',
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/consistent-type-imports': [
        'error',
        { prefer: 'type-imports' },
      ],
    },
  },

  // React Native / Expo ファイル設定
  {
    files: ['apps/mobile/**/*.tsx', 'apps/mobile/**/*.ts'],
    languageOptions: {
      globals: {
        ...globals.browser,
        __DEV__: 'readonly',
      },
    },
    rules: {
      // React 17+ では不要
      'react/react-in-jsx-scope': 'off',
    },
  },

  // JavaScript ファイル設定
  {
    files: ['**/*.js', '**/*.mjs', '**/*.cjs'],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: {
        ...globals.node,
        ...globals.es2022,
      },
    },
    rules: {
      'no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
    },
  },

  // テストファイル設定
  {
    files: ['**/*.test.ts', '**/*.test.tsx', '**/*.spec.ts', '**/*.spec.tsx', '**/tests/**'],
    languageOptions: {
      globals: {
        ...globals.jest,
      },
    },
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
    },
  },
];
