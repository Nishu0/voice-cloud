export type ConfigField = {
	key: string;
	label: string;
	kind?: 'json' | 'boolean' | 'number' | 'select';
	options?: string[];
	immutable?: boolean;
	optional?: boolean;
	initial?: string;
	help?: string;
};
export const configFields: Record<string, ConfigField[]> = {
	connections: [
		{ key: 'name', label: 'Name' },
		{
			key: 'adapter',
			label: 'Adapter',
			help: 'Provider adapter name, or custom for your own HTTP endpoint.',
		},
		{ key: 'location_kind', label: 'Location', kind: 'select', options: ['cloud', 'local'] },
		{ key: 'base_url', label: 'Base URL' },
		{
			key: 'credential_ref',
			label: 'Credential reference',
			optional: true,
			help: 'Name of an external secret, not the secret itself. Resolution is not implemented yet.',
		},
		{ key: 'enabled', label: 'Enabled', kind: 'boolean', initial: 'true' },
	],
	models: [
		{ key: 'name', label: 'Name' },
		{ key: 'publisher', label: 'Publisher' },
		{ key: 'task', label: 'Task', kind: 'select', options: ['stt', 'tts', 'llm'], immutable: true },
		{ key: 'version', label: 'Version', optional: true },
	],
	deployments: [
		{ key: 'name', label: 'Name' },
		{ key: 'model_id', label: 'Model', kind: 'select', immutable: true },
		{ key: 'connection_id', label: 'Provider connection', kind: 'select', immutable: true },
		{ key: 'upstream_model_id', label: 'Upstream model ID', immutable: true },
		{ key: 'enabled', label: 'Enabled', kind: 'boolean', initial: 'true' },
		{ key: 'capabilities_json', label: 'Capabilities', kind: 'json' },
		{ key: 'pricing_json', label: 'Pricing', kind: 'json' },
		{
			key: 'settings_json',
			label: 'Settings',
			kind: 'json',
			help: 'For custom/local probes, use {"healthcheck":{"url":"http://your-host/health"}}. The URL must match the connection origin.',
		},
	],
	policies: [
		{ key: 'name', label: 'Name' },
		{ key: 'task', label: 'Task', kind: 'select', options: ['stt', 'tts', 'llm'], immutable: true },
		{
			key: 'objective',
			label: 'Objective',
			kind: 'select',
			options: ['balanced', 'latency', 'cost', 'quality'],
		},
		{ key: 'fallback_enabled', label: 'Enable fallback', kind: 'boolean', initial: 'true' },
		{ key: 'max_attempts', label: 'Maximum attempts', kind: 'number', initial: '2' },
		{ key: 'constraints_json', label: 'Constraints', kind: 'json' },
	],
};
