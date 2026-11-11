import { error, type RequestEvent } from '@sveltejs/kit';
import { api, actionError } from './api.js';

export function resourceName(value: string) {
	if (!['connections', 'models', 'deployments', 'policies'].includes(value))
		error(404, 'Unknown resource');
	return value;
}
export function configurationActions(resource: (event: RequestEvent) => string) {
	return {
		save: async (event: RequestEvent) => {
			try {
				const kind = resourceName(resource(event)),
					form = await event.request.formData();
				const id = String(form.get('id') || '');
				if (id && !/^[1-9]\d*$/.test(id)) error(400, 'Invalid ID');
				const text = (key: string) => String(form.get(key) ?? '');
				const object = (key: string) => {
					try {
						const value = JSON.parse(text(key) || '{}');
						if (value && typeof value === 'object' && !Array.isArray(value)) return value;
					} catch {}
					error(400, `${key} must be a JSON object`);
				};
				let body: Record<string, unknown> = { name: text('name') };
				if (kind === 'connections')
					body = {
						...body,
						adapter: text('adapter'),
						location_kind: text('location_kind'),
						base_url: text('base_url'),
						credential_ref: text('credential_ref') || null,
						enabled: form.has('enabled'),
					};
				if (kind === 'models')
					body = {
						...body,
						publisher: text('publisher'),
						task: text('task'),
						version: text('version'),
					};
				if (kind === 'deployments')
					body = {
						...body,
						model_id: Number(text('model_id')),
						connection_id: Number(text('connection_id')),
						upstream_model_id: text('upstream_model_id'),
						enabled: form.has('enabled'),
						capabilities_json: object('capabilities_json'),
						pricing_json: object('pricing_json'),
						settings_json: object('settings_json'),
					};
				if (kind === 'policies')
					body = {
						...body,
						task: text('task'),
						objective: text('objective'),
						fallback_enabled: form.has('fallback_enabled'),
						max_attempts: Number(text('max_attempts')),
						constraints_json: object('constraints_json'),
					};
				await api(event, `${kind}${id ? '/' + id : ''}`, {
					method: id ? 'PUT' : 'POST',
					body: JSON.stringify(body),
				});
				return { saved: true };
			} catch (problem) {
				return actionError(problem);
			}
		},
		remove: async (event: RequestEvent) => {
			try {
				const form = await event.request.formData(),
					id = String(form.get('id'));
				if (!/^[1-9]\d*$/.test(id)) error(400, 'Invalid ID');
				await api(event, `${resourceName(resource(event))}/${id}`, { method: 'DELETE' });
				return { saved: true };
			} catch (problem) {
				return actionError(problem);
			}
		},
		language: async (event: RequestEvent) => {
			try {
				const form = await event.request.formData(),
					id = String(form.get('id')),
					language = String(form.get('language'));
				if (
					resource(event) !== 'deployments' ||
					!/^[1-9]\d*$/.test(id) ||
					!/^[A-Za-z]{1,8}(-[A-Za-z0-9]{1,8})*$/.test(language)
				)
					error(400, 'Invalid deployment or language');
				await api(event, `deployments/${id}/languages/${encodeURIComponent(language)}`, {
					method: form.get('remove') === 'true' ? 'DELETE' : 'PUT',
				});
				return { saved: true };
			} catch (problem) {
				return actionError(problem);
			}
		},
		target: async (event: RequestEvent) => {
			try {
				const form = await event.request.formData(),
					id = String(form.get('id')),
					deployment = String(form.get('deployment_id'));
				if (
					resource(event) !== 'policies' ||
					!/^[1-9]\d*$/.test(id) ||
					!/^[1-9]\d*$/.test(deployment)
				)
					error(400, 'Invalid policy or deployment');
				await api(event, `policies/${id}/targets/${deployment}`, {
					method: form.get('remove') === 'true' ? 'DELETE' : 'PUT',
					body:
						form.get('remove') === 'true'
							? undefined
							: JSON.stringify({ priority: Number(form.get('priority')) }),
				});
				return { saved: true };
			} catch (problem) {
				return actionError(problem);
			}
		},
	};
}
