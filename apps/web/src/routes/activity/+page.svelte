<script lang="ts">
	import InfoTip from '#lib/components/InfoTip.svelte';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Sheet from '#lib/components/ui/sheet/index.js';
	import { requests, type RequestRecord } from '#lib/fixtures.js';
	import { usePreview } from '#lib/preview.svelte.js';
	const preview = usePreview();
	let query = $state(''),
		task = $state('all'),
		status = $state('all'),
		detailOpen = $state(false),
		selected = $state<RequestRecord | null>(null);
	let ascending = $state(false);
	const filtered = $derived(
		(preview.sample ? requests : [])
			.filter(
				(request) =>
					(task === 'all' || request.task === task) &&
					(status === 'all' || request.status === status) &&
					`${request.id} ${request.model} ${request.provider}`
						.toLowerCase()
						.includes(query.toLowerCase()),
			)
			.sort((a, b) => (ascending ? a.time.localeCompare(b.time) : b.time.localeCompare(a.time))),
	);
	onMount(() => {
		const found = requests.find((r) => r.id === page.url.searchParams.get('request'));
		if (found && preview.sample) {
			selected = found;
			detailOpen = true;
		}
	});
</script>

<PageHeader
	title="Activity"
	path="observe / activity"
	description="Trace each request from route selection to response."
	>{#snippet actions()}<InfoTip
			label="About request data"
			text="Synthetic request metadata. No audio, transcripts, or prompts are recorded in this preview."
		/>{/snippet}</PageHeader
>
<div class="filter-bar">
	<div class="filter-field grow">
		<label for="request-search">Search requests</label><Input
			id="request-search"
			placeholder="Request ID, model, or provider..."
			bind:value={query}
		/>
	</div>
	<div class="filter-field">
		<label for="task">Task</label><select id="task" class="native-select" bind:value={task}
			><option value="all">All tasks</option>{#each ['stt', 'llm', 'tts'] as value}<option {value}
					>{value.toUpperCase()}</option
				>{/each}</select
		>
	</div>
	<div class="filter-field">
		<label for="status">Status</label><select id="status" class="native-select" bind:value={status}
			><option value="all">All statuses</option><option value="success">Success</option><option
				value="error">Error</option
			></select
		>
	</div>
</div>
{#if filtered.length}<div class="panel">
		<div class="table-wrap">
			<table class="data-table">
				<thead
					><tr
						><th>Request ID</th><th>Task</th><th>Model / provider</th><th>Status</th><th>Latency</th
						><th
							><button onclick={() => (ascending = !ascending)}>Time {ascending ? '↑' : '↓'}</button
							></th
						><th></th></tr
					></thead
				><tbody
					>{#each filtered as request}<tr
							><td
								><button
									class="table-link"
									onclick={() => {
										selected = request;
										detailOpen = true;
									}}>{request.id}</button
								></td
							><td class="muted">{request.task.toUpperCase()}</td><td
								>{request.model}
								<div class="muted text-[9px]">
									{request.provider}{request.fallback ? ' / fallback' : ''}
								</div></td
							><td class={request.status === 'success' ? 'positive' : 'negative'}
								>{request.status === 'success' ? '✓' : '×'} {request.status}</td
							><td>{request.latency.toLocaleString()}ms</td><td class="muted">{request.time}</td><td
								><Button
									size="icon-xs"
									variant="ghost"
									aria-label={`Inspect ${request.id}`}
									onclick={() => {
										selected = request;
										detailOpen = true;
									}}>↗</Button
								></td
							></tr
						>{/each}</tbody
				>
			</table>
		</div>
		<div class="table-footer">
			<span>{filtered.length} requests · UTC</span><span>Sample data</span>
		</div>
	</div>{:else}<EmptyState
		title={preview.sample ? 'No matching requests' : 'No router requests yet'}
		description={preview.sample
			? 'Try a different search or filter.'
			: 'Requests appear here once your first call reaches the router.'}
		href={preview.sample ? undefined : '/quickstart'}
		action="Open quickstart"
	/>{/if}

<Sheet.Root bind:open={detailOpen}
	><Sheet.Content
		><Sheet.Header
			><Sheet.Title>{selected?.id}</Sheet.Title><Sheet.Description class="sr-only"
				>Sample request trace · UTC</Sheet.Description
			></Sheet.Header
		>{#if selected}<div class="sheet-body">
				<dl class="detail-list">
					<dt>Status</dt>
					<dd class={selected.status === 'success' ? 'positive' : 'negative'}>{selected.status}</dd>
					<dt>Task</dt>
					<dd>{selected.task.toUpperCase()}</dd>
					<dt>Model</dt>
					<dd>{selected.model}</dd>
					<dt>Provider</dt>
					<dd>{selected.provider}</dd>
					<dt>Latency</dt>
					<dd>{selected.latency}ms</dd>
					<dt>Usage</dt>
					<dd>{selected.usage}</dd>
					<dt>Time</dt>
					<dd>{selected.time}</dd>
				</dl>
				<div class="section">
					<h2 class="mb-3">Attempts <span class="muted">/ {selected.fallback ? 2 : 1}</span></h2>
					{#if selected.fallback}<div class="attempt">
							<span class="warning">01 / rate limited</span>
							<p class="muted">Primary TTS deployment · HTTP 429</p>
						</div>{/if}
					<div class="attempt">
						<span class={selected.status === 'success' ? 'positive' : 'negative'}
							>{selected.fallback ? '02' : '01'} / {selected.status === 'success'
								? 'completed'
								: 'timeout'}</span
						>
						<p class="muted">{selected.provider} / {selected.model}</p>
						<span class="muted">{selected.latency}ms</span>
					</div>
				</div>
			</div>{/if}</Sheet.Content
	></Sheet.Root
>

<style>
	.attempt {
		border-left: 1px solid var(--border);
		padding: 0 0 16px 12px;
		font-size: 11px;
	}
	.attempt:last-child {
		padding-bottom: 0;
	}
</style>
