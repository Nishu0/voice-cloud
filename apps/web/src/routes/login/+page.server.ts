import { redirect } from '@sveltejs/kit';
import { api, actionError, setSession } from '#lib/server/api.js';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const status = await api<{ needs_setup: boolean }>(event, 'auth/status');
	if (status.needs_setup) redirect(303, '/signup');
};
export const actions: Actions = {
	default: async (event) => {
		const form = await event.request.formData();
		try {
			const session = await api<{ token: string; expires_at: number }>(event, 'auth/login', {
				method: 'POST',
				body: JSON.stringify({ email: form.get('email'), password: form.get('password') }),
			});
			setSession(event, session.token, session.expires_at);
		} catch (problem) {
			return actionError(problem);
		}
		redirect(303, '/');
	},
};
