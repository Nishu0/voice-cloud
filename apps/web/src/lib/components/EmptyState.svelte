<script lang="ts">
	import * as Empty from '#lib/components/ui/empty/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import InfoTip from './InfoTip.svelte';
	import type { Snippet } from 'svelte';
	let {
		title,
		description,
		href,
		action,
		children,
	}: { title: string; description: string; href?: string; action?: string; children?: Snippet } =
		$props();
</script>

<Empty.Root class="min-h-52 border border-dashed">
	<Empty.Header
		><div class="empty-symbol" aria-hidden="true">[ ∅ ]</div>
		<div class="flex items-center gap-2">
			<Empty.Title>{title}</Empty.Title><InfoTip
				text={description}
				label={`About ${title.toLowerCase()}`}
			/>
		</div></Empty.Header
	>
	{#if children}<Empty.Content>{@render children()}</Empty.Content
		>{:else if href && action}<Empty.Content
			><Button size="sm" {href} variant="outline">{action} →</Button></Empty.Content
		>{/if}
</Empty.Root>
