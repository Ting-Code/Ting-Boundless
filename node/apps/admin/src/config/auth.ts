/** Map admin router pathname (basename /admin) to Gateway BFF return_to. */
export function adminReturnTo(pathname: string): string {
  const path = pathname.startsWith('/') ? pathname : `/${pathname}`;
  return `/admin${path === '/' ? '/items' : path}`;
}

/** Gateway sign-in entry (dev uses /sign-in/dev when VITE_DEV_LOGIN=true). */
export function signInPath(returnTo = '/admin/items'): string {
  const base =
    import.meta.env.VITE_SIGN_IN_PATH ??
    (import.meta.env.VITE_DEV_LOGIN === 'true' ? '/sign-in/dev' : '/sign-in');
  return `${base}?return_to=${encodeURIComponent(returnTo)}`;
}

export function signOutPath(returnTo = '/admin/items'): string {
  return `/sign-out?return_to=${encodeURIComponent(returnTo)}`;
}

export function redirectToSignIn(returnTo = '/admin/items'): void {
  window.location.href = signInPath(returnTo);
}
