import { api, actionError } from '#lib/server/api.js';
import type { Actions } from './$types';
export const actions: Actions = {
	profile: async (event) => {
		try {
			const form = await event.request.formData();
			await api(event, 'auth/me', {
				method: 'PUT',
				body: JSON.stringify({ name: form.get('name') }),
			});
			return { saved: true };
		} catch (problem) {
			return actionError(problem);
		}
	},
	member: async (event) => {
		try {
			const form = await event.request.formData();
			await api(event, 'users', {
				method: 'POST',
				body: JSON.stringify({
					name: form.get('name'),
					email: form.get('email'),
					password: form.get('password'),
				}),
			});
			return { memberCreated: true };
		} catch (problem) {
			return actionError(problem);
		}
	},
};
