import { Injectable, Logger } from "@nestjs/common";
import { InjectRepository } from "@nestjs/typeorm";
import { Repository } from "typeorm";
import { PearlItem } from "./entities/pearl-item.entity";

@Injectable()
export class PearlItemsService {
  private readonly logger = new Logger(PearlItemsService.name);
  constructor(
    @InjectRepository(PearlItem)
    private pearlItemsRepository: Repository<PearlItem>,
  ) {}

  findAll(): Promise<PearlItem[]> {
    return this.pearlItemsRepository.find();
  }

  findOne(id: number): Promise<PearlItem | null> {
    return this.pearlItemsRepository.findOneBy({ id });
  }

  find(date: Date): Promise<PearlItem[]> {
    const dateParam = new Date(date);
    dateParam.setMinutes(0, 0, 0);
    const maxDateSubQuery = this.pearlItemsRepository
      .createQueryBuilder()
      .select("MAX(date)");

    const minDateSubQuery = this.pearlItemsRepository
      .createQueryBuilder()
      .select("MAX(date)")
      .where("date <= :date", { date: dateParam.toISOString() });

    return this.pearlItemsRepository
      .createQueryBuilder("a")
      .select([
        "a.name AS name",
        "a.sold - b.sold AS sold",
        "a.preorders AS preorders",
      ])
      .innerJoin("pearl_items", "b", "a.itemid = b.itemid")
      .where("a.date = (" + maxDateSubQuery.getQuery() + ")")
      .andWhere("b.date = (" + minDateSubQuery.getQuery() + ")")
      .orderBy("a.name")
      .setParameters(minDateSubQuery.getParameters())
      .getRawMany()
      .then((items) => items.map((item) => new PearlItem(item)));
  }
}
