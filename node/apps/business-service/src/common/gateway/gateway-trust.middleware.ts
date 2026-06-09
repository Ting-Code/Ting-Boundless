import { Injectable, NestMiddleware, UnauthorizedException } from '@nestjs/common';
import type { NextFunction, Request, Response } from 'express';

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

/** Rejects traffic not forwarded by Gateway (X-Internal-Token). Skipped when unset (local dev). */
@Injectable()
export class GatewayTrustMiddleware implements NestMiddleware {
  use(req: Request, _res: Response, next: NextFunction): void {
    const token = process.env.INTERNAL_API_TOKEN?.trim() ?? '';
    if (!token || PROBE_PATHS.has(req.path)) {
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
