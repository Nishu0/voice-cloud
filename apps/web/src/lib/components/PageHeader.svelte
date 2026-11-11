<script lang="ts">
	import type { Snippet } from 'svelte';
	import InfoTip from './InfoTip.svelte';
	let {
		title,
		path,
		description,
		actions,
	}: { title: string; path: string; description?: string; actions?: Snippet } = $props();
</script>

<svelte:head><title>{title} · Voice Cloud</title></svelte:head>
<header class="page-header">
	<div>
		<div class="eyebrow">~/ {path}</div>
		<div class="flex items-center gap-2">
			<h1>{title}<span class="cursor">_</span></h1>
			{#if description}<InfoTip text={description} label={`About ${title}`} />{/if}
		</div>
	</div>
	{#if actions}<div class="actions">{@render actions()}</div>{/if}
</header>
