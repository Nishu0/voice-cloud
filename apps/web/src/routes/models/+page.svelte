<script lang="ts">
	import type { PageProps } from './$types';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import ConfigurationTable from '#lib/components/ConfigurationTable.svelte';
	import ModelDetail from '#lib/components/ModelDetail.svelte';
	import * as Tabs from '#lib/components/ui/tabs/index.js';
	let { data }: PageProps = $props();
	let task = $state('all'),
		selectedId = $state<number | null>(null),
		open = $state(false);
	const selected = $derived(data.models.find((model) => model.id === selectedId));
</script>

<PageHeader title="Models" path="connect / models" description="Your configured model catalog." />
<Tabs.Root bind:value={task}
	><Tabs.List class="mb-3"
		>{#each [{ id: 'all', label: 'All' }, { id: 'stt', label: 'Speech to text' }, { id: 'llm', label: 'Language models' }, { id: 'tts', label: 'Text to speech' }] as item}<Tabs.Trigger
				value={item.id}>{item.label}</Tabs.Trigger
			>{/each}</Tabs.List
	></Tabs.Root
>
<ConfigurationTable
	resource="models"
	configuration={data}
	records={data.models.filter((model) => task === 'all' || model.task === task)}
	onview={(id) => {
		selectedId = id;
		open = true;
	}}
/>
{#if selected}<ModelDetail
		model={selected}
		configuration={data}
		health={data.health}
		languages={data.languages}
		bind:open
	/>{/if}
