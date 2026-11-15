/**
 * Reports whether a value is a non-null object, so its properties can be read
 * safely while checking the shape of untrusted JSON.
 *
 * @param value - Any value, typically decoded JSON.
 * @returns Whether `value` is an object (and not `null`).
 */
export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}
