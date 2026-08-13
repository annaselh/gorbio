import type { ComponentType } from "react";
import type { ModuleManifest, Registry, SlotProps } from "./types";

/**
 * Topological sort over module dependencies — the frontend mirror of the
 * backend's deterministic boot order. Throws on a cycle or a missing dep so a
 * broken module graph fails at boot rather than half-rendering.
 */
export function resolveDependencyOrder(
  mods: ModuleManifest[],
): ModuleManifest[] {
  const byName = new Map(mods.map((m) => [m.name, m]));
  const visited = new Set<string>();
  const ordered: ModuleManifest[] = [];

  function visit(m: ModuleManifest, stack: Set<string>) {
    if (visited.has(m.name)) return;
    if (stack.has(m.name)) {
      throw new Error(
        `Cyclic module dependency: ${[...stack, m.name].join(" -> ")}`,
      );
    }
    stack.add(m.name);
    for (const dep of m.dependencies ?? []) {
      const depMod = byName.get(dep);
      if (!depMod) {
        throw new Error(`Missing module dependency: "${dep}" (needed by "${m.name}")`);
      }
      visit(depMod, stack);
    }
    stack.delete(m.name);
    visited.add(m.name);
    ordered.push(m);
  }

  for (const m of mods) visit(m, new Set());
  return ordered;
}

export function buildRegistry(mods: ModuleManifest[]): Registry {
  const seen = new Set<string>();
  for (const m of mods) {
    if (seen.has(m.name)) throw new Error(`Duplicate module name: "${m.name}"`);
    seen.add(m.name);
  }

  const ordered = resolveDependencyOrder(mods);
  const routes = ordered.flatMap((m) => m.routes ?? []);

  // Array.prototype.sort is stable, so equal sequences fall back to boot order.
  const menuItems = ordered
    .flatMap((m) => m.menu ?? [])
    .sort((a, b) => a.sequence - b.sequence);

  // Many modules may fill the same slot; boot order decides render order.
  const slots: Record<string, ComponentType<SlotProps>[]> = {};
  for (const m of ordered) {
    for (const [slotName, Comp] of Object.entries(m.slots ?? {})) {
      (slots[slotName] ??= []).push(Comp);
    }
  }

  return { ordered, routes, menuItems, slots };
}
