# @ting/logger

ECS-style JSON logger for Node services. Shape mirrors `platform-contracts/schemas/logging.schema.json` and `go/pkg/logger`.

## Nest

```typescript
import { configureLogger, createNestLoggerAdapter } from '@ting/logger';
import { LoggingInterceptor } from '@ting/logger/nest';

configureLogger({ service: 'business-service', level: process.env.LOG_LEVEL });
Logger.overrideLogger(createNestLoggerAdapter());

// AppModule providers:
{ provide: APP_INTERCEPTOR, useClass: LoggingInterceptor }
```

Access logs use `message=http_request` with `request_id`, `trace_id`, and normalized `url.path`.
