import { Body, Controller, Get, HttpCode, Post } from '@nestjs/common';
import type { CreateItemRequest, CreateItemResponse, ListItemsResponse } from '@ting/api-types';
import { CurrentIdentity } from '../common/identity/current-identity.decorator';
import type { Identity } from '../common/identity/identity';
import { ItemsService } from './items.service';

@Controller('v1/business/items')
export class ItemsController {
  constructor(private readonly items: ItemsService) {}

  @Get()
  list(@CurrentIdentity() actor: Identity): Promise<ListItemsResponse> {
    return this.items.list(actor);
  }

  @Post()
  @HttpCode(201)
  create(
    @CurrentIdentity() actor: Identity,
    @Body() body: CreateItemRequest,
  ): Promise<CreateItemResponse> {
    return this.items.create(actor, body);
  }
}
