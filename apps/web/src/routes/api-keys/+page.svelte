<script lang="ts">
	import type { PageProps } from './$types';
	import { enhance, applyAction } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import EmptyState from '#lib/components/EmptyState.svelte';
	import InfoTip from '#lib/components/InfoTip.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import * as Field from '#lib/components/ui/field/index.js';
	import { toast } from 'svelte-sonner';
	let { data, form }: PageProps = $props();
	let open = $state(false),
		revoke = $state<number | null>(null),
		busy = $state(false),
		token = $state(''),
		issue = $state('');
	function start() {
		token = '';
		issue = '';
		open = true;
	}
	async function copy() {
		try {
			await navigator.clipboard.writeText(token);
			toast.success('Key copied');
		} catch {
			toast.error('Clipboard unavailable');
		}
	}
</script>

<PageHeader
	title="API keys"
	path="connect / api keys"
	description="Installation-wide router keys. Store secrets on your server."
	>{#snippet actions()}{#if data.user?.role === 'admin'}<Button size="sm" onclick={start}
				>Create API key +</Button
			>{/if}{/snippet}</PageHeader
>
{#if form?.error}<p class="negative mb-3" role="alert">{form.error}</p>{/if}
{#if data.keys.length}<div class="panel">
		<div class="table-wrap">
			<table class="data-table">
				<thead
					><tr
						><th>Name</th><th>Prefix</th><th>Policy</th><th>Expires</th><th>Status</th><th></th></tr
					></thead
				><tbody
					>{#each data.keys as key}<tr
							><td>{key.name}</td><td class="muted">{key.key_prefix}...</td><td
								>{data.policies.find((p) => p.id === key.default_policy_id)?.name ?? '—'}</td
							><td
								>{key.expires_at
									? new Date(key.expires_at).toISOString().slice(0, 10)
									: 'Never'}</td
							><td class={key.revoked_at ? 'muted' : 'positive'}
								>{key.revoked_at
									? 'revoked'
									: key.expires_at && key.expires_at <= Date.now()
										? 'expired'
										: 'active'}</td
							><td
								>{#if data.user?.role === 'admin'}<Button
										variant="destructive"
										size="xs"
										disabled={!!key.revoked_at}
										onclick={() => (revoke = key.id)}>Revoke</Button
									>{/if}</td
							></tr
						>{/each}</tbody
				>
			</table>
		</div>
		<div class="table-footer">
			<span>{data.keys.length} keys</span><InfoTip
				text="Secrets are shown only when created. Revoking a key preserves its history."
			/>
		</div>
	</div>{:else}<EmptyState
		title="No API keys"
		description="Create a key for your application."
	/>{/if}
<Dialog.Root
	bind:open
	onOpenChange={() => {
		token = '';
		issue = '';
	}}
	><Dialog.Content
		><Dialog.Header
			><Dialog.Title>{token ? 'Save your key' : 'Create API key'}</Dialog.Title><Dialog.Description
				>{token ? 'Shown once. Copy it now.' : 'Router access only.'}</Dialog.Description
			></Dialog.Header
		>{#if token}<Input aria-label="New API key" readonly value={token} /><Button
				variant="outline"
				onclick={copy}>Copy key</Button
			><Button
				onclick={() => {
					token = '';
					open = false;
				}}>Done</Button
			>{:else}<form
				method="POST"
				action="?/create"
				use:enhance={() => {
					busy = true;
					issue = '';
					return async ({ result }) => {
						busy = false;
						if (result.type === 'success') {
							token = String(result.data?.key ?? '');
							await invalidateAll();
						} else if (result.type === 'failure') {
							issue = String(result.data?.error ?? 'Could not create key');
						} else await applyAction(result);
					};
				}}
			>
				<Field.Group
					><Field.Field
						><Field.Label for="key-name">Name</Field.Label><Input
							id="key-name"
							name="name"
							required
							maxlength={100}
						/></Field.Field
					><Field.Field
						><Field.Label for="expiry">Expires in</Field.Label><select
							id="expiry"
							name="days"
							class="native-select"
							><option value="7">7 days</option><option value="30" selected>30 days</option><option
								value="90">90 days</option
							></select
						></Field.Field
					><Field.Field
						><Field.Label for="policy">Default policy</Field.Label><select
							id="policy"
							name="policy"
							class="native-select"
							><option value="">None</option>{#each data.policies as policy}<option
									value={policy.id}>{policy.name}</option
								>{/each}</select
						></Field.Field
					></Field.Group
				>{#if issue}<p class="negative mt-3" role="alert">{issue}</p>{/if}<Dialog.Footer
					class="mt-4"
					><Button variant="outline" onclick={() => (open = false)}>Cancel</Button><Button
						type="submit"
						disabled={busy}>{busy ? 'Creating...' : 'Create key'}</Button
					></Dialog.Footer
				>
			</form>{/if}</Dialog.Content
	></Dialog.Root
>
<Dialog.Root
	open={revoke !== null}
	onOpenChange={(value) => {
		if (!value) revoke = null;
	}}
	><Dialog.Content
		><Dialog.Header
			><Dialog.Title>Revoke key?</Dialog.Title><Dialog.Description
				>Applications using this key will lose access.</Dialog.Description
			></Dialog.Header
		>
		<form
			method="POST"
			action="?/revoke"
			use:enhance={() => {
				busy = true;
				return async ({ result, update }) => {
					await update();
					busy = false;
					if (result.type === 'success') revoke = null;
				};
			}}
		>
			<input type="hidden" name="id" value={revoke ?? ''} />{#if form?.error}<p
					class="negative mb-3"
					role="alert"
				>
					{form.error}
				</p>{/if}<Dialog.Footer
				><Button variant="outline" onclick={() => (revoke = null)}>Cancel</Button><Button
					variant="destructive"
					type="submit"
					disabled={busy}>Revoke key</Button
				></Dialog.Footer
			>
		</form></Dialog.Content
	></Dialog.Root
>
