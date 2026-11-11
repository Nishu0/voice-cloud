import { redirect } from '@sveltejs/kit';
import { api, sessionCookie } from '#lib/server/api.js';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async (event) => {
	await api(event, 'auth/logout', { method: 'POST' });
	event.cookies.delete(sessionCookie, { path: '/' });
	redirect(303, '/login');
};
