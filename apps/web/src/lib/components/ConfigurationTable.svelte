<script lang="ts">
	import { page } from '$app/state';
	import { enhance } from '$app/forms';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import * as Field from '#lib/components/ui/field/index.js';
	import InfoTip from './InfoTip.svelte';
	import EmptyState from './EmptyState.svelte';
	import { configFields } from '#lib/config-fields.js';
	import type { Configuration } from '#lib/api-types.js';
	import { toast } from 'svelte-sonner';
	let {
		resource,
		configuration,
		records,
		onview,
	}: {
		resource: string;
		configuration: Configuration;
		records: object[];
		onview?: (id: number) => void;
	} = $props();
	let open = $state(false),
		removeOpen = $state(false),
		busy = $state(false),
		query = $state(''),
		editing = $state<number | null>(null),
		draft = $state<Record<string, string>>({});
	const fields = $derived(configFields[resource]);
	const admin = $derived(page.data.user?.role === 'admin');
	const rows = $derived(
		(records as Record<string, unknown>[]).filter((row) =>
			JSON.stringify(row).toLowerCase().includes(query.toLowerCase()),
		),
	);
	const columns = $derived(
		fields.filter((field) => field.kind !== 'json' && field.key !== 'credential_ref').slice(0, 6),
	);
	function edit(row?: Record<string, unknown>) {
		editing = row ? Number(row.id) : null;
		draft = Object.fromEntries(
			fields.map((field) => [
				field.key,
				row?.[field.key] !== undefined
					? typeof row[field.key] === 'object'
						? JSON.stringify(row[field.key], null, 2)
						: String(row[field.key])
					: (field.initial ?? field.options?.[0] ?? (field.kind === 'json' ? '{}' : '')),
			]),
		);
		open = true;
	}
	function label(key: string, value: unknown) {
		if (key === 'model_id')
			return configuration.models.find((model) => model.id === value)?.name ?? value;
		if (key === 'connection_id')
			return configuration.connections.find((connection) => connection.id === value)?.name ?? value;
		return typeof value === 'boolean' ? (value ? 'enabled' : 'disabled') : String(value ?? '—');
	}
</script>

