import { defineEnvVars } from '@sveltejs/kit/env';

export const variables = defineEnvVars({
	API_URL: {
		schema: (value) => {
			const url = new URL(value || 'http://127.0.0.1:8080');
			if (
				!['http:', 'https:'].includes(url.protocol) ||
				url.username ||
				url.password ||
				url.pathname !== '/' ||
				url.search ||
				url.hash
			)
				throw new Error('API_URL must be an HTTP(S) origin');
			return url.origin;
		},
	},
});
