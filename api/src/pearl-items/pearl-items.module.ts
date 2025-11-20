import { Module } from "@nestjs/common";
import { PearlItemsService } from "./pearl-items.service";
import { PearlItemsController } from "./pearl-items.controller";
import { TypeOrmModule } from "@nestjs/typeorm";
import { PearlItem } from "./entities/pearl-item.entity";

@Module({
  imports: [TypeOrmModule.forFeature([PearlItem])],
  controllers: [PearlItemsController],
  providers: [PearlItemsService],
})
export class PearlItemsModule {}
