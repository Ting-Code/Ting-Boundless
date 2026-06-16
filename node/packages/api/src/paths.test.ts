import { describe, expect, it } from 'vitest';
import { auditPaths, filePaths, userPaths } from './paths';
import { ApiError, isApiError, resolveApiUrl } from './request';

describe('auditPaths.eventsQuery', () => {
  it('defaults limit to 50', () => {
    expect(auditPaths.eventsQuery()).toBe('/v1/audit/events?limit=50');
  });

  it('includes optional filters', () => {
    expect(
      auditPaths.eventsQuery({ limit: 10, type: 'user.login.success', source: 'gateway' }),
    ).toBe('/v1/audit/events?limit=10&type=user.login.success&source=gateway');
  });
});

describe('list query helpers', () => {
  it('userPaths.listQuery', () => {
    expect(userPaths.listQuery(20)).toBe('/v1/users/?limit=20');
  });

  it('filePaths.listQuery', () => {
    expect(filePaths.listQuery(100)).toBe('/v1/files/?limit=100');
  });
});

describe('resolveApiUrl', () => {
  it('joins base and path', () => {
    expect(resolveApiUrl('/v1/business/ping', 'http://127.0.0.1:8080')).toBe(
      'http://127.0.0.1:8080/v1/business/ping',
    );
  });

  it('returns absolute urls unchanged', () => {
    expect(resolveApiUrl('https://api.example.com/v1/x')).toBe('https://api.example.com/v1/x');
  });
});

describe('isApiError', () => {
  it('detects ApiError instances', () => {
    const err = new ApiError('auth.unauthenticated', 'unauthorized', 401);
    expect(isApiError(err)).toBe(true);
    expect(isApiError(new Error('nope'))).toBe(false);
  });
});
