import { NextResponse } from 'next/server';
import { createApiClient } from '@provemyself/openapi-client';

export async function GET() {
  try {
    const apiClient = createApiClient('http://localhost:8080/api/v1');
    
    // Test connection to backend API
    const backendHealth = await apiClient.health.get();
    
    return NextResponse.json({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      service: 'player',
      version: '0.1.0',
      backend: {
        status: backendHealth.status,
        version: backendHealth.version,
      },
    });
  } catch (error) {
    console.error('Health check failed:', error);
    
    return NextResponse.json(
      {
        status: 'degraded',
        timestamp: new Date().toISOString(),
        service: 'player',
        version: '0.1.0',
        error: error instanceof Error ? error.message : 'Unknown error',
      },
      { status: 503 }
    );
  }
}