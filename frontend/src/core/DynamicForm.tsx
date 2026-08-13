import type { ComponentType } from "react";

export type FieldType = "string" | "int" | "float" | "bool" | "date" | "text";

export interface FieldSchema {
  name: string;
  label?: string;
  type: FieldType;
  required?: boolean;
  /** Module that contributed the field, if not the model's owner. */
  source?: string;
  /** Added at runtime by an admin (JSONB custom field, PRD-F-04). */
  custom?: boolean;
}

export interface ModelSchema {
  fields: FieldSchema[];
}

const inputClass =
  "w-full rounded-lg border border-hairline bg-surface px-3 py-2 text-sm text-ink " +
  "placeholder:text-ink-muted focus:outline-2 focus:outline-offset-0 focus:outline-brand";

const FIELD_COMPONENTS: Record<FieldType, ComponentType<{ f: FieldSchema }>> = {
  string: ({ f }) => (
    <input id={f.name} name={f.name} type="text" required={f.required} className={inputClass} />
  ),
  text: ({ f }) => (
    <textarea id={f.name} name={f.name} rows={3} required={f.required} className={inputClass} />
  ),
  int: ({ f }) => (
    <input id={f.name} name={f.name} type="number" step="1" required={f.required} className={inputClass} />
  ),
  float: ({ f }) => (
    <input id={f.name} name={f.name} type="number" step="any" required={f.required} className={inputClass} />
  ),
  date: ({ f }) => (
    <input id={f.name} name={f.name} type="date" required={f.required} className={inputClass} />
  ),
  bool: ({ f }) => (
    <input id={f.name} name={f.name} type="checkbox" className="size-4 accent-[var(--color-brand)]" />
  ),
};

function humanise(name: string) {
  return name
    .replace(/[_-]+/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

/**
 * Renders a form from a schema the backend sends, so a custom field added by an
 * admin appears with no frontend change. Unknown types degrade to a text input
 * rather than crashing the page — a new backend field type must not white-screen
 * an older frontend build.
 */
export function DynamicForm({
  schema,
  children,
}: {
  schema: ModelSchema;
  children?: React.ReactNode;
}) {
  return (
    <form className="grid gap-4 sm:grid-cols-2">
      {schema.fields.map((f) => {
        const Field = FIELD_COMPONENTS[f.type] ?? FIELD_COMPONENTS.string;
        return (
          <div key={f.name} className="flex flex-col gap-1.5">
            <label
              htmlFor={f.name}
              className="flex items-center gap-2 text-xs font-medium text-ink-secondary"
            >
              {f.label ?? humanise(f.name)}
              {f.required && (
                <span className="text-status-critical" aria-hidden>
                  *
                </span>
              )}
              {f.custom && (
                <span className="rounded bg-brand-wash px-1.5 py-0.5 text-[10px] font-medium text-brand-strong">
                  custom
                </span>
              )}
            </label>
            <Field f={f} />
          </div>
        );
      })}
      {children}
    </form>
  );
}
