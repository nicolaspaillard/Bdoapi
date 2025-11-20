import { Test, TestingModule } from "@nestjs/testing";
import { PearlItemsController } from "./pearl-items.controller";
import { PearlItemsService } from "./pearl-items.service";
import { PearlItem } from "./entities/pearl-item.entity";

describe("PearlItemsController", () => {
  let controller: PearlItemsController;
  let service: PearlItemsService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [PearlItemsController],
      providers: [PearlItemsService],
    }).compile();

    service = module.get<PearlItemsService>(PearlItemsService);
    controller = module.get<PearlItemsController>(PearlItemsController);
  });

  it("should be defined", () => {
    expect(controller).toBeDefined();
  });
  describe("find", () => {
    it("should return an array of pearl items", async () => {
      const result = [new PearlItem({ name: "Item1", sold: 10, preorders: 5 })];
      jest.spyOn(service, "findAll").mockResolvedValue(result);
      expect(await controller.find()).toBe(result);
    });
  });
});
