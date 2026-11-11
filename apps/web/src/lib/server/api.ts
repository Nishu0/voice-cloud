import { error, fail, isHttpError, redirect, type RequestEvent } from '@sveltejs/kit';
import { API_URL } from '$app/env/private';

export const sessionCookie = 'voicecloud_session';

export async function api<T>(
	event: RequestEvent,
	path: string,
	init: RequestInit = {},
): Promise<T> {
	let response: Response;
	try {
		response = await fetch(`${API_URL}/v1/${path}`, {
			...init,
			headers: {
				'Content-Type': 'application/json',
				...(event.locals.token ? { Authorization: `Bearer ${event.locals.token}` } : {}),
			},
			signal: AbortSignal.timeout(10000),
			redirect: 'error',
		});
	} catch {
		error(503, 'Backend unavailable. Check the API service.');
	}
	if (!response.ok) {
		if (response.status === 401 && !path.startsWith('auth/login')) {
			event.cookies.delete(sessionCookie, { path: '/' });
			redirect(303, '/login');
		}
		const body = await response.json().catch(() => ({ error: 'Backend request failed' }));
		error(response.status, typeof body.error === 'string' ? body.error : 'Backend request failed');
	}
	return response.status === 204 ? (undefined as T) : response.json();
}

export function actionError(problem: unknown) {
	if (isHttpError(problem)) return fail(problem.status, { error: problem.body.message });
	throw problem;
}

export function setSession(event: RequestEvent, token: string, expires: number) {
	event.cookies.set(sessionCookie, token, {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		secure: event.url.protocol === 'https:',
		expires: new Date(expires),
	});
}
