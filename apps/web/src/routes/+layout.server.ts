import { api } from '#lib/server/api.js';
import type { Connection, Deployment, Model, Policy } from '#lib/api-types.js';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async (event) => {
	if (!event.locals.user)
		return { user: null, connections: [], models: [], deployments: [], policies: [] };
	const [connections, models, deployments, policies] = await Promise.all([
		api<Connection[]>(event, 'connections'),
		api<Model[]>(event, 'models'),
		api<Deployment[]>(event, 'deployments'),
		api<Policy[]>(event, 'policies'),
	]);
	return { user: event.locals.user, connections, models, deployments, policies };
};
