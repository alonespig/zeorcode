const AUTH_EXPIRED_EVENT = "zeorcode-admin:auth-expired";

export function notifyAuthExpired() {
  window.dispatchEvent(new Event(AUTH_EXPIRED_EVENT));
}

export function onAuthExpired(handler: () => void) {
  window.addEventListener(AUTH_EXPIRED_EVENT, handler);
  return () => window.removeEventListener(AUTH_EXPIRED_EVENT, handler);
}
