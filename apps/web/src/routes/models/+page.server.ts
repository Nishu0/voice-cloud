import { api } from '#lib/server/api.js';
import { configurationActions } from '#lib/server/configuration.js';
import type { Health, Language } from '#lib/api-types.js';
import type { PageServerLoad } from './$types';
export const load: PageServerLoad = async (event) => {
	const data = await event.parent(),
		health: Record<number, Health> = {},
		languages: Record<number, Language[]> = {};
	for (const deployment of data.deployments) {
		health[deployment.id] = await api<Health>(event, `deployments/${deployment.id}/health`);
		languages[deployment.id] = await api<Language[]>(
			event,
			`deployments/${deployment.id}/languages`,
		);
	}
	return { health, languages };
};
export const actions = configurationActions(() => 'models');
