const globals = require('globals')
const tsParser = require('@typescript-eslint/parser')
const tsPlugin = require('@typescript-eslint/eslint-plugin')
const vuePlugin = require('eslint-plugin-vue')
const vueParser = require('vue-eslint-parser')
const prettierPlugin = require('eslint-plugin-prettier')
const prettier = require('eslint-config-prettier')

module.exports = [
  {
    ignores: [
      'dist/**',
      'node_modules/**',
      'flushbonading/**',
      'src/assets/fsSvg.js',
      'src/main.ts',
      '*.d.ts',
      'public/tinymce-dataease-private/langs/**',
      'eslint.config.js'
    ]
  },
  ...vuePlugin.configs['flat/essential'],
  {
    files: ['**/*.{js,jsx,ts,tsx,vue}'],
    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tsParser,
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: {
          jsx: true
        }
      },
      globals: {
        ...globals.browser,
        ...globals.node
      }
    },
    plugins: {
      '@typescript-eslint': tsPlugin,
      prettier: prettierPlugin
    },
    rules: {
      ...tsPlugin.configs.recommended.rules,
      'no-undef': 'off',
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': 'off',
      '@typescript-eslint/no-unused-expressions': 'off',
      '@typescript-eslint/no-empty-object-type': 'off',
      '@typescript-eslint/no-require-imports': 'off',
      'vue/multi-word-component-names': 'off',
      '@typescript-eslint/no-explicit-any': 'off',
      'vue/no-setup-props-destructure': 'off'
    }
  },
  prettier
]
