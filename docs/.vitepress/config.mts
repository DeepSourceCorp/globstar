import { defineConfig } from 'vitepress'

export default defineConfig({
  lang: 'en-US',
  title: "Globstar, by DeepSource",
  description: "Fast, feature-rich, open-source static analysis toolkit for writing and running code quality and SAST checkers.",
  head: [
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon.png', media: '(prefers-color-scheme: light)' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-dark.png', media: '(prefers-color-scheme: dark)' }],
    ['link', { rel: 'icon', type: 'image/svg+xml', sizes: '32x32', href: '/favicon.svg' }],
  ],
  cleanUrls: true,
  themeConfig: {
    siteTitle: false,
    logo: {
      light: '/img/logo-wordmark.svg',
      dark: '/img/logo-wordmark-dark.svg'
    },
    nav: [
      { text: 'Home', link: '/' },
      { text: 'Examples', link: '/markdown-examples' }
    ],
    sidebar: [
      {
        text: 'Examples',
        items: [
          { text: 'Markdown Examples', link: '/markdown-examples' },
          { text: 'Runtime API Examples', link: '/api-examples' }
        ]
      }
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/DeepSourceCorp/globstar' },
      { icon: 'twitter', link: 'https://x.com/DeepSourceHQ' }
    ]
  }
})
