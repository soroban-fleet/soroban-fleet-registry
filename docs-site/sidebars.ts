import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'category',
      label: 'Introduction',
      collapsed: false,
      items: [
        'introduction/what-is-sfr',
        'introduction/problem',
        'introduction/why-cap85',
        'introduction/quick-start',
      ],
    },
    {
      type: 'category',
      label: 'Concepts',
      collapsed: false,
      items: [
        'concepts/fleet',
        'concepts/external-executable',
        'concepts/fleet-lifecycle',
        'concepts/verification-states',
        'concepts/indexing-coverage',
      ],
    },
    {
      type: 'category',
      label: 'Architecture',
      items: [
        'architecture/system-architecture',
        'architecture/data-flow',
        'architecture/indexing-pipeline',
        'architecture/state-model',
        'architecture/failure-recovery',
        'architecture/soroban-boundary',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/inspect-fleet',
        'guides/verify-fleet',
        'guides/inspect-contract',
        'guides/use-cli',
        'guides/use-web-app',
      ],
    },
    {
      type: 'category',
      label: 'API',
      items: [
        'api/overview',
        'api/fleets',
        'api/members',
        'api/history',
        'api/verification',
        'api/contracts',
      ],
    },
    {
      type: 'category',
      label: 'Development',
      items: [
        'development/prerequisites',
        'development/local-development',
        'development/testing',
        'development/fixtures',
        'development/configuration',
        'development/deployment',
      ],
    },
    {
      type: 'category',
      label: 'Operations',
      items: [
        'operations/indexing',
        'operations/health',
        'operations/recovery',
        'operations/troubleshooting',
      ],
    },
    {
      type: 'category',
      label: 'Contributing',
      items: [
        'contributing/workflow',
        'contributing/issues',
        'contributing/pull-requests',
        'contributing/security',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      items: [
        'reference/statuses',
        'reference/cli',
        'reference/environment',
        'reference/glossary',
      ],
    },
  ],
};

export default sidebars;
