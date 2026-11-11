export type User = {
	id: number;
	email: string;
	name: string;
	role: 'admin' | 'member';
	status: string;
	created_at: number;
};
export type Connection = {
	id: number;
	name: string;
	adapter: string;
	location_kind: 'local' | 'cloud';
	base_url: string;
	credential_ref: string | null;
	enabled: boolean;
};
export type Model = {
	id: number;
	name: string;
	publisher: string;
	task: 'stt' | 'tts' | 'llm';
	version: string;
};
export type Deployment = {
	id: number;
	name: string;
	model_id: number;
	connection_id: number;
	upstream_model_id: string;
	enabled: boolean;
	capabilities_json: Record<string, unknown>;
	pricing_json: Record<string, unknown>;
	settings_json: Record<string, unknown>;
};
export type Health = {
	deployment_id: number;
	status: 'healthy' | 'unhealthy' | 'unknown';
	checked_at: number | null;
	expires_at: number | null;
	latency_ms: number | null;
	error_code: string | null;
};
export type Policy = {
	id: number;
	name: string;
	task: Model['task'];
	objective: string;
	fallback_enabled: boolean;
	max_attempts: number;
	constraints_json: Record<string, unknown>;
};
export type APIKey = {
	id: number;
	name: string;
	key_prefix: string;
	default_policy_id: number | null;
	expires_at: number | null;
	revoked_at: number | null;
	created_at: number;
	created_by_user_id: number | null;
};
export type RequestRecord = {
	id: string;
	api_key_id: number;
	task: string;
	transport: string;
	language_tag: string | null;
	policy_id: number | null;
	status: string;
	started_at: number;
	finished_at: number | null;
	error_code: string | null;
	trace_id: string | null;
	routing_snapshot_json: Record<string, unknown>;
};
export type Attempt = {
	id: string;
	request_id: string;
	deployment_id: number;
	attempt_number: number;
	status: string;
	started_at: number;
	first_output_at: number | null;
	finished_at: number | null;
	error_code: string | null;
	selection_reason: string;
};
export type Usage = {
	attempt_id: string;
	metric: string;
	quantity: number;
	unit: string;
	cost_nanos: number | null;
	currency: string;
	cost_status: string;
};
export type Configuration = {
	connections: Connection[];
	models: Model[];
	deployments: Deployment[];
	policies: Policy[];
};
export type Target = { deployment_id: number; policy_id: number; priority: number };
export type Language = { deployment_id: number; language_tag: string };
