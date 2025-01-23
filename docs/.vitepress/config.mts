import { defineConfig } from 'vitepress'

export default defineConfig({
  lang: 'en-US',
  title: "Globstar by DeepSource - The Static Analysis Toolkit",
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
      { text: 'Docs', link: '/introduction' },
      { text: 'Manifesto', link: '/manifesto' },
    ],
    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Introduction', link: '/introduction' },
          { text: 'Quickstart', link: '/quickstart' },
          { text: 'Writing a Checker', link: '/writing-a-checker' },
          { text: 'CI/CD Integration', link: '/ci-cd-integration' },
          { text: 'Supported Languages', link: '/supported-languages' },
          { text: 'Roadmap', link: '/roadmap' }
        ]
      },
      {
        text: 'Examples',
        items: [
          { text: 'Python', link: '/examples/python' },
          { text: 'JavaScript', link: '/examples/javascript' },
          { text: 'Terraform', link: '/examples/terraform' }
        ]
      },
      {
        text: 'Reference',
        items: [
          { text: 'Configuration', link: '/reference/configuration' },
          { text: 'Checker API', link: '/reference/checker-api' },
          { text: 'CLI', link: '/reference/cli' },
          { text: 'Cross-file Analysis', link: '/reference/cross-file-analysis' },
        ]
      }
    ],
    socialLinks: [
      { icon: 'github', link: 'https://github.com/DeepSourceCorp/globstar' }
    ],
    search: {
      provider: 'local'
    },
    footer: {
      message: 'Globstar is an open-source project by DeepSource, released under the MIT License.',
      copyright: '©️ 2025 DeepSource Corp.',
    }
  }
})
