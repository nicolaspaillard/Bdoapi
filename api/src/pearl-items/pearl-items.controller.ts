import { Controller, Get, Logger, Query } from "@nestjs/common";
import { PearlItemsService } from "./pearl-items.service";

@Controller("pearl_items")
export class PearlItemsController {
  private readonly logger = new Logger(PearlItemsController.name);

  constructor(private readonly pearlItemsService: PearlItemsService) {}

  @Get()
  async find(@Query("date") date?: string) {
    if (!date) return this.pearlItemsService.findAll();
    const items = await this.pearlItemsService.find(new Date(date));
    this.logger.log(`Retrieved ${items.length} items for date ${date}`);
    return items;
  }
}
