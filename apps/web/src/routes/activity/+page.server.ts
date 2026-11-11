import { error } from '@sveltejs/kit';
import { api } from '#lib/server/api.js';
import type { RequestRecord, Attempt, Usage } from '#lib/api-types.js';
import type { PageServerLoad } from './$types';
export const load: PageServerLoad = async (event) => {
	const offset = Number(event.url.searchParams.get('offset') || 0);
	if (!Number.isSafeInteger(offset) || offset < 0) error(400, 'Invalid offset');
	const requests = await api<RequestRecord[]>(event, `requests?limit=50&offset=${offset}`);
	const id = event.url.searchParams.get('request');
	const selected = id
		? await api<RequestRecord>(event, `requests/${encodeURIComponent(id)}`)
		: null;
	const [attempts, usage] = id
		? await Promise.all([
				api<Attempt[]>(event, `requests/${encodeURIComponent(id)}/attempts`),
				api<Usage[]>(event, `requests/${encodeURIComponent(id)}/usage`),
			])
		: [[], []];
	return { requests, offset, selected, attempts, usage };
};
