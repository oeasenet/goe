import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "GOE Framework",
  description: "Modern Go Application Framework - Built on Fx & Fiber",
  base: '/goe/',
  cleanUrls: true,
  lastUpdated: true,
  head: [
    ['link', { rel: 'icon', href: '/favicon.ico' }],
    ['meta', { name: 'theme-color', content: '#00ADD8' }],
    ['meta', { name: 'og:type', content: 'website' }],
    ['meta', { name: 'og:locale', content: 'en' }],
    ['meta', { name: 'og:site_name', content: 'GOE Framework' }],
  ],
  themeConfig: {
    logo: { src: '/logo.svg', width: 24, height: 24 },
    siteTitle: 'GOE Framework',
    
    nav: [
      { text: 'Guide', link: '/guide/introduction', activeMatch: '/guide/' },
      { text: 'Examples', link: '/examples/basic-app', activeMatch: '/examples/' },
      { text: 'API Reference', link: '/reference/modules', activeMatch: '/reference/' },
      { text: 'FAQ', link: '/faq' },
      {
        text: 'v2.0',
        items: [
          { text: 'Changelog', link: '/changelog' },
          { text: 'Migration Guide', link: '/migration' },
        ]
      }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Introduction', link: '/guide/introduction' },
            { text: 'Quick Start', link: '/guide/getting-started' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Project Structure', link: '/guide/project-structure' },
            { text: 'Architecture', link: '/guide/architecture' }
          ]
        },
        {
          text: 'Core Concepts',
          items: [
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Logging', link: '/guide/logging' },
            { text: 'Dependency Injection', link: '/guide/dependency-injection' },
            { text: 'Modules', link: '/guide/modules' },
            { text: 'Error Handling', link: '/guide/error-handling' }
          ]
        },
        {
          text: 'Features',
          items: [
            { text: 'HTTP Server', link: '/guide/http-server' },
            { text: 'Database', link: '/guide/database' },
            { text: 'Caching', link: '/guide/caching' },
            { text: 'Observability', link: '/guide/observability' }
          ]
        },
        {
          text: 'Development',
          items: [
            { text: 'Testing', link: '/guide/testing' },
            { text: 'Best Practices', link: '/guide/best-practices' },
            { text: 'Deployment', link: '/guide/deployment' },
            { text: 'Contributing', link: '/guide/contributing' }
          ]
        }
      ],
      '/examples/': [
        {
          text: 'Basic Examples',
          items: [
            { text: 'Hello World', link: '/examples/basic-app' },
            { text: 'REST API', link: '/examples/rest-api' },
            { text: 'Database CRUD', link: '/examples/database-crud' },
            { text: 'Custom Module', link: '/examples/custom-module' }
          ]
        },
        {
          text: 'Advanced Examples',
          items: [
            { text: 'Authentication', link: '/examples/authentication' },
            { text: 'Real-time Features', link: '/examples/realtime' },
            { text: 'Microservices', link: '/examples/microservices' },
            { text: 'Production Setup', link: '/examples/production' }
          ]
        }
      ],
      '/reference/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Core Modules', link: '/reference/modules' },
            { text: 'Contracts', link: '/reference/contracts' },
            { text: 'Configuration Options', link: '/reference/configuration' },
            { text: 'Middleware', link: '/reference/middleware' },
            { text: 'Utilities', link: '/reference/utilities' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/oeasenet/goe' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024 GOE Framework'
    },

    editLink: {
      pattern: 'https://github.com/oeasenet/goe/edit/v2/docs/:path',
      text: 'Edit this page on GitHub'
    },

    search: {
      provider: 'local'
    },

    outline: {
      level: [2, 3]
    }
  },
  
  markdown: {
    theme: {
      light: 'github-light',
      dark: 'github-dark'
    },
    lineNumbers: true
  }
})
