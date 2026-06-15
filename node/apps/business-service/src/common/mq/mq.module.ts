import { Global, Module } from '@nestjs/common';
import { JobPublisher } from './job-publisher';

@Global()
@Module({
  providers: [JobPublisher],
  exports: [JobPublisher],
})
export class MqModule {}
