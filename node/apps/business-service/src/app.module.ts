import { MiddlewareConsumer, Module, NestModule, RequestMethod } from '@nestjs/common';
import { BusinessModule } from './business/business.module';
import { ItemsModule } from './items/items.module';
import { HttpExceptionFilter } from './common/filters/http-exception.filter';
import { HealthModule } from './common/health/health.module';
import { GatewayTrustMiddleware } from './common/gateway/gateway-trust.middleware';
import { IdentityMiddleware } from './common/identity/identity.middleware';
import { RequireAuthenticatedMiddleware } from './common/identity/require-authenticated.middleware';
import { DrizzleModule } from './db/drizzle.module';
import { MqModule } from './common/mq/mq.module';
import { APP_FILTER, APP_INTERCEPTOR } from '@nestjs/core';
import { LoggingInterceptor, TraceContextMiddleware } from '@ting/logger';

@Module({
  imports: [DrizzleModule, MqModule, HealthModule, BusinessModule, ItemsModule],
  providers: [
    {
      provide: APP_FILTER,
      useClass: HttpExceptionFilter,
    },
    {
      provide: APP_INTERCEPTOR,
      useClass: LoggingInterceptor,
    },
  ],
})
export class AppModule implements NestModule {
  configure(consumer: MiddlewareConsumer): void {
    consumer
      .apply(TraceContextMiddleware, GatewayTrustMiddleware, IdentityMiddleware)
      .forRoutes('*');

    consumer
      .apply(RequireAuthenticatedMiddleware)
      .exclude(
        { path: 'healthz', method: RequestMethod.GET },
        { path: 'readyz', method: RequestMethod.GET },
        { path: 'metrics', method: RequestMethod.GET },
        { path: 'v1/business/ping', method: RequestMethod.GET },
      )
      .forRoutes('*');
  }
}
