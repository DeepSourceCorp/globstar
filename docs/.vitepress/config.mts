import { defineConfig } from 'vitepress'

const SITE_TITLE = 'Globstar by DeepSource - The Open-Source Static Analysis Toolkit'
const SITE_DESCRIPTION = 'Fast, feature-rich, open-source static analysis toolkit for writing and running code quality and SAST checkers.'

// Helper function to get the site URL
const getSiteUrl = () => {
  if (process.env.VERCEL_URL) {
    return `https://${process.env.VERCEL_URL}`
  }
  return 'http://localhost:5173'
}

const SITE_URL = getSiteUrl()
const OG_IMAGE = `${SITE_URL}/img/meta.png`
const OG_IMAGE_WIDTH = '1200'
const OG_IMAGE_HEIGHT = '630'

export default defineConfig({
  lang: 'en-US',
  title: SITE_TITLE,
  description: SITE_DESCRIPTION,
  head: [
    // Favicons
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon.png', media: '(prefers-color-scheme: light)' }],
    ['link', { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-dark.png', media: '(prefers-color-scheme: dark)' }],
    ['link', { rel: 'icon', type: 'image/svg+xml', sizes: '32x32', href: '/favicon.svg' }],
    
    // Primary Meta Tags
    ['meta', { name: 'title', content: SITE_TITLE }],
    ['meta', { name: 'description', content: SITE_DESCRIPTION }],
    ['meta', { name: 'keywords', content: 'static analysis, code quality, SAST, security analysis, developer tools' }],
    ['link', { rel: 'canonical', href: SITE_URL }],
    
    // Open Graph / Facebook
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:url', content: SITE_URL }],
    ['meta', { property: 'og:title', content: SITE_TITLE }],
    ['meta', { property: 'og:description', content: SITE_DESCRIPTION }],
    ['meta', { property: 'og:image', content: OG_IMAGE }],
    ['meta', { property: 'og:image:type', content: 'image/png' }],
    ['meta', { property: 'og:image:width', content: OG_IMAGE_WIDTH }],
    ['meta', { property: 'og:image:height', content: OG_IMAGE_HEIGHT }],
    ['meta', { property: 'og:image:alt', content: 'Globstar by DeepSource' }],
    ['meta', { property: 'og:site_name', content: 'Globstar' }],
    
    // Twitter
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:url', content: SITE_URL }],
    ['meta', { name: 'twitter:title', content: 'Globstar by DeepSource' }],
    ['meta', { name: 'twitter:description', content: SITE_DESCRIPTION }],
    ['meta', { name: 'twitter:image', content: OG_IMAGE }],
    ['meta', { name: 'twitter:site', content: '@deepsourceHQ' }],
    ['meta', { name: 'twitter:creator', content: '@deepsourceHQ' }],
    ['meta', { name: 'twitter:image:alt', content: 'Globstar by DeepSource' }],
  ],
  sitemap: {
    hostname: SITE_URL
  },
  appearance: 'dark',
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
      message: 'Made with &hearts; by DeepSource, released under the MIT License.',
      copyright: '&copy; 2025 DeepSource Corp.',
    }
  }
})
