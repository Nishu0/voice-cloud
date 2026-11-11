<script lang="ts">
	let { size = 20 }: { size?: number } = $props();
</script>

<svg width={size} height={size} viewBox="0 0 100 100" aria-hidden="true">
	<path d="M50 4C45 29 29 45 4 50c25 0 46-21 46-46Z" fill="#c9adf9" />
	<path d="M50 4c5 25 21 41 46 46-25 0-46-21-46-46Z" fill="#5a4fd8" />
	<path d="M4 50c25 5 41 21 46 46 0-25-21-46-46-46Z" fill="#5b54ee" />
	<path d="M96 50c-25 5-41 21-46 46 0-25 21-46 46-46Z" fill="#bb8ff3" />
</svg>
