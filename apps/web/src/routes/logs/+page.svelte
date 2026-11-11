<script lang="ts">
	import PageHeader from '#lib/components/PageHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
</script>

<PageHeader
	title="Logs"
	path="observe / logs"
	description="A live log API is not available yet. Backend and worker logs are available in the process output."
/>
<EmptyState
	title="Log stream unavailable"
	description="The backend does not expose a log stream. View worker probe results under Deployments."
	href="/manage/deployments"
	action="Deployment health"
/>
