<script lang="ts">
	import InfoTip from '#lib/components/InfoTip.svelte';
	import { enhance } from '$app/forms';
	import type { PageProps } from './$types';
	import { setMode, userPrefersMode } from 'mode-watcher';
	import { toast } from 'svelte-sonner';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Field from '#lib/components/ui/field/index.js';

	let { data, form }: PageProps = $props();
	let name = $state('');
	let busy = $state(false);
	$effect(() => {
		name = data.user?.name ?? '';
	});
</script>

<PageHeader
	title="Account"
	path="settings / account"
	description="Your profile and console preferences."
/>
<div class="settings-grid">
	<nav class="settings-nav" aria-label="Account sections">
		<a href="#profile">Profile</a><a href="#appearance">Appearance</a><a href="#installation"
			>Installation</a
		>
	</nav>
	<div class="flex flex-col gap-4">
		<section id="profile" class="panel">
			<div class="panel-heading">
				<h2>Profile</h2>
				<span class="eyebrow">{data.user?.role?.toUpperCase()}</span>
			</div>
			<div class="panel-body">
				<div class="flex items-center gap-3 mb-4">
					<span class="avatar size-10!"
						>{(data.user?.name ?? 'Account').slice(0, 2).toUpperCase()}</span
					>
					<div><p>{data.user?.name ?? 'Account'}</p></div>
				</div>
				<form
					method="POST"
					action="?/profile"
					use:enhance={() => {
						busy = true;
						return async ({ result, update }) => {
							await update({ reset: false });
							busy = false;
							if (result.type === 'success') toast.success('Profile saved');
						};
					}}
				>
					<Field.Group
						><Field.Field
							><Field.Label for="display-name">Display name</Field.Label><Input
								id="display-name"
								name="name"
								bind:value={name}
								required
								maxlength={80}
								pattern=".*\S.*"
							/></Field.Field
						><Field.Field
							><div class="flex items-center gap-1">
								<Field.Label for="email">Email</Field.Label><InfoTip
									text="This email identifies your account. Email changes are not supported yet."
									label="About account email"
								/>
							</div>
							<Input id="email" value={data.user?.email ?? ''} readonly /></Field.Field
						></Field.Group
					>
					<div class="flex justify-end mt-4">
						<Button
							size="sm"
							type="submit"
							disabled={busy || !name.trim() || name.trim() === data.user?.name}
							>Save changes</Button
						>
					</div>
				</form>
			</div>
		</section>
		<section id="appearance" class="panel">
			<div class="panel-heading">
				<h2>Appearance</h2>
				<InfoTip text="Theme is saved on this device." label="About preferences" />
			</div>
			<div class="panel-body">
				<Field.Set
					><Field.Legend class="sr-only">Color theme</Field.Legend>
					<div class="flex gap-4 flex-wrap">
						{#each ['light', 'dark', 'system'] as option}<label class="theme-choice"
								><input
									class="sr-only"
									type="radio"
									name="theme"
									value={option}
									bind:group={userPrefersMode.current}
									onchange={() => setMode(option as 'light' | 'dark' | 'system')}
								/>
								<div
									class="theme-preview"
									class:dark-preview={option === 'dark'}
									class:system-preview={option === 'system'}
									aria-hidden="true"
								>
									<span></span>
									<div><i></i><i></i></div>
								</div>
								<span class={userPrefersMode.current === option ? 'positive' : 'muted'}
									>{userPrefersMode.current === option ? '●' : '○'}
									{option[0].toUpperCase() + option.slice(1)}</span
								></label
							>{/each}
					</div></Field.Set
				>
			</div>
		</section>
		<section id="installation" class="panel">
			<div class="panel-heading">
				<h2>Installation</h2>
				<span class="eyebrow">SELF-HOSTED</span>
			</div>
			<div class="panel-body">
				<dl class="detail-list">
					<dt>Name</dt>
					<dd>default installation</dd>
					<dt>Access model</dt>
					<dd>Shared · {data.user?.role}</dd>
					<dt>Database</dt>
					<dd>SQLite</dd>
					<dt>Console</dt>
					<dd>v0.1</dd>
				</dl>
			</div>
		</section>
	</div>
</div>

{#if form?.error}<p class="negative mt-3" role="alert">{form.error}</p>{/if}
{#if data.user?.role === 'admin'}<section class="panel section">
		<div class="panel-heading">
			<h2>Add member</h2>
			<InfoTip
				text="Members can view shared configuration and history. Only administrators can change router configuration."
			/>
		</div>
		<form
			method="POST"
			action="?/member"
			class="panel-body"
			use:enhance={() => {
				busy = true;
				return async ({ result, update }) => {
					await update();
					busy = false;
					if (result.type === 'success') toast.success('Member created');
				};
			}}
		>
			<Field.Group
				><Field.Field
					><Field.Label for="member-name">Name</Field.Label><Input
						id="member-name"
						name="name"
						required
					/></Field.Field
				><Field.Field
					><Field.Label for="member-email">Email</Field.Label><Input
						id="member-email"
						name="email"
						type="email"
						autocomplete="off"
						required
					/></Field.Field
				><Field.Field
					><Field.Label for="member-password">Initial password</Field.Label><Input
						id="member-password"
						name="password"
						type="password"
						minlength={12}
						autocomplete="new-password"
						required
					/></Field.Field
				></Field.Group
			><Button size="sm" type="submit" class="mt-3" disabled={busy}>Create member</Button>
		</form>
	</section>{/if}
