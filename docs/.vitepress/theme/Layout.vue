<script setup lang="ts">
import DefaultTheme from 'vitepress/theme'
import type { Theme } from 'vitepress'
import GlobstarAnimation from './GlobstarAnimation.vue'

const { Layout } = DefaultTheme as Theme
</script>

<template>
  <Layout>
    <template #home-hero-info-before>
      <GlobstarAnimation />
    </template>
  </Layout>
</template>
