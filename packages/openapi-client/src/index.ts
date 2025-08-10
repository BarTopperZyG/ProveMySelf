import ky from 'ky';
import type { paths } from './types.js';

export type ApiClient = {
  health: {
    get: () => Promise<paths['/health']['get']['responses']['200']['content']['application/json']>;
  };
  projects: {
    list: (params?: paths['/projects']['get']['parameters']['query']) => Promise<paths['/projects']['get']['responses']['200']['content']['application/json']>;
    create: (data: paths['/projects']['post']['requestBody']['content']['application/json']) => Promise<paths['/projects']['post']['responses']['201']['content']['application/json']>;
    get: (projectId: string) => Promise<paths['/projects/{projectId}']['get']['responses']['200']['content']['application/json']>;
  };
};

export function createApiClient(baseUrl: string, options?: { headers?: Record<string, string> }): ApiClient {
  const client = ky.create({
    prefixUrl: baseUrl,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  return {
    health: {
      get: async () => {
        const response = await client.get('health');
        return response.json();
      },
    },
    projects: {
      list: async (params) => {
        const response = await client.get('projects', {
          searchParams: params,
        });
        return response.json();
      },
      create: async (data) => {
        const response = await client.post('projects', {
          json: data,
        });
        return response.json();
      },
      get: async (projectId) => {
        const response = await client.get(`projects/${projectId}`);
        return response.json();
      },
    },
  };
}

export * from './types.js';