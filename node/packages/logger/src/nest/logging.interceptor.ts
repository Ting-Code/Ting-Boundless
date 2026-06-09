import {
  type CallHandler,
  type ExecutionContext,
  Injectable,
  type NestInterceptor,
} from '@nestjs/common';
import type { Request, Response } from 'express';
import { tap } from 'rxjs';
import { runWithRequestContext } from '../context';
import { requestContextFromHeaders } from '../headers';
import { logger } from '../logger';
import { normalizePath } from '../pathnorm';

@Injectable()
export class LoggingInterceptor implements NestInterceptor {
  intercept(context: ExecutionContext, next: CallHandler) {
    if (context.getType() !== 'http') {
      return next.handle();
    }

    const http = context.switchToHttp();
    const req = http.getRequest<Request>();
    const res = http.getResponse<Response>();
    const started = Date.now();
    const routePath =
      (req.route as { path?: string } | undefined)?.path ??
      req.path ??
      req.url.split('?')[0] ??
      '/';

    const logCtx = requestContextFromHeaders(req.headers);

    return runWithRequestContext(logCtx, () =>
      next.handle().pipe(
        tap({
          finalize: () => {
            logger.info('http_request', {
              'http.request.method': req.method,
              'url.path': normalizePath(routePath),
              'http.response.status_code': res.statusCode,
              duration_ms: Date.now() - started,
            });
          },
        }),
      ),
    );
  }
}
