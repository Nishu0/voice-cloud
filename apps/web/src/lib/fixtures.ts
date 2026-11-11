export const navigation = [
	{
		label: 'Workspace',
		items: [{ title: 'Overview', href: '/', icon: 'overview', shortcut: '01' }],
	},
	{
		label: 'Connect',
		items: [
			{
				title: 'Quickstart',
				href: '/quickstart',
				icon: 'code',
				shortcut: '02',
			},
			{ title: 'Models', href: '/models', icon: 'models', shortcut: '03' },
			{ title: 'API keys', href: '/api-keys', icon: 'key', shortcut: '04' },
		],
	},
	{
		label: 'Observe',
		items: [
			{ title: 'Logs', href: '/logs', icon: 'logs', shortcut: '05' },
			{
				title: 'Activity',
				href: '/activity',
				icon: 'activity',
				shortcut: '06',
			},
		],
	},
	{
		label: 'Platform',
		items: [
			{
				title: 'Settings',
				href: '/settings/account',
				icon: 'settings',
				shortcut: '07',
			},
		],
	},
];

export type Model = {
	id: string;
	provider: string;
	mark: string;
	task: string;
	latency: number;
	price: number;
	unit: string;
	capabilities: string[];
	languages: string[];
	location: string;
};
export const models: Model[] = [
	{
		id: 'nova-3',
		provider: 'Deepgram',
		mark: 'dg',
		task: 'stt',
		latency: 278,
		price: 0.0077,
		unit: 'min',
		capabilities: ['Streaming', 'Diarization'],
		languages: ['English', 'Spanish', 'French'],
		location: 'Cloud',
	},
	{
		id: 'universal-2',
		provider: 'AssemblyAI',
		mark: 'aa',
		task: 'stt',
		latency: 589,
		price: 0.0062,
		unit: 'min',
		capabilities: ['Diarization', 'Word timestamps'],
		languages: ['English', 'Spanish'],
		location: 'Cloud',
	},
	{
		id: 'whisper-large-v3',
		provider: 'Local',
		mark: '~/',
		task: 'stt',
		latency: 412,
		price: 0,
		unit: 'min',
		capabilities: ['Self-hosted', 'Multilingual'],
		languages: ['English', 'Spanish', 'French', 'Hindi'],
		location: 'Local',
	},
	{
		id: 'gpt-4o-mini-transcribe',
		provider: 'OpenAI',
		mark: 'oa',
		task: 'stt',
		latency: 340,
		price: 0.003,
		unit: 'min',
		capabilities: ['Multilingual'],
		languages: ['English', 'Spanish', 'French', 'Hindi'],
		location: 'Cloud',
	},
	{
		id: 'chirp-3',
		provider: 'Google',
		mark: 'gg',
		task: 'stt',
		latency: 380,
		price: 0.016,
		unit: 'min',
		capabilities: ['Streaming', 'Diarization'],
		languages: ['English', 'Hindi'],
		location: 'Cloud',
	},
	{
		id: 'gpt-4o-mini',
		provider: 'OpenAI',
		mark: 'oa',
		task: 'llm',
		latency: 186,
		price: 0.15,
		unit: '1M input tokens',
		capabilities: ['Tools', 'Streaming'],
		languages: ['English', 'Spanish', 'French', 'Hindi'],
		location: 'Cloud',
	},
	{
		id: 'llama-3.3-70b',
		provider: 'Groq',
		mark: 'gq',
		task: 'llm',
		latency: 124,
		price: 0.59,
		unit: '1M input tokens',
		capabilities: ['Tools', 'Streaming'],
		languages: ['English', 'Spanish'],
		location: 'Cloud',
	},
	{
		id: 'qwen3:8b',
		provider: 'Local',
		mark: '~/',
		task: 'llm',
		latency: 230,
		price: 0,
		unit: '1M input tokens',
		capabilities: ['Self-hosted', 'Tools'],
		languages: ['English', 'Hindi'],
		location: 'Local',
	},
	{
		id: 'sonic-2',
		provider: 'Cartesia',
		mark: 'ca',
		task: 'tts',
		latency: 90,
		price: 50,
		unit: '1M characters',
		capabilities: ['Streaming', 'Voice cloning'],
		languages: ['English', 'Spanish', 'French'],
		location: 'Cloud',
	},
	{
		id: 'eleven-flash-v2.5',
		provider: 'ElevenLabs',
		mark: 'el',
		task: 'tts',
		latency: 75,
		price: 75,
		unit: '1M characters',
		capabilities: ['Streaming', 'Voice cloning'],
		languages: ['English', 'Spanish', 'Hindi'],
		location: 'Cloud',
	},
	{
		id: 'kokoro-82m',
		provider: 'Local',
		mark: '~/',
		task: 'tts',
		latency: 110,
		price: 0,
		unit: '1M characters',
		capabilities: ['Self-hosted'],
		languages: ['English'],
		location: 'Local',
	},
	{
		id: 'gpt-4o-realtime',
		provider: 'OpenAI',
		mark: 'oa',
		task: 's2s',
		latency: 296,
		price: 40,
		unit: '1M audio input tokens',
		capabilities: ['Realtime', 'Tools'],
		languages: ['English', 'Spanish', 'French'],
		location: 'Cloud',
	},
];

