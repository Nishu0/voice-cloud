<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '#lib/components/ui/button/index.js';
</script>

<div class="panel panel-body">
	<div class="eyebrow">ERROR / {page.status}</div>
	<h1>{page.error?.message ?? 'Request failed'}</h1>
	<div class="actions mt-4">
		<Button size="sm" onclick={() => location.reload()}>Retry ↻</Button><Button
			size="sm"
			variant="outline"
			href="/login">Sign in</Button
		>
	</div>
</div>
