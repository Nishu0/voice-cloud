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
			{ title: 'Providers', href: '/manage/connections', icon: 'models', shortcut: '' },
			{ title: 'Deployments', href: '/manage/deployments', icon: 'activity', shortcut: '' },
			{ title: 'Routing policies', href: '/manage/policies', icon: 'code', shortcut: '' },
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
