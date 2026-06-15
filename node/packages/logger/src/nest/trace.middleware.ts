import { Injectable, type NestMiddleware } from '@nestjs/common';
import type { NextFunction, Request, Response } from 'express';
import { ensureTraceparent } from '../trace';

/** Ensures W3C traceparent on every HTTP request (preserve inbound or generate). */
@Injectable()
export class TraceContextMiddleware implements NestMiddleware {
  use(req: Request, res: Response, next: NextFunction): void {
    ensureTraceparent(req, res);
    next();
  }
}
