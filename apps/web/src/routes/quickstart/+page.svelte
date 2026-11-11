<script lang="ts">
	import InfoTip from '#lib/components/InfoTip.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import CodeBlock from '#lib/components/CodeBlock.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Tabs from '#lib/components/ui/tabs/index.js';
	let language = $state('curl');
	const examples: Record<string, string> = {
		curl: '# List the models configured in your installation\ncurl http://localhost:8080/v1/models \\\n  -H "Authorization: Bearer $SESSION_TOKEN"',
		python:
			'import json\nimport os\nfrom urllib.request import Request, urlopen\n\nrequest = Request(\n    "http://localhost:8080/v1/models",\n    headers={"Authorization": f"Bearer {os.environ[\'SESSION_TOKEN\']}"},\n)\nwith urlopen(request) as response:\n    print(json.load(response))',
		typescript:
			'// Run on the server; keep SESSION_TOKEN private\nconst response = await fetch(\n  "http://localhost:8080/v1/models",\n  { headers: { Authorization: `Bearer ${process.env.SESSION_TOKEN}` } }\n);\nif (!response.ok) throw new Error(`HTTP ${response.status}`);\nconst models = await response.json();\nconsole.log(models);',
	};
	const prompt =
		'# Voice Cloud — context for your coding agent\n\nVoice Cloud is a self-hosted voice model router.\nThe backend is Go, with SQLite and a maintenance worker.\nThe console uses Svelte, TypeScript, and shadcn-svelte.\n\nUse the management API at http://localhost:8080/v1.\nAuthenticate management requests with a user session token.\nKeep session and provider credentials on the server.\n\nConfigure connections, models, deployments, and routing policies.\nCustom and local deployments support explicit reachability probes.\nInference endpoints and SDK integrations are not implemented yet.';
</script>

<PageHeader
	title="Quickstart"
	path="connect / quickstart"
	description="From an empty installation to a configured router."
	>{#snippet actions()}<Button href="/api-keys" size="sm">Create API key +</Button
		>{/snippet}</PageHeader
>
<div class="two-column">
	<section>
		<h2 class="mb-4">Setup</h2>
		{#each [{ title: 'Start the backend', text: 'Run the API and maintenance worker with your SQLite database.' }, { title: 'Configure your providers', text: 'Add cloud connections, or point a custom connection at a local model.' }, { title: 'Define your routes', text: 'Choose deployments and fallback targets for each task.' }] as step, i}<div
				class="step"
			>
				<span class="step-number">0{i + 1}</span>
				<div class="flex items-center gap-1">
					<h3>{step.title}</h3>
					<InfoTip text={step.text} label={`About ${step.title.toLowerCase()}`} />
				</div>
			</div>{/each}
	</section>
	<CodeBlock
		title="terminal / start backend"
		code={'# From the repository root\n# Open the web console to create your first account\n# Development: API and worker in separate terminals\nmake api-run\nmake worker-run'}
	/>
</div>
<details class="section panel">
	<summary class="panel-heading">Coding agent context <span class="muted">+</span></summary
	><CodeBlock title="AGENT_CONTEXT.md" code={prompt} />
</details>
<section class="section">
	<div class="section-heading">
		<h2>Check your connection</h2>
		<InfoTip
			text="Server-side management API examples. Keep SESSION_TOKEN private. Inference and SDK integrations are not available yet."
			label="About these examples"
		/>
	</div>
	<Tabs.Root bind:value={language}
		><Tabs.List class="mb-2"
			><Tabs.Trigger value="curl">cURL</Tabs.Trigger><Tabs.Trigger value="python"
				>Python</Tabs.Trigger
			><Tabs.Trigger value="typescript">TypeScript</Tabs.Trigger></Tabs.List
		>{#each ['curl', 'python', 'typescript'] as lang}<Tabs.Content value={lang}
				><CodeBlock
					title={lang === 'curl'
						? 'terminal / list models'
						: lang === 'python'
							? 'connect.py'
							: 'connect.ts'}
					code={examples[lang]}
				/></Tabs.Content
			>{/each}</Tabs.Root
	>
</section>
