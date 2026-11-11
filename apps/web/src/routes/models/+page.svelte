<script lang="ts">
	import InfoTip from '#lib/components/InfoTip.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Badge } from '#lib/components/ui/badge/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import * as Tabs from '#lib/components/ui/tabs/index.js';
	import ModelDetail from '#lib/components/ModelDetail.svelte';
	import { models, type Model } from '#lib/fixtures.js';
	import { usePreview } from '#lib/preview.svelte.js';
	const preview = usePreview();
	let query = $state(''),
		language = $state('all'),
		objective = $state('balanced'),
		task = $state('stt');
	let sort = $state('latency'),
		ascending = $state(true),
		selected = $state<Model | null>(null),
		detailOpen = $state(false);
	const tasks = [
		{ id: 'stt', name: 'Speech to text' },
		{ id: 'llm', name: 'Language models' },
		{ id: 'tts', name: 'Text to speech' },
		{ id: 's2s', name: 'Speech to speech' },
	];
	const filtered = $derived(
		(preview.sample ? models : [])
			.filter(
				(model) =>
					model.task === task &&
					`${model.id} ${model.provider}`.toLowerCase().includes(query.toLowerCase()) &&
					(language === 'all' || model.languages.includes(language)),
			)
			.sort((a, b) => {
				const result =
					sort === 'name'
						? a.id.localeCompare(b.id)
						: sort === 'price'
							? a.price - b.price
							: a.latency - b.latency;
				return ascending ? result : -result;
			}),
	);
	function order(column: string) {
		ascending = sort === column ? !ascending : true;
		sort = column;
	}
</script>

<PageHeader
	title="Models"
	path="connect / models"
	description="Explore cloud, custom, and local models."
	>{#snippet actions()}<Badge variant="outline"
			>{preview.sample ? models.length : 0} sample models</Badge
		>{/snippet}</PageHeader
>
<div class="filter-bar">
	<div class="filter-field grow">
		<label for="model-search">Search catalog</label><Input
			id="model-search"
			placeholder="Provider or model..."
			bind:value={query}
		/>
	</div>
	<div class="filter-field">
		<label for="language">Language</label><select
			id="language"
			class="native-select"
			bind:value={language}
			><option value="all">All languages</option
			>{#each ['English', 'Spanish', 'French', 'Hindi'] as lang}<option>{lang}</option
				>{/each}</select
		>
	</div>
	<div class="filter-field">
		<label for="objective">Objective</label><select
			id="objective"
			class="native-select"
			bind:value={objective}
			onchange={() => {
				sort = objective === 'cost' ? 'price' : 'latency';
				ascending = true;
			}}
			><option value="balanced">Balanced</option><option value="latency">Lowest latency</option
			><option value="cost">Lowest cost</option></select
		>
	</div>
</div>
<Tabs.Root bind:value={task}
	><div class="overflow-x-auto mb-3">
		<Tabs.List
			>{#each tasks as item}<Tabs.Trigger value={item.id}
					>{item.name}<span class="muted ml-1"
						>{preview.sample ? models.filter((m) => m.task === item.id).length : 0}</span
					></Tabs.Trigger
				>{/each}</Tabs.List
		>
	</div></Tabs.Root
>
{#if filtered.length}<div class="panel">
		<div class="table-wrap">
			<table class="data-table">
				<thead
					><tr
						><th>#</th><th
							><button onclick={() => order('name')}
								>Model {sort === 'name' ? (ascending ? '↑' : '↓') : '↕'}</button
							></th
						><th
							><button onclick={() => order('latency')}
								>First byte {sort === 'latency' ? (ascending ? '↑' : '↓') : '↕'}</button
							></th
						><th
							><button onclick={() => order('price')}
								>Price {sort === 'price' ? (ascending ? '↑' : '↓') : '↕'}</button
							></th
						><th>Capabilities</th><th>Location</th><th></th></tr
					></thead
				><tbody
					>{#each filtered as model, i}<tr
							><td class="muted">{String(i + 1).padStart(2, '0')}</td><td
								><button
									class="table-link model-name"
									onclick={() => {
										selected = model;
										detailOpen = true;
									}}
									><span class="provider-mark">{model.mark}</span><span
										>{model.id}<small>{model.provider}</small></span
									></button
								></td
							><td>{model.latency} ms</td><td
								>{model.price === 0 ? 'Self-hosted' : `$${model.price.toFixed(4)}`}
								<div class="muted text-[9px]">
									{model.price > 0 ? `/ ${model.unit}` : 'Infrastructure costs apply'}
								</div></td
							><td
								><div class="flex gap-1">
									{#each model.capabilities as capability}<Badge variant="outline"
											>{capability}</Badge
										>{/each}
								</div></td
							><td class="muted">{model.location}</td><td
								><Button
									size="icon-xs"
									variant="ghost"
									aria-label={`View ${model.id}`}
									onclick={() => {
										selected = model;
										detailOpen = true;
									}}>↗</Button
								></td
							></tr
						>{/each}</tbody
				>
			</table>
		</div>
		<div class="table-footer">
			<span>{filtered.length} models · {tasks.find((t) => t.id === task)?.name}</span><InfoTip
				text="Prices, latency, and capabilities are illustrative sample data."
				label="About model data"
			/>
		</div>
	</div>{:else}<EmptyState
		title={preview.sample ? 'No matching models' : 'Your catalog is empty'}
		description={preview.sample
			? 'Try a different search, language, or task.'
			: 'Models will appear after you configure your providers.'}
	/>{/if}

{#if selected}<ModelDetail model={selected} bind:open={detailOpen} />{/if}
