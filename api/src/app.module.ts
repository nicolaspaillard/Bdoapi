import { Module } from "@nestjs/common";
import { AppController } from "./app.controller";
import { AppService } from "./app.service";
import { PearlItemsModule } from "./pearl-items/pearl-items.module";
import { TypeOrmModule } from "@nestjs/typeorm";
import { PearlItem } from "./pearl-items/entities/pearl-item.entity";

@Module({
  imports: [
    PearlItemsModule,
    TypeOrmModule.forRoot({
      type: "postgres",
      host: process.env.DB_HOST ?? "database",
      port: Number(process.env.DB_PORT ?? "5432"),
      database: process.env.DB_NAME ?? "bdoapi",
      username: process.env.DB_USER ?? "postgres",
      password: process.env.DB_PASSWORD ?? "password",
      entities: [PearlItem],
    }),
  ],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
