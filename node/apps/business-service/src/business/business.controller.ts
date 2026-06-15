import { Controller, Get, UnauthorizedException } from '@nestjs/common';
import type { BusinessMeResponse, BusinessPingResponse } from '@ting/api';
import { CurrentIdentity } from '../common/identity/current-identity.decorator';
import type { Identity } from '../common/identity/identity';

@Controller('v1/business')
export class BusinessController {
  @Get('ping')
  ping(): BusinessPingResponse {
    return { ok: true, service: 'business-service' };
  }

  @Get('me')
  me(@CurrentIdentity() id: Identity): BusinessMeResponse {
    if (!id.userId) {
      throw new UnauthorizedException({
        code: 'auth.unauthenticated',
        message: 'missing trusted identity (call through Gateway)',
      });
    }
    return {
      user_id: id.userId,
      tenant_id: id.tenantId,
      roles: id.roles,
      scopes: id.scopes,
      subject: id.subject,
      request_id: id.requestId,
    };
  }
}
