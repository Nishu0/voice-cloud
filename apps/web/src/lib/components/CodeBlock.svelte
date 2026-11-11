<script lang="ts">
	import { Button } from '#lib/components/ui/button/index.js';
	import Copy from 'phosphor-svelte/lib/Copy';
	import { toast } from 'svelte-sonner';
	let { title, code }: { title: string; code: string } = $props();
	async function copy() {
		try {
			await navigator.clipboard.writeText(code);
			toast.success('Copied to clipboard');
		} catch {
			toast.error('Clipboard unavailable. Select and copy the code.');
		}
	}
</script>

<div class="code-block">
	<div class="panel-heading">
		<span><span class="status-dot"></span>{title}</span><Button
			size="icon-xs"
			variant="ghost"
			aria-label={`Copy ${title}`}
			onclick={copy}><Copy /></Button
		>
	</div>
	<pre><code
			>{#each code.split('\n') as line}<span
					class:code-comment={line.trim().startsWith('#') || line.trim().startsWith('//')}
					class:code-command={/^(curl|import|from|const|export|bun|docker)/.test(line)}
					>{line || ' '}</span
				>{/each}</code
		></pre>
</div>
