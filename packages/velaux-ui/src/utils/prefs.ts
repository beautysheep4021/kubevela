// readPref reads a remembered value, or the fallback when the browser keeps
// none or refuses storage.
export function readPref<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key);
    return raw === null ? fallback : (JSON.parse(raw) as T);
  } catch (e) {
    return fallback;
  }
}

// writePref remembers a value; a browser that refuses storage keeps it for
// this page only.
export function writePref<T>(key: string, value: T) {
  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch (e) {
    // Nothing to do: the caller's state still holds the value.
  }
}
