<script lang="ts">
	import InfoTip from '#lib/components/InfoTip.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { usePreview } from '#lib/preview.svelte.js';
	import { requests } from '#lib/fixtures.js';
	const preview = usePreview();
	const bars = [
		16, 23, 19, 31, 25, 39, 32, 47, 43, 28, 34, 51, 45, 38, 54, 62, 48, 57, 65, 52, 41, 59, 67, 55,
		71, 62, 78, 68, 59, 74, 81, 72, 88, 74, 63, 78, 68, 85, 77, 91, 83, 72, 86, 78, 94, 87, 76, 89,
	];
</script>

<PageHeader
	title="Overview"
	path="overview"
	description="Your routing infrastructure, at a glance."
>
	{#snippet actions()}<Button variant="outline" size="sm" href="/quickstart">Quickstart ↗</Button
		><Button size="sm" href="/api-keys">Create API key +</Button>{/snippet}
</PageHeader>
<div class="onboarding">
	<div class="flex items-center gap-2">
		<h2>Configure your router</h2>
		<InfoTip
			text="Connect providers, choose deployments, and define fallback routes."
			label="Router setup"
		/>
	</div>
	<Button href="/quickstart" variant="outline" size="sm">Quickstart →</Button>
</div>
<div class="metric-grid">
	{#each [{ label: 'REQUESTS / 24H', value: '12,846', note: '+18.6% vs. previous day' }, { label: 'SUCCESS RATE', value: '99.82%', note: '23 failed requests' }, { label: 'MEDIAN FIRST BYTE', value: '184ms', note: 'Across all task types' }, { label: 'ACTIVE DEPLOYMENTS', value: '8 / 9', note: '1 needs attention' }] as metric, i}<div
			class="metric"
		>
			<div class="metric-label flex items-center gap-1">
				{metric.label}<InfoTip
					text={preview.sample ? metric.note : 'Waiting for your first request.'}
					label={`About ${metric.label.toLowerCase()}`}
				/>
			</div>
			<div class="metric-value">
				{preview.sample ? metric.value : i === 1 || i === 2 ? '—' : '0'}
			</div>
		</div>{/each}
</div>
<div class="two-column section">
	<section class="panel">
		<div class="panel-heading">
			<h2>Router traffic <span class="muted">/ 24h</span></h2>
			<a class="muted hover:underline" href="/activity">View activity ↗</a>
		</div>
		{#if preview.sample}<div class="chart-body">
				<div class="chart-scale"><span>800</span><span>400</span><span>0</span></div>
				<div
					class="chart"
					role="img"
					aria-label="Sample request volume increasing over 24 hours; synthetic data"
				>
					<div class="chart-lines"></div>
					<div class="bars">
						{#each bars as bar, i}<div class="bar-track">
								<div class="bar" style={`height:${bar}%`}></div>
								{#if i % 13 === 0}<div
										class="failure-bar"
										style={`height:${2 + (i % 4)}%`}
									></div>{/if}
							</div>{/each}
					</div>
				</div>
				<div class="chart-times">
					<span>00:00</span><span>06:00</span><span>12:00</span><span>18:00</span><span>23:59</span>
				</div>
			</div>
			<div class="chart-legend">
				<span><i class="legend-request"></i>Requests</span><span
					><i class="legend-failure"></i>Failures</span
				><span class="muted">Sample data</span>
			</div>{:else}<div class="p-3">
				<EmptyState
					title="No traffic yet"
					description="Request volume will appear after your first routed call."
					href="/quickstart"
					action="Get started"
				/>
			</div>{/if}
	</section>
	<section class="panel">
		<div class="panel-heading">
			<h2>Route health</h2>
			<span class={preview.sample ? 'positive' : 'muted'}
				>{preview.sample ? '● operational' : '○ not checked'}</span
			>
		</div>
		<div class="panel-body">
			<div class="health-list">
				<div>
					<span class="muted">Failure rate</span><span>{preview.sample ? '0.18%' : '—'}</span>
				</div>
				<div><span class="muted">Failovers</span><span>{preview.sample ? '14' : '0'}</span></div>
				<div>
					<span class="muted">Cloud connections</span><span>{preview.sample ? '6' : '0'}</span>
				</div>
				<div>
					<span class="muted">Local deployments</span><span>{preview.sample ? '3' : '0'}</span>
				</div>
				<div>
					<span class="muted">Last probe</span><span>{preview.sample ? '8s ago' : '—'}</span>
				</div>
			</div>
			<Button href="/logs" variant="outline" size="xs" class="mt-3">Inspect events →</Button>
		</div>
	</section>
</div>
<section class="section">
	<div class="section-heading">
		<h2>Recent requests <span class="muted">/ latest 5</span></h2>
		<Button href="/activity" variant="ghost" size="xs">View all →</Button>
	</div>
	{#if preview.sample}<div class="panel table-wrap">
			<table class="data-table">
				<thead
					><tr
						><th>Request ID</th><th>Task</th><th>Model</th><th>Status</th><th>Latency</th><th
							>Time</th
						></tr
					></thead
				><tbody
					>{#each requests.slice(0, 5) as request}<tr
							><td
								><a class="table-link" href={`/activity?request=${request.id}`}>{request.id}</a></td
							><td class="muted">{request.task.toUpperCase()}</td><td>{request.model}</td><td
								class={request.status === 'success' ? 'positive' : 'negative'}
								>{request.status === 'success' ? '✓' : '×'} {request.status}</td
							><td>{request.latency.toLocaleString()}ms</td><td class="muted">{request.time}</td
							></tr
						>{/each}</tbody
				>
			</table>
		</div>{:else}<EmptyState
			title="No requests recorded"
			description="Your request history will appear here."
			href="/quickstart"
			action="Read quickstart"
		/>{/if}
</section>
<div class="section panel">
	<div class="panel-heading">
		<h2>Deployment status</h2>
		<a href="/models" class="muted hover:underline">Explore models ↗</a>
	</div>
	<div class="deployment-grid">
		{#each [{ name: 'whisper-large-v3', address: 'local / speech to text', status: 'healthy', ms: '12ms' }, { name: 'qwen3:8b', address: 'local / language model', status: 'healthy', ms: '8ms' }, { name: 'kokoro-82m', address: 'local / text to speech', status: 'unknown', ms: '—' }] as deployment}<div
				class="deployment"
			>
				<div class="flex items-center justify-between gap-2">
					<span>{deployment.name}</span><span
						class={preview.sample && deployment.status === 'healthy' ? 'positive' : 'muted'}
						>{preview.sample ? deployment.status : 'not configured'}</span
					>
				</div>
				<div class="flex justify-between mt-1 muted">
					<span>{deployment.address}</span><span>{preview.sample ? deployment.ms : '—'}</span>
				</div>
			</div>{/each}
	</div>
</div>

<style>
	.chart-body {
		position: relative;
		padding: 20px 14px 0 42px;
	}
	.chart {
		height: 176px;
		position: relative;
		border-bottom: 1px solid var(--border);
	}
	.chart-lines {
		position: absolute;
		inset: 0;
		background: repeating-linear-gradient(
			to top,
			transparent 0,
			transparent calc(50% - 1px),
			var(--border) 50%
		);
		opacity: 0.7;
	}
	.bars {
		position: absolute;
		inset: 8px 0 0;
		display: flex;
		align-items: end;
		gap: 4px;
	}
	.bar-track {
		height: 100%;
		flex: 1;
		display: flex;
		align-items: end;
		position: relative;
	}
	.bar {
		width: 100%;
		background: var(--success);
		opacity: 0.65;
	}
	.bar-track:hover .bar {
		opacity: 1;
	}
	.failure-bar {
		position: absolute;
		bottom: 0;
		width: 100%;
		background: var(--destructive);
	}
	.chart-scale {
		position: absolute;
		left: 12px;
		top: 14px;
		bottom: 22px;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		color: var(--muted-foreground);
		font-size: 9px;
	}
	.chart-times {
		display: flex;
		justify-content: space-between;
		color: var(--muted-foreground);
		font-size: 9px;
		padding-top: 7px;
	}
	.chart-legend {
		padding: 12px;
		display: flex;
		gap: 16px;
		font-size: 10px;
	}
	.chart-legend span {
		display: flex;
		align-items: center;
		gap: 5px;
	}
	.chart-legend span:last-child {
		margin-left: auto;
	}
	.chart-legend i {
		width: 6px;
		height: 6px;
	}
	.legend-request {
		background: var(--success);
	}
	.legend-failure {
		background: var(--destructive);
	}
	.deployment-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}
	.deployment {
		padding: 12px;
		border-right: 1px solid var(--border);
		font-size: 10px;
	}
	.deployment:last-child {
		border-right: 0;
	}
	@media (max-width: 800px) {
		.deployment-grid {
			grid-template-columns: 1fr;
		}
		.deployment {
			border-right: 0;
			border-bottom: 1px solid var(--border);
		}
		.deployment:last-child {
			border-bottom: 0;
		}
		.bars {
			gap: 2px;
		}
	}
</style>
