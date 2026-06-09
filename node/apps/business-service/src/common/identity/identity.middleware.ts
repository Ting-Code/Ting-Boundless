import { Injectable, NestMiddleware } from '@nestjs/common';
import type { NextFunction, Request, Response } from 'express';
import { identityFromHeaders, type Identity } from './identity';

export const IDENTITY_REQUEST_KEY = 'tingIdentity';

export type RequestWithIdentity = Request & { [IDENTITY_REQUEST_KEY]: Identity };

@Injectable()
export class IdentityMiddleware implements NestMiddleware {
  use(req: RequestWithIdentity, _res: Response, next: NextFunction): void {
    req[IDENTITY_REQUEST_KEY] = identityFromHeaders(req.headers);
    next();
  }
}
