import { defineConfig } from "vitepress";

const SITE_TITLE =
  "Globstar by DeepSource - The Open-Source Static Analysis Toolkit";
const SITE_DESCRIPTION =
  "Fast, feature-rich, open-source static analysis toolkit for writing and running code quality and SAST checkers.";

const getSiteUrl = () => {
  if (process.env.VERCEL_URL) {
    return `https://${process.env.VERCEL_PROJECT_PRODUCTION_URL}`;
  }
  return "http://localhost:5173";
};

const SITE_URL = getSiteUrl();
const OG_IMAGE = `${SITE_URL}/img/meta.png`;
const OG_IMAGE_WIDTH = "1200";
const OG_IMAGE_HEIGHT = "630";

const isProd = process.env.NODE_ENV === "production";

export default defineConfig({
  lang: "en-US",
  title: SITE_TITLE,
  titleTemplate: ":title | Globstar",
  description: SITE_DESCRIPTION,
  head: [
    ...(isProd
      ? [
        [
          "script",
          {
            async: "",
            src: "https://www.googletagmanager.com/gtag/js?id=G-Z6WPGB3WNL",
            type: "text/javascript",
          },
        ] as [string, Record<string, string>],
        [
          "script",
          {},
          "window.dataLayer = window.dataLayer || [];\nfunction gtag(){dataLayer.push(arguments);}\ngtag('js', new Date());\ngtag('config', 'G-Z6WPGB3WNL');",
        ] as [string, Record<string, string>, string],
      ]
      : []),

    // Favicons
    [
      "link",
      {
        rel: "icon",
        type: "image/svg+xml",
        sizes: "32x32",
        href: "/favicon.svg",
      },
    ],
    ["link", { rel: "shortcut icon", href: "/favicon.ico" }],
    [
      "link",
      {
        rel: "apple-touch-icon",
        sizes: "180x180",
        href: "/apple-touch-icon.png",
      },
    ],

    ["link", { rel: "manifest", href: "/site.webmanifest" }],
    ["meta", { name: "apple-mobile-web-app-title", content: "Globstar" }],

    // Primary Meta Tags
    ["meta", { name: "title", content: SITE_TITLE }],
    ["meta", { name: "description", content: SITE_DESCRIPTION }],
    [
      "meta",
      {
        name: "keywords",
        content:
          "static analysis, code quality, SAST, security analysis, developer tools",
      },
    ],
    ["link", { rel: "canonical", href: SITE_URL }],
    // Go imports
    ["meta", { name: "go-import", content: "globstar.dev git https://github.com/DeepSourceCorp/globstar" }],

    // Open Graph / Facebook
    ["meta", { property: "og:type", content: "website" }],
    ["meta", { property: "og:url", content: SITE_URL }],
    ["meta", { property: "og:title", content: SITE_TITLE }],
    ["meta", { property: "og:description", content: SITE_DESCRIPTION }],
    ["meta", { property: "og:image", content: OG_IMAGE }],
    ["meta", { property: "og:image:type", content: "image/png" }],
    ["meta", { property: "og:image:width", content: OG_IMAGE_WIDTH }],
    ["meta", { property: "og:image:height", content: OG_IMAGE_HEIGHT }],
    ["meta", { property: "og:image:alt", content: "Globstar by DeepSource" }],
    ["meta", { property: "og:site_name", content: "Globstar" }],

    // Twitter
    ["meta", { name: "twitter:card", content: "summary_large_image" }],
    ["meta", { name: "twitter:url", content: SITE_URL }],
    ["meta", { name: "twitter:title", content: "Globstar by DeepSource" }],
    ["meta", { name: "twitter:description", content: SITE_DESCRIPTION }],
    ["meta", { name: "twitter:image", content: OG_IMAGE }],
    ["meta", { name: "twitter:site", content: "@deepsourceHQ" }],
    ["meta", { name: "twitter:creator", content: "@deepsourceHQ" }],
    ["meta", { name: "twitter:image:alt", content: "Globstar by DeepSource" }],
  ],
  sitemap: {
    hostname: SITE_URL,
  },
  appearance: "dark",
  cleanUrls: true,
  themeConfig: {
    siteTitle: false,
    logo: {
      light: "/img/logo-wordmark.svg",
      dark: "/img/logo-wordmark-dark.svg",
    },
    nav: [
      { text: "Docs", link: "/introduction" },
      { text: "Roadmap", link: "/roadmap" },
      {
        text: "Community",
        link: "https://github.com/DeepSourceCorp/globstar/discussions",
      },
    ],
    outline: "deep",
    sidebar: [
      {
        text: "Getting Started",
        items: [
          { text: "Introduction", link: "/introduction" },
          { text: "Quickstart", link: "/quickstart" },
          { text: "CI/CD Integration", link: "/ci-cd-integration" },
          { text: "Supported Languages", link: "/supported-languages" },
          { text: "Roadmap", link: "/roadmap" },
          { text: "Contributing", link: "/contributing" },
        ],
      },
      {
        text: "Reference",
        items: [
          { text: "CLI", link: "/reference/cli" },
          { text: "Configuration", link: "/reference/configuration" },
          { text: "Checker YAML Interface", link: "/reference/checker-yaml" },
          { text: "Checker Go Interface", link: "/reference/checker-go" },
        ],
      },
      {
        text: 'Guides',
        items: [
          { text: 'Writing a checker in YAML', link: '/guides/writing-yaml-checker' },
          { text: 'Writing a checker in Go', link: '/guides/writing-go-checker' },
          { text: 'Contributing built-in checkers', link: '/guides/contributing-built-in-checkers' },
        ]
      },
      {
        text: "Examples",
        items: [
          { text: "JavaScript", link: "/examples/javascript" },
          { text: "Python", link: "/examples/python" }
        ],
      }
    ],
    socialLinks: [
      { icon: "github", link: "https://github.com/DeepSourceCorp/globstar" },
    ],
    search: {
      provider: "local",
    },
    footer: {
      message:
        "Made with &hearts; by <a href='https://deepsource.com'>DeepSource</a>, released under the MIT License.",
      copyright: "&copy; 2025 DeepSource Corp.",
    },
  },
});
