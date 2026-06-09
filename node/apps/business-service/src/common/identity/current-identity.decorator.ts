import { createParamDecorator, ExecutionContext } from '@nestjs/common';
import type { Identity } from './identity';
import { IDENTITY_REQUEST_KEY, type RequestWithIdentity } from './identity.middleware';

export const CurrentIdentity = createParamDecorator(
  (_data: unknown, ctx: ExecutionContext): Identity => {
    const req = ctx.switchToHttp().getRequest<RequestWithIdentity>();
    return req[IDENTITY_REQUEST_KEY];
  },
);
