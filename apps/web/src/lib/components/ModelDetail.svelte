<script lang="ts">
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import InfoTip from './InfoTip.svelte';
	import EmptyState from './EmptyState.svelte';
	import { toast } from 'svelte-sonner';
	import type { Model, Configuration, Health, Language } from '#lib/api-types.js';
	let {
		model,
		configuration,
		health,
		languages,
		open = $bindable(false),
	}: {
		model: Model;
		configuration: Configuration;
		health: Record<number, Health>;
		languages: Record<number, Language[]>;
		open?: boolean;
	} = $props();
	const deployments = $derived(
		configuration.deployments.filter((deployment) => deployment.model_id === model.id),
	);
	const tags = $derived([
		...new Set(
			deployments.flatMap((deployment) =>
				(languages[deployment.id] ?? []).map((language) => language.language_tag),
			),
		),
	]);
	async function copyId() {
		try {
			await navigator.clipboard.writeText(String(model.id));
			toast.success('Model ID copied');
		} catch {
			toast.error('Clipboard unavailable');
		}
	}
</script>

<Dialog.Root bind:open
	><Dialog.Content
		class="model-dialog flex max-h-[90svh] flex-col gap-0 overflow-hidden p-0 sm:max-w-[760px]"
	>
		<Dialog.Header class="model-heading"
			><div class="flex items-center gap-3 pr-6">
				<span class="model-avatar">{model.publisher.slice(0, 2)}</span>
				<div class="min-w-0">
					<div class="eyebrow">{model.publisher} / {model.task.toUpperCase()}</div>
					<Dialog.Title class="break-words">{model.name}</Dialog.Title>
				</div>
			</div>
			<Dialog.Description class="sr-only"
				>Model configuration and deployment health</Dialog.Description
			></Dialog.Header
		>
		<div class="model-scroll">
			<div class="model-summary">
				{#each [{ label: 'Deployments', value: deployments.length }, { label: 'Enabled', value: deployments.filter((d) => d.enabled).length }, { label: 'Languages', value: tags.length }, { label: 'Healthy probes', value: deployments.filter((d) => health[d.id]?.status === 'healthy').length }] as metric}<div
					>
						<div class="summary-label">{metric.label}</div>
						<strong>{metric.value}</strong>
					</div>{/each}
			</div>
			<section class="model-specs">
				<div>
					<h2 class="flex items-center gap-1">
						Benchmarks<InfoTip
							text="The backend does not store benchmark scores or quality rankings yet."
						/>
					</h2>
					<div class="benchmark-number">—</div>
					<span class="eyebrow">NO EVIDENCE</span>
				</div>
				<dl class="model-spec-list">
					{#each [{ label: 'Model ID', value: model.id }, { label: 'Name', value: model.name }, { label: 'Publisher', value: model.publisher }, { label: 'Task', value: model.task.toUpperCase() }, { label: 'Version', value: model.version || '—' }, { label: 'Languages', value: tags.join(', ') || '—' }] as item}<div
						>
							<dt>{item.label}</dt>
							<dd>{item.value}</dd>
						</div>{/each}
				</dl>
			</section>
			<section class="model-section">
				<div class="section-heading">
					<h2>Deployments</h2>
					<InfoTip
						text="Probe latency measures endpoint reachability, not model inference latency."
					/>
				</div>
				{#each deployments as deployment}<div class="panel mb-3">
						<div class="panel-heading">
							<h2>{deployment.name}</h2>
							<span class={health[deployment.id]?.status === 'healthy' ? 'positive' : 'muted'}
								>{health[deployment.id]?.status ?? 'unknown'}</span
							>
						</div>
						<div class="panel-body">
							<dl class="detail-list">
								<dt>Provider</dt>
								<dd>
									{configuration.connections.find((c) => c.id === deployment.connection_id)?.name ??
										'—'}
								</dd>
								<dt>Upstream ID</dt>
								<dd>{deployment.upstream_model_id}</dd>
								<dt>Probe latency</dt>
								<dd>{health[deployment.id]?.latency_ms ?? '—'} ms</dd>
								<dt>Probe result</dt>
								<dd>{health[deployment.id]?.error_code ?? 'checked'}</dd>
								<dt>Enabled</dt>
								<dd>{deployment.enabled ? 'yes' : 'no'}</dd>
							</dl>
							{#each ['capabilities_json', 'pricing_json', 'settings_json'] as key}<details
									class="axis-evidence"
								>
									<summary>{key.replace('_json', '')}</summary>
									<pre class="overflow-x-auto text-[10px]">{JSON.stringify(
											deployment[key as 'capabilities_json' | 'pricing_json' | 'settings_json'],
											null,
											2,
										)}</pre>
								</details>{/each}
						</div>
					</div>{:else}<EmptyState
						title="No deployments"
						description="Create a deployment for this model."
						href="/manage/deployments"
						action="Deployments"
					/>{/each}
			</section>
		</div>
		<Dialog.Footer class="model-footer"
			><span class="eyebrow mr-auto">MODEL #{model.id}</span><Button
				variant="outline"
				size="sm"
				onclick={copyId}>Copy ID</Button
			><Button variant="outline" size="sm" onclick={() => (open = false)}>Close</Button
			></Dialog.Footer
		></Dialog.Content
	></Dialog.Root
>

<style>
	:global(.model-heading) {
		padding: 16px;
		border-bottom: 1px solid var(--border);
	}
	.model-avatar {
		display: grid;
		place-items: center;
		width: 42px;
		height: 42px;
		flex-shrink: 0;
		border: 1px solid var(--border);
		background: var(--muted);
		color: var(--success);
		font-size: 16px;
	}
	.model-scroll {
		overflow-y: auto;
		min-height: 0;
		padding: 16px;
		scrollbar-width: thin;
		scrollbar-color: var(--border) transparent;
	}
	.model-summary {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		border: 1px solid var(--border);
	}
	.model-summary > div {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		padding: 10px 6px;
		border-right: 1px solid var(--border);
	}
	.model-summary > div:last-child {
		border-right: 0;
	}
	.summary-label {
		font-size: 9px;
		color: var(--muted-foreground);
	}
	.model-summary strong {
		font-size: 22px;
		font-weight: 500;
		letter-spacing: -1px;
	}
	.model-specs {
		display: grid;
		grid-template-columns: minmax(0, 0.85fr) minmax(0, 1.4fr);
		gap: 24px;
		padding: 20px 0;
		border-bottom: 1px solid var(--border);
	}
	.benchmark-number {
		font-size: 36px;
		margin-top: 4px;
	}
	.model-spec-list > div {
		display: grid;
		grid-template-columns: 92px minmax(0, 1fr);
		gap: 8px;
		padding: 8px 0;
		border-bottom: 1px solid var(--border);
		font-size: 10px;
	}
	.model-spec-list dt {
		color: var(--muted-foreground);
	}
	.model-spec-list dd {
		text-align: right;
		overflow-wrap: anywhere;
	}
	.model-section {
		padding-top: 16px;
	}
	.axis-evidence {
		margin-top: 10px;
		font-size: 10px;
	}
	.axis-evidence summary {
		color: var(--muted-foreground);
		margin-bottom: 8px;
	}
	:global(.model-footer) {
		padding: 12px 16px;
		border-top: 1px solid var(--border);
		background: var(--card);
		align-items: center;
		flex-direction: row;
		gap: 6px;
	}
	@media (max-width: 600px) {
		.model-scroll {
			padding: 12px;
		}
		.model-summary {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
		.model-summary > div:nth-child(2) {
			border-right: 0;
		}
		.model-summary > div:nth-child(-n + 2) {
			border-bottom: 1px solid var(--border);
		}
		.model-specs {
			grid-template-columns: 1fr;
			gap: 16px;
			padding: 14px 0;
		}
	}
</style>
