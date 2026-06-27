import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://rodionlim.github.io',
  base: '/portfolio-manager-go',
  integrations: [
    starlight({
      title: 'Portfolio Manager',
      description:
        'Documentation for Portfolio Manager, a Go application for portfolio valuation, market data, analytics, and LLM integrations.',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/rodionlim/portfolio-manager-go',
        },
      ],
      sidebar: [
        {
          label: 'Start Here',
          items: [
            { label: 'Overview', slug: 'guides/overview' },
            { label: 'Installation', slug: 'guides/installation' },
            { label: 'Quickstart', slug: 'guides/quickstart' },
          ],
        },
        {
          label: 'Concepts',
          items: [
            { label: 'MCP Integration', slug: 'concepts/mcp-integration' },
            { label: 'Market Rotation', slug: 'concepts/market-rotation' },
          ],
        },
      ],
    }),
  ],
});
