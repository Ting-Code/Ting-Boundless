import {
  BadRequestException,
  Inject,
  Injectable,
  NotFoundException,
  ServiceUnavailableException,
  UnauthorizedException,
} from '@nestjs/common';
import type {
  BusinessItem,
  CreateItemRequest,
  CreateItemResponse,
  GetItemResponse,
  ListItemsResponse,
  UpdateItemRequest,
} from '@ting/api';
import { and, desc, eq } from 'drizzle-orm';
import type { Identity } from '../common/identity/identity';
import { JobPublisher } from '../common/mq/job-publisher';
import type { AppDatabase } from '../db/drizzle.module';
import { DRIZZLE } from '../db/drizzle.tokens';
import { items, outbox } from '../db/schema';

function toDto(row: typeof items.$inferSelect): BusinessItem {
  return {
    id: row.id,
    tenant_id: row.tenantId,
    title: row.title,
    body: row.body,
    created_by: row.createdBy,
    created_at: row.createdAt.toISOString(),
    updated_at: row.updatedAt.toISOString(),
  };
}

@Injectable()
export class ItemsService {
  constructor(
    @Inject(DRIZZLE) private readonly db: AppDatabase | null,
    private readonly jobs: JobPublisher,
  ) {}

  private requireDb(): AppDatabase {
    if (!this.db) {
      throw new ServiceUnavailableException({
        code: 'db.not_configured',
        message: 'database not configured',
      });
    }
    return this.db;
  }

  private requireActor(id: Identity): Identity {
    if (!id.userId) {
      throw new UnauthorizedException({
        code: 'auth.unauthenticated',
        message: 'missing trusted identity (call through Gateway)',
      });
    }
    return id;
  }

  async list(actor: Identity): Promise<ListItemsResponse> {
    const db = this.requireDb();
    this.requireActor(actor);
    const tenantId = actor.tenantId ?? '';

    const rows = await db
      .select()
      .from(items)
      .where(eq(items.tenantId, tenantId))
      .orderBy(desc(items.createdAt))
      .limit(100);

    return { items: rows.map(toDto) };
  }

  async create(actor: Identity, input: CreateItemRequest): Promise<CreateItemResponse> {
    const db = this.requireDb();
    this.requireActor(actor);

    const title = input.title?.trim();
    if (!title) {
      throw new BadRequestException({
        code: 'validation.title_required',
        message: 'title is required',
      });
    }

    const tenantId = actor.tenantId ?? '';
    const body = (input.body ?? '').trim();

    const item = await db.transaction(async (tx) => {
      const [row] = await tx
        .insert(items)
        .values({
          tenantId,
          title,
          body,
          createdBy: actor.userId,
        })
        .returning();

      if (!row) {
        throw new ServiceUnavailableException({
          code: 'db.insert_failed',
          message: 'failed to create item',
        });
      }

      await tx.insert(outbox).values({
        eventType: 'business.item.created',
        payload: {
          item_id: row.id,
          tenant_id: tenantId,
          actor_user_id: actor.userId,
          title: row.title,
        },
      });

      return row;
    });

    void this.jobs.publish({
      type: 'business.item.created',
      actor_user_id: actor.userId,
      tenant_id: tenantId,
      payload: {
        item_id: item.id,
        title: item.title,
      },
    });

    return { item: toDto(item) };
  }

  async get(actor: Identity, id: string): Promise<GetItemResponse> {
    const row = await this.findOwned(actor, id);
    return { item: toDto(row) };
  }

  async update(actor: Identity, id: string, input: UpdateItemRequest): Promise<GetItemResponse> {
    const db = this.requireDb();
    this.requireActor(actor);

    const title = input.title?.trim();
    const body = input.body?.trim();
    if (title === undefined && body === undefined) {
      throw new BadRequestException({
        code: 'validation.nothing_to_update',
        message: 'provide title and/or body',
      });
    }
    if (title !== undefined && !title) {
      throw new BadRequestException({
        code: 'validation.title_required',
        message: 'title cannot be empty',
      });
    }

    const tenantId = actor.tenantId ?? '';
    await this.findOwned(actor, id);

    const patch: Partial<typeof items.$inferInsert> = {
      updatedAt: new Date(),
    };
    if (title !== undefined) {
      patch.title = title;
    }
    if (body !== undefined) {
      patch.body = body;
    }

    const item = await db.transaction(async (tx) => {
      const [row] = await tx
        .update(items)
        .set(patch)
        .where(and(eq(items.id, id), eq(items.tenantId, tenantId)))
        .returning();

      if (!row) {
        throw new NotFoundException({
          code: 'business.item_not_found',
          message: 'item not found',
        });
      }

      await tx.insert(outbox).values({
        eventType: 'business.item.updated',
        payload: {
          item_id: row.id,
          tenant_id: tenantId,
          actor_user_id: actor.userId,
          title: row.title,
        },
      });

      return row;
    });

    void this.jobs.publish({
      type: 'business.item.updated',
      actor_user_id: actor.userId,
      tenant_id: tenantId,
      payload: {
        item_id: item.id,
        title: item.title,
      },
    });

    return { item: toDto(item) };
  }

  async remove(actor: Identity, id: string): Promise<void> {
    const db = this.requireDb();
    this.requireActor(actor);
    const tenantId = actor.tenantId ?? '';

    await this.findOwned(actor, id);

    await db.transaction(async (tx) => {
      const deleted = await tx
        .delete(items)
        .where(and(eq(items.id, id), eq(items.tenantId, tenantId)))
        .returning({ id: items.id });

      if (deleted.length === 0) {
        throw new NotFoundException({
          code: 'business.item_not_found',
          message: 'item not found',
        });
      }

      await tx.insert(outbox).values({
        eventType: 'business.item.deleted',
        payload: {
          item_id: id,
          tenant_id: tenantId,
          actor_user_id: actor.userId,
        },
      });
    });

    void this.jobs.publish({
      type: 'business.item.deleted',
      actor_user_id: actor.userId,
      tenant_id: tenantId,
      payload: { item_id: id },
    });
  }

  private async findOwned(actor: Identity, id: string): Promise<typeof items.$inferSelect> {
    const db = this.requireDb();
    this.requireActor(actor);
    const tenantId = actor.tenantId ?? '';

    const [row] = await db
      .select()
      .from(items)
      .where(and(eq(items.id, id), eq(items.tenantId, tenantId)))
      .limit(1);

    if (!row) {
      throw new NotFoundException({
        code: 'business.item_not_found',
        message: 'item not found',
      });
    }
    return row;
  }
}
