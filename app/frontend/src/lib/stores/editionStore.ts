import { writable } from 'svelte/store';

/**
 * True in a build without a Select server, which offers nothing that needs one.
 * Set once at startup, before the session check.
 */
export const localEditionStore = writable(false);