<div class="flex items-center justify-between gap-3 mb-3">
	<Input
		aria-label="Search configuration"
		placeholder="Search..."
		bind:value={query}
		class="max-w-sm"
	/>{#if admin}<Button size="sm" onclick={() => edit()}>Create +</Button>{/if}
</div>
{#if page.form?.error && !open && !removeOpen}<p role="alert" class="negative mb-3">
		{page.form.error}
	</p>{/if}
{#if rows.length}<div class="panel">
		<div class="table-wrap">
			<table class="data-table">
				<thead
					><tr
						><th>ID</th>{#each columns as column}<th>{column.label}</th>{/each}<th></th></tr
					></thead
				><tbody
					>{#each rows as row}<tr
							><td class="muted">{String(row.id)}</td>{#each columns as column}<td
									>{#if column.key === 'name' && onview}<button
											class="table-link"
											onclick={() => onview?.(Number(row.id))}>{String(row.name)}</button
										>{:else}{label(column.key, row[column.key])}{/if}</td
								>{/each}<td
								><div class="flex justify-end gap-1">
									{#if onview}<Button
											size="xs"
											variant="ghost"
											onclick={() => onview?.(Number(row.id))}>View ↗</Button
										>{/if}{#if admin}<Button size="xs" variant="outline" onclick={() => edit(row)}
											>Edit</Button
										><Button
											size="xs"
											variant="ghost"
											onclick={() => {
												editing = Number(row.id);
												removeOpen = true;
											}}>Delete</Button
										>{/if}
								</div></td
							></tr
						>{/each}</tbody
				>
			</table>
		</div>
		<div class="table-footer"><span>{rows.length} records</span></div>
	</div>{:else}<EmptyState
		title="No records"
		description="Create a record or adjust your search."
	/>{/if}
<Dialog.Root bind:open
	><Dialog.Content class="max-h-[90svh] overflow-y-auto"
		><Dialog.Header
			><Dialog.Title
				>{editing ? 'Edit' : 'Create'}
				{resource === 'connections' ? 'provider' : resource.replace(/s$/, '')}</Dialog.Title
			><Dialog.Description class="sr-only">Configure this resource.</Dialog.Description
			></Dialog.Header
		>
		<form
			method="POST"
			action="?/save"
			use:enhance={() => {
				busy = true;
				return async ({ result, update }) => {
					await update({ reset: false });
					busy = false;
					if (result.type === 'success') {
						open = false;
						toast.success('Saved');
					}
				};
			}}
		>
			<input type="hidden" name="id" value={editing ?? ''} /><Field.Group
				>{#each fields as field}<Field.Field
						><div class="flex items-center gap-1">
							<Field.Label for={`config-${field.key}`}>{field.label}</Field.Label
							>{#if field.help}<InfoTip
									text={field.help}
								/>{/if}{#if editing && field.immutable}<InfoTip
									text="This field cannot be changed after creation."
								/>{/if}
						</div>
						{#if field.kind === 'boolean'}<input
								type="checkbox"
								id={`config-${field.key}`}
								name={field.key}
								checked={draft[field.key] === 'true'}
								onchange={(event) => (draft[field.key] = String(event.currentTarget.checked))}
							/>{:else if field.kind === 'json'}<textarea
								id={`config-${field.key}`}
								name={field.key}
								bind:value={draft[field.key]}
								rows="3"
								class="native-textarea"
							></textarea>{:else if field.kind === 'select'}{@const choices =
								field.key === 'model_id'
									? configuration.models.map((model) => ({
											value: String(model.id),
											label: model.name,
										}))
									: field.key === 'connection_id'
										? configuration.connections.map((connection) => ({
												value: String(connection.id),
												label: connection.name,
											}))
										: (field.options ?? []).map((value) => ({
												value,
												label: value.toUpperCase(),
											}))}<select
								id={`config-${field.key}`}
								name={field.key}
								class="native-select"
								bind:value={draft[field.key]}
								disabled={!!editing && field.immutable}
								required
								><option value="" disabled>Select...</option>{#each choices as choice}<option
										value={choice.value}>{choice.label}</option
									>{/each}</select
							>{#if editing && field.immutable}<input
									type="hidden"
									name={field.key}
									value={draft[field.key]}
								/>{/if}{:else}<Input
								id={`config-${field.key}`}
								name={field.key}
								type={field.kind === 'number' ? 'number' : 'text'}
								min={field.kind === 'number' ? 1 : undefined}
								bind:value={draft[field.key]}
								readonly={!!editing && field.immutable}
								required={!field.optional}
							/>{/if}</Field.Field
					>{/each}</Field.Group
			>{#if page.form?.error}<p class="negative mt-3" role="alert">
					{page.form.error}
				</p>{/if}<Dialog.Footer class="mt-4"
				><Button variant="outline" onclick={() => (open = false)}>Cancel</Button><Button
					type="submit"
					disabled={busy}>{busy ? 'Saving...' : 'Save'}</Button
				></Dialog.Footer
			>
		</form></Dialog.Content
	></Dialog.Root
>
<Dialog.Root bind:open={removeOpen}
	><Dialog.Content
		><Dialog.Header
			><Dialog.Title>Delete record #{editing}?</Dialog.Title><Dialog.Description
				>Referenced records cannot be deleted.</Dialog.Description
			></Dialog.Header
		>
		<form
			method="POST"
			action="?/remove"
			use:enhance={() => {
				busy = true;
				return async ({ result, update }) => {
					await update();
					busy = false;
					if (result.type === 'success') removeOpen = false;
				};
			}}
		>
			<input type="hidden" name="id" value={editing ?? ''} />{#if page.form?.error}<p
					class="negative mb-3"
					role="alert"
				>
					{page.form.error}
				</p>{/if}<Dialog.Footer
				><Button variant="outline" onclick={() => (removeOpen = false)}>Cancel</Button><Button
					variant="destructive"
					type="submit"
					disabled={busy}>Delete</Button
				></Dialog.Footer
			>
		</form></Dialog.Content
	></Dialog.Root
>