// Synthetic benchmark fixtures for the UI; replace with measured evidence when connected.
export const benchmarkAxes: Record<
	string,
	{
		label: string;
		unit: string;
		lowerIsBetter: boolean;
		samples: number;
		reading: string;
	}
> = {
	stt: {
		label: 'Word error rate',
		unit: '%',
		lowerIsBetter: true,
		samples: 1541,
		reading: 'Transcription errors as a percentage of words. Lower is better.',
	},
	llm: {
		label: 'Task accuracy',
		unit: '%',
		lowerIsBetter: false,
		samples: 500,
		reading:
			'Correct responses on the sample evaluation set. Higher is better.',
	},
	tts: {
		label: 'Voice quality',
		unit: '/ 5',
		lowerIsBetter: false,
		samples: 240,
		reading:
			'Mean opinion score on the sample listening set. Higher is better.',
	},
	s2s: {
		label: 'Response latency',
		unit: 'ms',
		lowerIsBetter: true,
		samples: 320,
		reading: 'Time to the first audio response. Lower is better.',
	},
};
export const benchmarkScores: Record<string, number> = {
	'nova-3': 4.94,
	'universal-2': 5.21,
	'whisper-large-v3': 6.18,
	'gpt-4o-mini-transcribe': 5.08,
	'chirp-3': 5.74,
	'gpt-4o-mini': 86.4,
	'llama-3.3-70b': 83.7,
	'qwen3:8b': 78.2,
	'sonic-2': 4.32,
	'eleven-flash-v2.5': 4.41,
	'kokoro-82m': 4.12,
	'gpt-4o-realtime': 296,
};

export const requests = [
	{
		id: 'req_8f2a91',
		time: '14:32:08',
		task: 'stt',
		model: 'nova-3',
		provider: 'Deepgram',
		status: 'success',
		latency: 278,
		usage: '12.4s audio',
		fallback: false,
	},
	{
		id: 'req_8f2a90',
		time: '14:32:06',
		task: 'llm',
		model: 'gpt-4o-mini',
		provider: 'OpenAI',
		status: 'success',
		latency: 186,
		usage: '428 tokens',
		fallback: false,
	},
	{
		id: 'req_8f2a89',
		time: '14:32:05',
		task: 'tts',
		model: 'sonic-2',
		provider: 'Cartesia',
		status: 'success',
		latency: 94,
		usage: '186 characters',
		fallback: true,
	},
	{
		id: 'req_8f2a88',
		time: '14:31:58',
		task: 'stt',
		model: 'whisper-large-v3',
		provider: 'Local',
		status: 'success',
		latency: 412,
		usage: '8.1s audio',
		fallback: false,
	},
	{
		id: 'req_8f2a87',
		time: '14:31:54',
		task: 'tts',
		model: 'eleven-flash-v2.5',
		provider: 'ElevenLabs',
		status: 'error',
		latency: 5000,
		usage: '—',
		fallback: false,
	},
	{
		id: 'req_8f2a86',
		time: '14:31:49',
		task: 'llm',
		model: 'llama-3.3-70b',
		provider: 'Groq',
		status: 'success',
		latency: 124,
		usage: '312 tokens',
		fallback: false,
	},
];
export type RequestRecord = (typeof requests)[number];

export const events = [
	{
		id: 'evt_001',
		time: '14:32:08.284',
		level: 'info',
		source: 'router',
		message: 'Request completed',
		detail: 'req_8f2a91 → deepgram/nova-3 · 278ms · policy: balanced',
	},
	{
		id: 'evt_002',
		time: '14:32:05.119',
		level: 'warn',
		source: 'router',
		message: 'Fallback target selected',
		detail:
			'Primary target returned HTTP 429. Selected cartesia/sonic-2 on attempt 2.',
	},
	{
		id: 'evt_003',
		time: '14:32:00.041',
		level: 'info',
		source: 'worker',
		message: 'Deployment probe completed',
		detail:
			'local/whisper-large-v3 · reachability: healthy · latency: 12ms · expires in 60s',
	},
	{
		id: 'evt_004',
		time: '14:31:59.002',
		level: 'debug',
		source: 'worker',
		message: 'Expired credentials cleaned up',
		detail: 'Removed 2 expired sessions and 0 expired one-use tokens.',
	},
	{
		id: 'evt_005',
		time: '14:31:54.302',
		level: 'error',
		source: 'router',
		message: 'Upstream request timed out',
		detail:
			'req_8f2a87 · elevenlabs/eleven-flash-v2.5 · timeout: 5000ms · no fallback configured',
	},
	{
		id: 'evt_006',
		time: '14:31:30.084',
		level: 'info',
		source: 'worker',
		message: 'Deployment probe completed',
		detail:
			'local/qwen3:8b · reachability: healthy · latency: 8ms · expires in 60s',
	},
	{
		id: 'evt_007',
		time: '14:30:00.001',
		level: 'info',
		source: 'system',
		message: 'Backend started',
		detail:
			'SQLite connected · migrations applied · maintenance worker started',
	},
];
