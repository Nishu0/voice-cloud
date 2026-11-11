<script lang="ts">
	import type { PageProps } from './$types';
	import { goto } from '$app/navigation';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Sheet from '#lib/components/ui/sheet/index.js';
	let { data }: PageProps = $props();
	let query = $state(''),
		status = $state('all');
	const filtered = $derived(
		data.requests.filter(
			(r) =>
				(status === 'all' || r.status === status) &&
				r.id.toLowerCase().includes(query.toLowerCase()),
		),
	);
	function close() {
		goto(`/activity?offset=${data.offset}`, { reset: false });
	}
</script>

<PageHeader
	title="Activity"
	path="observe / activity"
	description="Real request history. Filters apply to the current page."
/>
<div class="filter-bar">
	<div class="filter-field grow">
		<label for="search">Request ID</label><Input
			id="search"
			placeholder="Search..."
			bind:value={query}
		/>
	</div>
	<div class="filter-field">
		<label for="status">Status</label><select class="native-select" id="status" bind:value={status}
			><option value="all">All</option
			>{#each ['pending', 'running', 'succeeded', 'failed', 'cancelled'] as status}<option
					>{status}</option
				>{/each}</select
		>
	</div>
</div>
{#if filtered.length}<div class="panel table-wrap">
		<table class="data-table">
			<thead
				><tr
					><th>Request ID</th><th>Task</th><th>Policy</th><th>Status</th><th>Duration</th><th
						>Started / UTC</th
					></tr
				></thead
			><tbody
				>{#each filtered as request}<tr
						><td
							><a
								class="table-link"
								href={`/activity?offset=${data.offset}&request=${encodeURIComponent(request.id)}`}
								>{request.id}</a
							></td
						><td>{request.task.toUpperCase()}</td><td
							>{data.policies.find((p) => p.id === request.policy_id)?.name ?? '—'}</td
						><td
							class={request.status === 'succeeded'
								? 'positive'
								: request.status === 'failed'
									? 'negative'
									: 'muted'}>{request.status}</td
						><td
							>{request.finished_at !== null
								? `${request.finished_at - request.started_at}ms`
								: '—'}</td
						><td class="muted">{new Date(request.started_at).toISOString()}</td></tr
					>{/each}</tbody
			>
		</table>
	</div>{:else}<EmptyState
		title="No requests"
		description="Inference and activity recording are not implemented yet. Existing records will appear here."
	/>{/if}
<div class="flex items-center justify-between mt-3">
	<span class="eyebrow"
		>PAGE {Math.floor(data.offset / 50) + 1} / {data.requests.length} RECORDS</span
	>
	<div class="actions">
		<Button
			href={`/activity?offset=${Math.max(0, data.offset - 50)}`}
			size="xs"
			variant="outline"
			disabled={data.offset === 0}>← Previous</Button
		><Button
			href={`/activity?offset=${data.offset + 50}`}
			size="xs"
			variant="outline"
			disabled={data.requests.length < 50}>Next →</Button
		>
	</div>
</div>
<Sheet.Root
	open={!!data.selected}
	onOpenChange={(value) => {
		if (!value) close();
	}}
	><Sheet.Content
		><Sheet.Header
			><Sheet.Title>{data.selected?.id}</Sheet.Title><Sheet.Description class="sr-only"
				>Request attempts and usage</Sheet.Description
			></Sheet.Header
		>{#if data.selected}<div class="sheet-body">
				<dl class="detail-list">
					<dt>Status</dt>
					<dd>{data.selected.status}</dd>
					<dt>Task</dt>
					<dd>{data.selected.task}</dd>
					<dt>Transport</dt>
					<dd>{data.selected.transport}</dd>
					<dt>Language</dt>
					<dd>{data.selected.language_tag ?? '—'}</dd>
					<dt>Error</dt>
					<dd>{data.selected.error_code ?? '—'}</dd>
					<dt>Trace ID</dt>
					<dd>{data.selected.trace_id ?? '—'}</dd>
				</dl>
				<h2 class="section mb-3">Attempts / {data.attempts.length}</h2>
				{#each data.attempts as attempt}<div class="panel panel-body mb-2">
						<span class={attempt.status === 'succeeded' ? 'positive' : 'muted'}
							>#{attempt.attempt_number} / {attempt.status}</span
						>
						<p>
							{data.deployments.find((d) => d.id === attempt.deployment_id)?.name ??
								`Deployment #${attempt.deployment_id}`}
						</p>
						<p class="muted text-[10px]">
							{attempt.selection_reason} / {attempt.error_code ?? '—'}
						</p>
						<p class="muted text-[10px]">
							First output: {attempt.first_output_at !== null
								? `${attempt.first_output_at - attempt.started_at}ms`
								: '—'}
						</p>
					</div>{:else}<p class="muted">No attempts</p>{/each}
				<h2 class="section mb-3">Usage</h2>
				{#each data.usage as usage}<div class="panel panel-body mb-2">
						<p>{usage.metric}: {usage.quantity} {usage.unit}</p>
						<p class="muted text-[10px]">
							{usage.cost_nanos === null
								? 'Cost unknown'
								: `$${(usage.cost_nanos / 1e9).toFixed(6)}`} / {usage.cost_status}
						</p>
					</div>{:else}<p class="muted">No usage recorded</p>{/each}
				<details class="section">
					<summary>Routing snapshot</summary>
					<pre class="overflow-x-auto text-[10px] mt-2">{JSON.stringify(
							data.selected.routing_snapshot_json,
							null,
							2,
						)}</pre>
				</details>
			</div>{/if}</Sheet.Content
	></Sheet.Root
>
