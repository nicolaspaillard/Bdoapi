import { Test, TestingModule } from "@nestjs/testing";
import { PearlItemsService } from "./pearl-items.service";

describe("PearlItemsService", () => {
  let service: PearlItemsService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [PearlItemsService],
    }).compile();

    service = module.get<PearlItemsService>(PearlItemsService);
  });

  it("should be defined", () => {
    expect(service).toBeDefined();
  });
});
