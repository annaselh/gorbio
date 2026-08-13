import { baseModule } from "@/modules/base";
import { inventoryModule } from "@/modules/inventory";
import { salesModule } from "@/modules/sales";
import { plannedModules } from "@/modules/planned";
import { buildRegistry } from "./registry";

/**
 * The only file that knows which modules are installed. Strategy 1 (single
 * build) per the architecture doc — swap this for a runtime fetch + federated
 * import when FE-5 lands.
 */
export const registry = buildRegistry([
  baseModule,
  inventoryModule,
  salesModule,
  ...plannedModules,
]);
