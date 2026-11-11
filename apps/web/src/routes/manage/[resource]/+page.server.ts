import { configurationActions, resourceName } from '#lib/server/configuration.js';
import { api } from '#lib/server/api.js';
import type { Language, Target, Health } from '#lib/api-types.js';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const resource = resourceName(event.params.resource);
	const data = await event.parent();
	const languages: Record<number, Language[]> = {},
		targets: Record<number, Target[]> = {},
		health: Record<number, Health> = {};
	if (resource === 'deployments')
		for (const deployment of data.deployments) {
			languages[deployment.id] = await api<Language[]>(
				event,
				`deployments/${deployment.id}/languages`,
			);
			health[deployment.id] = await api<Health>(event, `deployments/${deployment.id}/health`);
		}
	if (resource === 'policies')
		for (const policy of data.policies)
			targets[policy.id] = await api<Target[]>(event, `policies/${policy.id}/targets`);
	return { resource, languages, targets, health };
};
export const actions = configurationActions((event) => String(event.params.resource));
