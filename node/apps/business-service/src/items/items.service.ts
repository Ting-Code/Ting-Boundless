import {
  BadRequestException,
  Inject,
  Injectable,
  ServiceUnavailableException,
  UnauthorizedException,
} from '@nestjs/common';
import type {
  BusinessItem,
  CreateItemRequest,
  CreateItemResponse,
  ListItemsResponse,
} from '@ting/api-types';
import { desc, eq } from 'drizzle-orm';
import type { Identity } from '../common/identity/identity';
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
  constructor(@Inject(DRIZZLE) private readonly db: AppDatabase | null) {}

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

    return { item: toDto(item) };
  }
}
