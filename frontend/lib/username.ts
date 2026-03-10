export const USERNAME_STORAGE_KEY = "bitforge:username";
const LEGACY_USERNAME_STORAGE_KEY = "bitforge:user-handle";
export const USER_KEY_STORAGE_KEY = "bitforge:user-key";

const USERNAME_MIN_LEN = 3;
const USERNAME_MAX_LEN = 24;

export function sanitizeUsername(input: string): string {
  const trimmed = input.trim().replace(/\s+/g, " ");
  const cleaned = trimmed.replace(/[^A-Za-z0-9 ._-]/g, "");
  return cleaned.slice(0, USERNAME_MAX_LEN);
}

export function validateUsername(input: string): string | null {
  const value = sanitizeUsername(input);
  if (value.length < USERNAME_MIN_LEN) {
    return `Username must be at least ${USERNAME_MIN_LEN} characters`;
  }
  if (value.length > USERNAME_MAX_LEN) {
    return `Username must be at most ${USERNAME_MAX_LEN} characters`;
  }
  return null;
}

export function loadStoredUsername(): string {
  if (typeof window === "undefined") {
    return "";
  }
  const current = window.localStorage.getItem(USERNAME_STORAGE_KEY);
  if (current && current.trim()) {
    return sanitizeUsername(current);
  }
  const legacy = window.localStorage.getItem(LEGACY_USERNAME_STORAGE_KEY);
  if (legacy && legacy.trim()) {
    const migrated = sanitizeUsername(legacy);
    if (migrated) {
      saveStoredUsername(migrated);
      return migrated;
    }
  }
  return "";
}

export function saveStoredUsername(input: string): string {
  const value = sanitizeUsername(input);
  window.localStorage.setItem(USERNAME_STORAGE_KEY, value);
  window.localStorage.setItem(LEGACY_USERNAME_STORAGE_KEY, value);
  return value;
}

export function loadOrCreateUserKey(): string {
  if (typeof window === "undefined") {
    return "";
  }
  const existing = window.localStorage.getItem(USER_KEY_STORAGE_KEY);
  if (existing && existing.trim()) {
    return existing;
  }

  let generated = "";
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    generated = crypto.randomUUID();
  } else {
    generated = `bf-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
  }
  window.localStorage.setItem(USER_KEY_STORAGE_KEY, generated);
  return generated;
}
