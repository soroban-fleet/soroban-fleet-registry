import { themes as prismThemes } from 'prism-react-renderer';
import type { Config } from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Soroban Fleet Registry',
  tagline: 'Indexing, resolution, and verification for CAP-85 externally managed executables',
  favicon: 'img/logo.svg',

  url: 'https://soroban-fleet.github.io',
  baseUrl: '/soroban-fleet-registry/',

  organizationName: 'soroban-fleet',
  projectName: 'soroban-fleet-registry',
  deploymentBranch: 'gh-pages',
  trailingSlash: false,

  onBrokenLinks: 'throw',
  onBrokenAnchors: 'throw',
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'throw',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl:
            'https://github.com/soroban-fleet/soroban-fleet-registry/tree/main/docs-site/',
          showLastUpdateAuthor: false,
          showLastUpdateTime: false,
          versions: {
            current: {
              label: 'v1.0.0',
              path: '',
              banner: 'none',
            },
          },
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  plugins: [
    [
      '@easyops-cn/docusaurus-search-local',
      {
        hashed: true,
        indexDocs: true,
        indexBlog: false,
        indexPages: true,
        docsRouteBasePath: '/docs',
        language: ['en'],
        highlightSearchTermsOnTargetPage: true,
        searchResultLimits: 8,
        searchResultContextMaxLength: 60,
      },
    ],
  ],

  themeConfig: {
    colorMode: {
      defaultMode: 'light',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Soroban Fleet Registry',
      logo: {
        alt: 'Soroban Fleet Registry logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          to: '/docs/introduction/what-is-sfr',
          label: 'Quick Start',
          position: 'left',
        },
        {
          href: 'https://github.com/soroban-fleet/soroban-fleet-registry',
          label: 'GitHub',
          position: 'right',
        },
        {
          href: 'https://github.com/soroban-fleet/soroban-fleet-registry/releases/tag/v1.0.0',
          label: 'v1.0.0',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Introduction',
              to: '/docs/introduction/what-is-sfr',
            },
            {
              label: 'Architecture',
              to: '/docs/architecture/system-architecture',
            },
            {
              label: 'API Reference',
              to: '/docs/api/overview',
            },
            {
              label: 'CLI Reference',
              to: '/docs/reference/cli',
            },
          ],
        },
        {
          title: 'Project',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/soroban-fleet/soroban-fleet-registry',
            },
            {
              label: 'Release v1.0.0',
              href: 'https://github.com/soroban-fleet/soroban-fleet-registry/releases/tag/v1.0.0',
            },
            {
              label: 'License (Apache-2.0)',
              href: 'https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/LICENSE',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Security',
              href: 'https://github.com/soroban-fleet/soroban-fleet-registry/security/policy',
            },
            {
              label: 'Contributing',
              href: 'https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/CONTRIBUTING.md',
            },
            {
              label: 'Issues',
              href: 'https://github.com/soroban-fleet/soroban-fleet-registry/issues',
            },
          ],
        },
      ],
      copyright: `Soroban Fleet Registry v1.0.0 — Apache License 2.0. Not affiliated with the Stellar Development Foundation.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'sql', 'go', 'json', 'yaml'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
