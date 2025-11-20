import { Entity, Column, PrimaryGeneratedColumn } from "typeorm";

@Entity("pearl_items")
export class PearlItem {
  constructor(partial: Partial<PearlItem>) {
    Object.assign(this, partial);
  }
  @PrimaryGeneratedColumn()
  id: number;

  @Column()
  itemid: number;

  @Column()
  name: string;

  @Column()
  date: Date;

  @Column()
  sold: number;

  @Column()
  preorders: number;
}
