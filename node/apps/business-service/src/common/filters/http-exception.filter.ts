import {
  ArgumentsHost,
  Catch,
  ExceptionFilter,
  HttpException,
  HttpStatus,
} from '@nestjs/common';
import type { Response } from 'express';
import type { ErrorEnvelope } from '@ting/api-types';
import { IDENTITY_REQUEST_KEY, type RequestWithIdentity } from '../identity/identity.middleware';

@Catch()
export class HttpExceptionFilter implements ExceptionFilter {
  catch(exception: unknown, host: ArgumentsHost): void {
    const ctx = host.switchToHttp();
    const res = ctx.getResponse<Response>();
    const req = ctx.getRequest<RequestWithIdentity>();

    let status = HttpStatus.INTERNAL_SERVER_ERROR;
    let code = 'internal.error';
    let message = 'Internal server error';

    if (exception instanceof HttpException) {
      status = exception.getStatus();
      const body = exception.getResponse();
      if (typeof body === 'string') {
        message = body;
        code = `http.${status}`;
      } else if (typeof body === 'object' && body !== null) {
        const obj = body as Record<string, unknown>;
        message = String(obj.message ?? message);
        code = String(obj.code ?? `http.${status}`);
      }
    }

    const payload: ErrorEnvelope = {
      error: {
        code,
        message,
        request_id: req[IDENTITY_REQUEST_KEY]?.requestId,
      },
    };

    res.status(status).json(payload);
  }
}
