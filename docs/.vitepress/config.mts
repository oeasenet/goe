import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

export default withMermaid(defineConfig(
  {
    title: 'GOE Framework',
    description: 'Modern Go Application Framework - Built on Fx & Fiber',
    base: '/goe/',
    cleanUrls: true,
    lastUpdated: true,
    head: [
      ['link', { rel: 'icon', href: '/goe/favicon.ico' }],
      ['meta', { name: 'theme-color', content: '#00ADD8' }],
      ['meta', { name: 'og:type', content: 'website' }],
      ['meta', { name: 'og:locale', content: 'en' }],
      ['meta', { name: 'og:site_name', content: 'GOE Framework' }],
    ],
    themeConfig: {
      logo: { src: '/goe_gopher_logo.png', width: 24, height: 24 },
      siteTitle: 'GOE Framework',
      nav: [
        { text: 'Documentation', link: 'https://deepwiki.com/oeasenet/goe' },
        { text: 'GitHub', link: 'https://github.com/oeasenet/goe' }
      ],

      sidebar: {},

      socialLinks: [
        { icon: 'github', link: 'https://github.com/oeasenet/goe' }
      ],

      footer: {
        message: 'Released under the MIT License.',
        copyright: 'Copyright © 2024 GOE Framework'
      },

      editLink: false,

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
  }
))
