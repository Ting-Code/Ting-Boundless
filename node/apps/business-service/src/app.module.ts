import { MiddlewareConsumer, Module, NestModule } from '@nestjs/common';
import { BusinessModule } from './business/business.module';
import { ItemsModule } from './items/items.module';
import { HttpExceptionFilter } from './common/filters/http-exception.filter';
import { HealthModule } from './common/health/health.module';
import { GatewayTrustMiddleware } from './common/gateway/gateway-trust.middleware';
import { IdentityMiddleware } from './common/identity/identity.middleware';
import { DrizzleModule } from './db/drizzle.module';
import { APP_FILTER, APP_INTERCEPTOR } from '@nestjs/core';
import { LoggingInterceptor } from '@ting/logger';

@Module({
  imports: [DrizzleModule, HealthModule, BusinessModule, ItemsModule],
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
    consumer.apply(GatewayTrustMiddleware, IdentityMiddleware).forRoutes('*');
  }
}
