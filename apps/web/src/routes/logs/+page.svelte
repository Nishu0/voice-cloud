<script lang="ts">
	import InfoTip from '#lib/components/InfoTip.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { events } from '#lib/fixtures.js';
	import { usePreview } from '#lib/preview.svelte.js';
	const preview = usePreview();
	let query = $state(''),
		level = $state('all'),
		source = $state('all');
	const filtered = $derived(
		(preview.sample ? events : []).filter(
			(event) =>
				(level === 'all' || event.level === level) &&
				(source === 'all' || event.source === source) &&
				`${event.message} ${event.detail}`.toLowerCase().includes(query.toLowerCase()),
		),
	);
</script>

<PageHeader
	title="Logs"
	path="observe / logs"
	description="System events, probe results, and routing diagnostics."
	>{#snippet actions()}<span class="eyebrow"
			>{preview.sample ? '7 SAMPLE EVENTS' : 'NO EVENTS'}</span
		>{/snippet}</PageHeader
>
<div class="filter-bar">
	<div class="filter-field grow">
		<label for="log-search">Search events</label><Input
			id="log-search"
			placeholder="Message or request ID..."
			bind:value={query}
		/>
	</div>
	<div class="filter-field">
		<label for="level">Severity</label><select id="level" class="native-select" bind:value={level}
			><option value="all">All levels</option
			>{#each ['info', 'warn', 'error', 'debug'] as value}<option>{value}</option>{/each}</select
		>
	</div>
	<div class="filter-field">
		<label for="source">Source</label><select id="source" class="native-select" bind:value={source}
			><option value="all">All sources</option
			>{#each ['router', 'worker', 'system'] as value}<option>{value}</option>{/each}</select
		>
	</div>
</div>
{#if filtered.length}<section class="panel">
		<div class="panel-heading">
			<span>EVENT STREAM <span class="muted">/ newest first</span></span><span
				class="flex items-center gap-1 muted"
				>UTC<InfoTip
					text="Sample events. Live log collection is not connected."
					label="About log data"
				/></span
			>
		</div>
		{#each filtered as event}<details class="event-row">
				<summary
					><span class="muted text-[10px]">{event.time}</span><span
						class="level"
						class:positive={event.level === 'info'}
						class:warning={event.level === 'warn'}
						class:negative={event.level === 'error'}
						class:muted={event.level === 'debug'}>{event.level}</span
					><span class="muted w-12">{event.source}</span><span class="event-message"
						>{event.message}</span
					></summary
				>
				<p>{event.detail}</p>
			</details>{/each}
		<div class="table-footer">
			<span>{filtered.length} events</span><span>Expand to inspect</span>
		</div>
	</section>{:else}<EmptyState
		title={preview.sample ? 'No matching events' : 'No events yet'}
		description={preview.sample
			? 'Try another search or clear your filters.'
			: 'Backend and worker events will appear here when connected.'}
	/>{/if}
{#if query || level !== 'all' || source !== 'all'}<Button
		variant="ghost"
		size="xs"
		class="mt-3"
		onclick={() => {
			query = '';
			level = 'all';
			source = 'all';
		}}>Clear filters ×</Button
	>{/if}
