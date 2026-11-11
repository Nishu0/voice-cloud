import { api } from '#lib/server/api.js';
import type { Health, RequestRecord } from '#lib/api-types.js';
import type { PageServerLoad } from './$types';
export const load: PageServerLoad = async (event) => {
	const data = await event.parent();
	const requests = await api<RequestRecord[]>(event, 'requests?limit=100');
	const health: Record<number, Health> = {};
	for (const deployment of data.deployments)
		health[deployment.id] = await api<Health>(event, `deployments/${deployment.id}/health`);
	return { requests, health };
};
