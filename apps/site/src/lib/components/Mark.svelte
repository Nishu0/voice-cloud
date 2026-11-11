<script lang="ts">
	// The Voice Cloud mark: a dot-matrix cloud on an indigo tile.
	let { size = 22 }: { size?: number } = $props();

	const rows = ['...XX..', '.XXXXX.', 'XXXXXXX', 'XXXXXXX'];
	const dots = rows.flatMap((row, y) =>
		[...row].flatMap((cell, x) => (cell === 'X' ? [{ x: 12.5 + x * 11, y: 29 + y * 11 }] : []))
	);
</script>

<svg width={size} height={size} viewBox="0 0 100 100" aria-hidden="true">
	<rect width="100" height="100" rx="24" fill="#3f38de" />
	{#each dots as dot}
		<rect x={dot.x} y={dot.y} width="9" height="9" rx="1.5" fill="#fff" />
	{/each}
</svg>
