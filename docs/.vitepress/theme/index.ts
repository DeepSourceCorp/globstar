import type { Theme } from 'vitepress'
import DefaultTheme from 'vitepress/theme'
import type { DefineComponent } from 'vue'
import Layout from './Layout.vue'
import GlobstarAnimation from './GlobstarAnimation.vue'
import './style.css'

export default {
  extends: DefaultTheme,
  Layout: Layout as DefineComponent,
  enhanceApp({ app }) {
    app.component('GlobstarAnimation', GlobstarAnimation)
  }
} satisfies Theme
