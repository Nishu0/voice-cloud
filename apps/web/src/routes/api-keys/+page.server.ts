import { error } from '@sveltejs/kit';
import { api, actionError } from '#lib/server/api.js';
import type { APIKey } from '#lib/api-types.js';
import type { PageServerLoad, Actions } from './$types';
export const load: PageServerLoad = async (event) => ({
	keys: await api<APIKey[]>(event, 'api-keys'),
});
export const actions: Actions = {
	create: async (event) => {
		try {
			const form = await event.request.formData(),
				days = Number(form.get('days'));
			if (![7, 30, 90].includes(days)) error(400, 'Invalid expiry');
			const result = await api<{ key: string }>(event, 'api-keys', {
				method: 'POST',
				body: JSON.stringify({
					name: form.get('name'),
					expires_at: Date.now() + days * 86400000,
					default_policy_id: form.get('policy') ? Number(form.get('policy')) : null,
				}),
			});
			return { key: result.key };
		} catch (problem) {
			return actionError(problem);
		}
	},
	revoke: async (event) => {
		try {
			const form = await event.request.formData(),
				id = String(form.get('id'));
			if (!/^[1-9]\d*$/.test(id)) error(400, 'Invalid key');
			await api(event, `api-keys/${id}`, { method: 'DELETE' });
			return { saved: true };
		} catch (problem) {
			return actionError(problem);
		}
	},
};
