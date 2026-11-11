<script lang="ts">
	import type { PageProps } from './$types';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
	import InfoTip from '#lib/components/InfoTip.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	let { data }: PageProps = $props();
	const buckets = $derived(
		Array.from({ length: 24 }, (_, hour) => {
			const start = Math.floor(Date.now() / 3600000) * 3600000 - (23 - hour) * 3600000;
			return data.requests.filter((r) => r.started_at >= start && r.started_at < start + 3600000)
				.length;
		}),
	);
	const max = $derived(Math.max(1, ...buckets));
</script>

<PageHeader title="Overview" path="overview"
	>{#snippet actions()}<Button variant="outline" size="sm" href="/quickstart">Quickstart ↗</Button
		><Button size="sm" href="/manage/connections">Providers →</Button>{/snippet}</PageHeader
>
<div class="metric-grid">
	{#each [{ label: 'MODELS', value: data.models.length }, { label: 'ENABLED PROVIDERS', value: data.connections.filter((c) => c.enabled).length }, { label: 'ENABLED DEPLOYMENTS', value: data.deployments.filter((d) => d.enabled).length }, { label: 'RECENT REQUESTS', value: data.requests.length }] as metric}<div
			class="metric"
		>
			<div class="metric-label flex items-center gap-1">
				{metric.label}{#if metric.label === 'RECENT REQUESTS'}<InfoTip
						text="Latest 100 stored requests, not an installation-wide total."
					/>{/if}
			</div>
			<div class="metric-value">{metric.value}</div>
		</div>{/each}
</div>
<div class="two-column section">
	<section class="panel">
		<div class="panel-heading">
			<h2>Recent traffic / 24h</h2>
			<InfoTip
				text="Hourly counts within the latest 100 requests. Older records may not be included."
			/>
		</div>
		{#if data.requests.length}<div
				class="traffic-chart"
				role="img"
				aria-label="Hourly request counts from the latest 100 requests over 24 hours"
			>
				{#each buckets as count}<div class="traffic-track" title={`${count} requests`}>
						<div style={`height:${(count / max) * 100}%`}></div>
					</div>{/each}
			</div>
			<div class="table-footer"><span>24 hours ago</span><span>Now</span></div>{:else}<div
				class="p-3"
			>
				<EmptyState
					title="No traffic"
					description="Inference and activity recording are not implemented yet."
				/>
			</div>{/if}
	</section>
	<section class="panel">
		<div class="panel-heading">
			<h2>Deployment health</h2>
			<InfoTip text="Worker reachability probes, not model performance or inference readiness." />
		</div>
		<div class="panel-body health-list">
			{#each ['healthy', 'unhealthy', 'unknown'] as status}<div>
					<span class="muted">{status}</span><span
						>{Object.values(data.health).filter((h) => h.status === status).length}</span
					>
				</div>{/each}
			<div><a href="/manage/deployments" class="table-link">Inspect deployments →</a></div>
		</div>
	</section>
</div>
<section class="section">
	<div class="section-heading">
		<h2>Recent requests</h2>
		<Button href="/activity" variant="ghost" size="xs">View all →</Button>
	</div>
	{#if data.requests.length}<div class="panel table-wrap">
			<table class="data-table">
				<thead
					><tr
						><th>Request ID</th><th>Task</th><th>Status</th><th>Duration</th><th>Started / UTC</th
						></tr
					></thead
				><tbody
					>{#each data.requests.slice(0, 5) as request}<tr
							><td
								><a class="table-link" href={`/activity?request=${encodeURIComponent(request.id)}`}
									>{request.id}</a
								></td
							><td>{request.task.toUpperCase()}</td><td>{request.status}</td><td
								>{request.finished_at !== null
									? `${request.finished_at - request.started_at}ms`
									: '—'}</td
							><td class="muted">{new Date(request.started_at).toISOString()}</td></tr
						>{/each}</tbody
				>
			</table>
		</div>{:else}<EmptyState
			title="No requests"
			description="Requests will appear when activity recording is available."
		/>{/if}
</section>

<style>
	.traffic-chart {
		height: 210px;
		padding: 20px 12px 0;
		display: flex;
		align-items: end;
		gap: 6px;
	}
	.traffic-track {
		height: 100%;
		flex: 1;
		display: flex;
		align-items: end;
		border-bottom: 1px solid var(--border);
	}
	.traffic-track > div {
		width: 100%;
		min-height: 1px;
		background: var(--success);
		opacity: 0.7;
	}
</style>
