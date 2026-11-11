<script lang="ts">
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Badge } from '#lib/components/ui/badge/index.js';
	import InfoTip from './InfoTip.svelte';
	import Copy from 'phosphor-svelte/lib/Copy';
	import { toast } from 'svelte-sonner';
	import { models, benchmarkAxes, benchmarkScores, type Model } from '#lib/fixtures.js';

	let { model, open = $bindable(false) }: { model: Model; open?: boolean } = $props();
	const axis = $derived(benchmarkAxes[model.task]);
	const score = $derived(benchmarkScores[model.id]);
	const peers = $derived(
		models
			.filter((peer) => peer.task === model.task)
			.sort((a, b) =>
				axis.lowerIsBetter
					? benchmarkScores[a.id] - benchmarkScores[b.id]
					: benchmarkScores[b.id] - benchmarkScores[a.id],
			),
	);
	const rank = $derived(peers.findIndex((peer) => peer.id === model.id) + 1);
	const maxScore = $derived(Math.max(...peers.map((peer) => benchmarkScores[peer.id])));
	const regions = $derived(model.location === 'Local' ? ['local'] : ['us-east-1', 'eu-central-1']);
	const price = $derived(model.price === 0 ? 'Self-hosted' : `$${model.price.toFixed(4)}`);
	async function copyId() {
		try {
			await navigator.clipboard.writeText(model.id);
			toast.success('Model ID copied');
		} catch {
			toast.error('Clipboard unavailable. Select and copy the model ID.');
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		class="model-dialog flex max-h-[90svh] flex-col gap-0 overflow-hidden p-0 sm:max-w-[760px]"
	>
		<Dialog.Header class="model-heading">
			<div class="flex min-w-0 items-center gap-3 pr-6">
				<span class="model-avatar">{model.mark}</span>
				<div class="min-w-0">
					<div class="eyebrow">{model.provider} / {model.task.toUpperCase()}</div>
					<Dialog.Title class="break-words">{model.id}</Dialog.Title>
				</div>
			</div>
			<Dialog.Description class="sr-only"
				>Model specifications and illustrative benchmark evidence.</Dialog.Description
			>
		</Dialog.Header>
		<div class="model-scroll">
			<div class="model-summary">
				<div>
					<div class="summary-label">
						{axis.label}<InfoTip text={axis.reading} label="About benchmark score" />
					</div>
					<strong>{score}<small>{axis.unit}</small></strong><span>English · sample</span>
				</div>
				<div>
					<div class="summary-label">
						Benchmark rank<InfoTip
							text="Rank among the sample models for this task, using the English benchmark score."
						/>
					</div>
					<strong>#{rank}</strong><span>of {peers.length} models</span>
				</div>
				<div>
					<div class="summary-label">
						Price<InfoTip
							text={model.price === 0
								? 'Provider fee excluded. Your infrastructure costs still apply.'
								: 'Illustrative price, not a current provider quote.'}
						/>
					</div>
					<strong class:local-price={model.price === 0}>{price}</strong><span
						>{model.price === 0 ? 'Infrastructure costs apply' : `/ ${model.unit}`}</span
					>
				</div>
				<div>
					<div class="summary-label">
						Deployment<InfoTip
							text="Example configuration only. No real provider connection or deployment has been created."
						/>
					</div>
					<Badge variant="outline">{model.location}</Badge><span
						>{regions.length} {model.location === 'Local' ? 'instance' : 'sample regions'}</span
					>
				</div>
			</div>
			<section class="model-specs">
				<div>
					<h2 class="flex items-center gap-1">
						Benchmark<InfoTip text={axis.reading} label="Benchmark reading" />
					</h2>
					<div class="benchmark-number">{score}<span>{axis.unit}</span></div>
					<div class="eyebrow">
						{axis.lowerIsBetter ? '↓ LOWER IS BETTER' : '↑ HIGHER IS BETTER'}
					</div>
					<div class="flex items-center gap-1 mt-3">
						<Badge variant="outline">Sample evidence</Badge><InfoTip
							text="Synthetic metrics for UI review. These are not measured benchmarks or provider claims."
							label="About sample evidence"
						/>
					</div>
				</div>
				<dl class="model-spec-list">
					<div>
						<dt>Model ID</dt>
						<dd class="select-all">{model.id}</dd>
					</div>
					<div>
						<dt>Provider</dt>
						<dd>{model.provider}</dd>
					</div>
					<div>
						<dt>Languages</dt>
						<dd>{model.languages.join(', ')}</dd>
					</div>
					<div>
						<dt>First byte</dt>
						<dd>{model.latency} ms</dd>
					</div>
					<div>
						<dt>Price</dt>
						<dd>{price}{model.price > 0 ? ` / ${model.unit}` : ''}</dd>
					</div>
					<div>
						<dt>Example regions</dt>
						<dd>{regions.join(', ')}</dd>
					</div>
					<div>
						<dt>Capabilities</dt>
						<dd>{model.capabilities.join(', ')}</dd>
					</div>
				</dl>
			</section>
			<section class="model-section">
				<div class="section-heading">
					<h2>Language metrics & evidence</h2>
					<span class="eyebrow">ENGLISH</span>
				</div>
				<dl class="evidence-list">
					<div>
						<dt>{axis.label}</dt>
						<dd>{score} {axis.unit}</dd>
					</div>
					<div>
						<dt>Evaluation samples</dt>
						<dd>{axis.samples.toLocaleString()}</dd>
					</div>
				</dl>
				<details class="axis-evidence">
					<summary>All axis evidence</summary>
					<div class="table-wrap">
						<table class="data-table">
							<thead><tr><th>Axis</th><th>Value</th><th>Source</th></tr></thead><tbody
								><tr><td>{axis.label}</td><td>{score} {axis.unit}</td><td>UI fixture</td></tr><tr
									><td>First byte</td><td>{model.latency} ms</td><td>UI fixture</td></tr
								><tr
									><td>Evaluation samples</td><td>{axis.samples.toLocaleString()}</td><td
										>UI fixture</td
									></tr
								></tbody
							>
						</table>
					</div>
				</details>
			</section>
			<section class="model-section">
				<div class="section-heading">
					<h2>Models on this axis</h2>
					<span class="eyebrow">{axis.label} {axis.lowerIsBetter ? '↓' : '↑'}</span>
				</div>
				<div class="comparison-list">
					{#each peers as peer}<div
							class="comparison-row"
							class:current-model={peer.id === model.id}
						>
							<span class="comparison-name"
								>{peer.provider.toLowerCase()} / {peer.id}{peer.id === model.id ? ' ←' : ''}</span
							>
							<div class="comparison-track" aria-hidden="true">
								<div style={`width:${(benchmarkScores[peer.id] / maxScore) * 100}%`}></div>
							</div>
							<span class="comparison-score">{benchmarkScores[peer.id]}</span>
						</div>{/each}
				</div>
			</section>
		</div>
		<Dialog.Footer class="model-footer"
			><span class="eyebrow mr-auto">SAMPLE DATA</span><Button
				variant="outline"
				size="sm"
				onclick={copyId}><Copy data-icon="inline-start" />Copy ID</Button
			><Button variant="outline" size="sm" onclick={() => (open = false)}>Close</Button
			></Dialog.Footer
		>
	</Dialog.Content>
</Dialog.Root>

<style>
	.model-heading {
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
		justify-content: center;
		gap: 4px;
		text-align: center;
		padding: 10px 6px;
		border-right: 1px solid var(--border);
	}
	.model-summary > div:last-child {
		border-right: 0;
	}
	.summary-label {
		display: flex;
		gap: 2px;
		align-items: center;
		font-size: 9px;
		color: var(--muted-foreground);
	}
	.model-summary strong {
		font-size: 22px;
		font-weight: 500;
		letter-spacing: -1px;
		line-height: 1.4;
	}
	.model-summary strong small {
		font-size: 10px;
		color: var(--muted-foreground);
		margin-left: 3px;
		letter-spacing: 0;
	}
	.model-summary strong.local-price {
		font-size: 15px;
		letter-spacing: -0.5px;
	}
	.model-summary > div > span {
		font-size: 9px;
		color: var(--muted-foreground);
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
		letter-spacing: -2px;
		line-height: 1.5;
		margin-top: 4px;
	}
	.benchmark-number > span {
		font-size: 14px;
		letter-spacing: 0;
		margin-left: 4px;
		color: var(--muted-foreground);
	}
	.model-spec-list > div {
		display: grid;
		grid-template-columns: 92px minmax(0, 1fr);
		gap: 8px;
		padding: 8px 0;
		border-bottom: 1px solid var(--border);
		font-size: 10px;
	}
	.model-spec-list > div:first-child {
		padding-top: 0;
	}
	.model-spec-list > div:last-child {
		border-bottom: 0;
	}
	.model-spec-list dt,
	.evidence-list dt {
		color: var(--muted-foreground);
	}
	.model-spec-list dd {
		text-align: right;
		overflow-wrap: anywhere;
	}
	.model-section {
		padding: 16px 0;
		border-bottom: 1px solid var(--border);
	}
	.model-section:last-child {
		padding-bottom: 0;
		border-bottom: 0;
	}
	.evidence-list {
		max-width: 380px;
	}
	.evidence-list > div {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		font-size: 11px;
		padding: 7px 0;
	}
	.axis-evidence {
		margin-top: 10px;
		font-size: 10px;
	}
	.axis-evidence summary {
		color: var(--muted-foreground);
		margin-bottom: 8px;
	}
	.comparison-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
		padding-top: 4px;
	}
	.comparison-row {
		display: grid;
		grid-template-columns: minmax(0, 1.1fr) minmax(40px, 1fr) 40px;
		align-items: center;
		gap: 12px;
		font-size: 10px;
		color: var(--muted-foreground);
	}
	.comparison-name {
		overflow-wrap: anywhere;
	}
	.comparison-score {
		text-align: right;
		font-variant-numeric: tabular-nums;
	}
	.comparison-track {
		height: 4px;
		background: var(--muted);
	}
	.comparison-track > div {
		height: 100%;
		background: var(--muted-foreground);
	}
	.current-model {
		color: var(--success);
	}
	.current-model .comparison-track > div {
		background: var(--success);
	}
	.model-footer {
		padding: 12px 16px;
		border-top: 1px solid var(--border);
		background: var(--card);
		align-items: center;
		flex-direction: row;
		gap: 6px;
	}
	@media (max-width: 600px) {
		.model-heading {
			padding: 12px;
		}
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
		.model-specs > div:first-child {
			display: none;
		}
		.comparison-row {
			grid-template-columns: minmax(0, 1.3fr) minmax(32px, 0.7fr) 32px;
			gap: 8px;
			font-size: 9px;
		}
		.model-footer {
			padding: 10px 12px;
		}
		.model-spec-list > div {
			grid-template-columns: 84px minmax(0, 1fr);
		}
	}
</style>
