import { getContext, setContext } from 'svelte';

export type PreviewKey = {
	id: string;
	name: string;
	prefix: string;
	expiry: string;
	revoked: boolean;
};
const context = Symbol('preview');

export function createPreview() {
	const state = $state({
		sample: true,
		name: '44 Chan',
		keys: [] as PreviewKey[],
	});
	setContext(context, state);
	return state;
}

export function usePreview() {
	return getContext<ReturnType<typeof createPreview>>(context);
}
