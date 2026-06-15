import { Injectable, Logger, OnModuleDestroy } from '@nestjs/common';
import { connect, type Channel, type ChannelModel } from 'amqplib';
import { randomUUID } from 'node:crypto';
import { env, isPlaceholder } from '../../config/env';

export type PlatformJob = {
  id: string;
  type: string;
  time: string;
  actor_user_id?: string;
  tenant_id?: string;
  payload?: Record<string, unknown>;
};

type QueueTopology = {
  exchange: string;
  queue: string;
  dlx: string;
  dlq: string;
  routingKey: string;
};

function defaultTopology(): QueueTopology {
  const prefix = env('RABBITMQ_QUEUE_PREFIX', 'ting.jobs');
  return {
    exchange: prefix,
    queue: `${prefix}.platform`,
    dlx: `${prefix}.dlx`,
    dlq: `${prefix}.platform.dlq`,
    routingKey: 'platform',
  };
}

/** Publishes async jobs to the platform RabbitMQ work queue (optional). */
@Injectable()
export class JobPublisher implements OnModuleDestroy {
  private readonly log = new Logger(JobPublisher.name);
  private conn: ChannelModel | null = null;
  private ch: Channel | null = null;
  private connecting: Promise<void> | null = null;
  private readonly topology = defaultTopology();

  enabled(): boolean {
    const url = env('RABBITMQ_URL', 'amqp://guest:guest@127.0.0.1:5672/').trim();
    return url !== '' && !isPlaceholder(url);
  }

  async publish(job: Omit<PlatformJob, 'id' | 'time'> & { id?: string; time?: string }): Promise<void> {
    if (!this.enabled()) {
      return;
    }
    const body: PlatformJob = {
      id: job.id ?? randomUUID(),
      type: job.type,
      time: job.time ?? new Date().toISOString(),
      actor_user_id: job.actor_user_id,
      tenant_id: job.tenant_id,
      payload: job.payload,
    };
    try {
      await this.ensureChannel();
      if (!this.ch) {
        return;
      }
      this.ch.publish(
        this.topology.exchange,
        this.topology.routingKey,
        Buffer.from(JSON.stringify(body)),
        { contentType: 'application/json', persistent: true },
      );
    } catch (err) {
      this.log.warn(`job publish failed type=${job.type}: ${String(err)}`);
    }
  }

  async onModuleDestroy(): Promise<void> {
    await this.ch?.close().catch(() => undefined);
    await this.conn?.close().catch(() => undefined);
    this.ch = null;
    this.conn = null;
  }

  private async ensureChannel(): Promise<void> {
    if (this.ch) {
      return;
    }
    if (!this.connecting) {
      this.connecting = this.connect();
    }
    await this.connecting;
  }

  private async connect(): Promise<void> {
    const url = env('RABBITMQ_URL', 'amqp://guest:guest@127.0.0.1:5672/');
    this.conn = await connect(url);
    this.ch = await this.conn.createChannel();
    await this.declareTopology(this.ch);
    this.log.log(`connected exchange=${this.topology.exchange}`);
  }

  private async declareTopology(ch: Channel): Promise<void> {
    const t = this.topology;
    await ch.assertExchange(t.exchange, 'direct', { durable: true });
    await ch.assertExchange(t.dlx, 'direct', { durable: true });
    await ch.assertQueue(t.dlq, { durable: true });
    await ch.bindQueue(t.dlq, t.dlx, t.routingKey);
    await ch.assertQueue(t.queue, {
      durable: true,
      arguments: {
        'x-dead-letter-exchange': t.dlx,
        'x-dead-letter-routing-key': t.routingKey,
      },
    });
    await ch.bindQueue(t.queue, t.exchange, t.routingKey);
  }
}
