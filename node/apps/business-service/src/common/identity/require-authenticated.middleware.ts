import { Injectable, NestMiddleware, UnauthorizedException } from '@nestjs/common';
import type { NextFunction, Response } from 'express';
import { isAuthenticated } from './identity';
import { IDENTITY_REQUEST_KEY, type RequestWithIdentity } from './identity.middleware';

/** Rejects requests without a trusted user id (after IdentityMiddleware). */
@Injectable()
export class RequireAuthenticatedMiddleware implements NestMiddleware {
  use(req: RequestWithIdentity, _res: Response, next: NextFunction): void {
    const id = req[IDENTITY_REQUEST_KEY];
    if (!isAuthenticated(id)) {
      throw new UnauthorizedException({
        code: 'auth.unauthenticated',
        message: 'authentication required',
      });
    }
    next();
  }
}
