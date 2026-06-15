import { Injectable, NestMiddleware, UnauthorizedException } from '@nestjs/common';
import type { NextFunction, Request, Response } from 'express';
import { requireInternalToken, env } from '../../config/env';

const PROBE_PATHS = new Set(['/healthz', '/readyz', '/metrics']);

function internalTokenOK(req: Request, want: string): boolean {
  const header = req.header('x-internal-token');
  if (header === want) {
    return true;
  }
  const auth = req.header('authorization');
  if (auth?.startsWith('Bearer ')) {
    return auth.slice('Bearer '.length) === want;
  }
  return false;
}

/** Rejects traffic not forwarded by Gateway (X-Internal-Token). Skipped in dev when unset. */
@Injectable()
export class GatewayTrustMiddleware implements NestMiddleware {
  use(req: Request, _res: Response, next: NextFunction): void {
    const token = env('INTERNAL_API_TOKEN').trim();
    if (!token) {
      if (!requireInternalToken()) {
        next();
        return;
      }
      throw new UnauthorizedException({
        code: 'internal_auth_misconfigured',
        message: 'INTERNAL_API_TOKEN is required',
      });
    }
    if (PROBE_PATHS.has(req.path)) {
      next();
      return;
    }
    if (!internalTokenOK(req, token)) {
      throw new UnauthorizedException({
        code: 'untrusted_caller',
        message: 'request must come through Gateway',
      });
    }
    next();
  }
}
