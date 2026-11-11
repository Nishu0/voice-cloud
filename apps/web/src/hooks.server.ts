import { redirect } from '@sveltejs/kit';
import type { Handle } from '@sveltejs/kit/hooks';
import { api, sessionCookie } from '#lib/server/api.js';
import type { User } from '#lib/api-types.js';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.token = event.cookies.get(sessionCookie);
	if (!['/login', '/signup', '/healthz'].includes(event.url.pathname)) {
		if (!event.locals.token) redirect(303, '/login');
		event.locals.user = await api<User>(event, 'auth/me');
	}
	const response = await resolve(event);
	response.headers.set('Cache-Control', 'no-store');
	return response;
};
