import { Body, Controller, Delete, Get, HttpCode, Param, ParseUUIDPipe, Patch, Post } from '@nestjs/common';
import type {
  CreateItemRequest,
  CreateItemResponse,
  GetItemResponse,
  ListItemsResponse,
  UpdateItemRequest,
} from '@ting/api';
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

  @Get(':id')
  get(
    @CurrentIdentity() actor: Identity,
    @Param('id', ParseUUIDPipe) id: string,
  ): Promise<GetItemResponse> {
    return this.items.get(actor, id);
  }

  @Patch(':id')
  update(
    @CurrentIdentity() actor: Identity,
    @Param('id', ParseUUIDPipe) id: string,
    @Body() body: UpdateItemRequest,
  ): Promise<GetItemResponse> {
    return this.items.update(actor, id, body);
  }

  @Delete(':id')
  @HttpCode(204)
  remove(
    @CurrentIdentity() actor: Identity,
    @Param('id', ParseUUIDPipe) id: string,
  ): Promise<void> {
    return this.items.remove(actor, id);
  }
}
