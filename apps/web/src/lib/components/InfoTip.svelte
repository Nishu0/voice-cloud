<script lang="ts">
	import * as Tooltip from '#lib/components/ui/tooltip/index.js';
	import Info from 'phosphor-svelte/lib/Info';

	let { text, label = 'More information' }: { text: string; label?: string } = $props();
	let open = $state(false);
</script>

<Tooltip.Root bind:open>
	<Tooltip.Trigger
		type="button"
		class="info-trigger"
		aria-label={label}
		onclick={() => (open = !open)}
	>
		<Info size={14} />
	</Tooltip.Trigger>
	<Tooltip.Content sideOffset={6}><p>{text}</p></Tooltip.Content>
</Tooltip.Root>
