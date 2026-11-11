<script lang="ts">
	import InfoTip from '#lib/components/InfoTip.svelte';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import { Badge } from '#lib/components/ui/badge/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import * as Field from '#lib/components/ui/field/index.js';
	import { usePreview, type PreviewKey } from '#lib/preview.svelte.js';
	import { toast } from 'svelte-sonner';
	const preview = usePreview();
	let open = $state(false),
		name = $state(''),
		expiry = $state('30 days'),
		token = $state(''),
		error = $state('');
	let revoke = $state<PreviewKey | null>(null),
		revokeOpen = $state(false);
	function create(event: SubmitEvent) {
		event.preventDefault();
		if (!name.trim()) {
			error = 'Give this key a name.';
			return;
		}
		const id = crypto.randomUUID().replaceAll('-', '');
		token = `vc_demo_not_valid_${id}`;
		preview.keys.push({
			id,
			name: name.trim(),
			prefix: token.slice(0, 24),
			expiry,
			revoked: false,
		});
		error = '';
		toast.success('Demo key created locally');
	}
	function start() {
		name = '';
		token = '';
		error = '';
		open = true;
	}
	async function copy() {
		try {
			await navigator.clipboard.writeText(token);
			toast.success('Demo key copied');
		} catch {
			toast.error('Clipboard unavailable. Select and copy the key.');
		}
	}
</script>

<PageHeader title="API keys" path="connect / api keys" description="Manage access to your router."
	>{#snippet actions()}<Button size="sm" onclick={start}>Create API key +</Button
		>{/snippet}</PageHeader
>

<section class="section">
	<div class="section-heading">
		<h2>Application keys <span class="muted">/ {preview.keys.length}</span></h2>
		<div class="flex items-center gap-1">
			<Badge variant="outline">Demo</Badge><InfoTip
				label="About API keys"
				text="Installation-wide keys. Demo changes last until reload. Keep real keys on your server; never expose them in frontend code."
			/>
		</div>
	</div>
	{#if preview.keys.length}<div class="panel">
			<div class="table-wrap">
				<table class="data-table">
					<thead
						><tr><th>Name</th><th>Key prefix</th><th>Expires in</th><th>Status</th><th></th></tr
						></thead
					><tbody
						>{#each preview.keys as key}<tr
								><td>{key.name}</td><td class="muted">{key.prefix}...</td><td>{key.expiry}</td><td
									class={key.revoked ? 'muted' : 'positive'}
									>{key.revoked ? 'revoked' : 'active · demo'}</td
								><td
									><Button
										size="xs"
										variant="destructive"
										disabled={key.revoked}
										onclick={() => {
											revoke = key;
											revokeOpen = true;
										}}>Revoke</Button
									></td
								></tr
							>{/each}</tbody
					>
				</table>
			</div>
			<div class="table-footer">
				<span>{preview.keys.filter((k) => !k.revoked).length} active demo keys</span><span
					>Cleared on reload</span
				>
			</div>
		</div>{:else}<EmptyState
			title="No API keys yet"
			description="Demo keys cannot authenticate requests."
			><Button size="sm" onclick={start}>Create API key +</Button></EmptyState
		>{/if}
</section>

<Dialog.Root bind:open
	><Dialog.Content
		><Dialog.Header
			><Dialog.Title>{token ? 'Save your demo key' : 'Create API key'}</Dialog.Title
			><Dialog.Description
				>{token ? 'Shown once. Copy it now.' : 'Demo key · no API access.'}</Dialog.Description
			></Dialog.Header
		>{#if token}<div class="flex flex-col gap-3">
				<label for="new-token" class="muted text-[10px]">DEMO KEY</label><Input
					id="new-token"
					readonly
					value={token}
				/><Button variant="outline" onclick={copy}>Copy demo key</Button><Button
					onclick={() => {
						open = false;
						token = '';
					}}>I have saved it</Button
				>
			</div>{:else}<form onsubmit={create}>
				<Field.Group
					><Field.Field data-invalid={!!error}
						><Field.Label for="key-name">Name</Field.Label><Input
							id="key-name"
							placeholder="e.g. development-agent"
							maxlength={80}
							bind:value={name}
							aria-invalid={!!error}
							required
						/>{#if error}<Field.Error>{error}</Field.Error>{/if}</Field.Field
					><Field.Field
						><Field.Label for="expiry">Expires in</Field.Label><select
							id="expiry"
							class="native-select"
							bind:value={expiry}
							><option>7 days</option><option>30 days</option><option>90 days</option></select
						></Field.Field
					></Field.Group
				><Dialog.Footer class="mt-5"
					><Button variant="outline" onclick={() => (open = false)}>Cancel</Button><Button
						type="submit">Create demo key</Button
					></Dialog.Footer
				>
			</form>{/if}</Dialog.Content
	></Dialog.Root
>
<Dialog.Root bind:open={revokeOpen}
	><Dialog.Content
		><Dialog.Header
			><Dialog.Title>Revoke {revoke?.name}?</Dialog.Title><Dialog.Description
				>Revoke this demo key for this session.</Dialog.Description
			></Dialog.Header
		><Dialog.Footer
			><Button variant="outline" onclick={() => (revokeOpen = false)}>Cancel</Button><Button
				variant="destructive"
				onclick={() => {
					if (revoke) revoke.revoked = true;
					revokeOpen = false;
					toast.success('Demo key revoked');
				}}>Revoke key</Button
			></Dialog.Footer
		></Dialog.Content
	></Dialog.Root
>
