<script lang="ts">
	import InfoTip from '#lib/components/InfoTip.svelte';
	import { onMount } from 'svelte';
	import { setMode } from 'mode-watcher';
	import { toast } from 'svelte-sonner';
	import PageHeader from '#lib/components/PageHeader.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import * as Field from '#lib/components/ui/field/index.js';
	import { usePreview } from '#lib/preview.svelte.js';
	const preview = usePreview();
	let name = $state(preview.name),
		theme = $state('dark');
	onMount(() => {
		try {
			const saved = localStorage.getItem('mode-watcher-mode');
			if (saved && ['light', 'dark', 'system'].includes(saved)) theme = saved;
		} catch {
			/* Theme still works when browser storage is unavailable. */
		}
	});
	function save(event: SubmitEvent) {
		event.preventDefault();
		if (name.trim()) {
			preview.name = name.trim();
			toast.success('Profile updated for this preview session');
		}
	}
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
				<span class="eyebrow">ADMINISTRATOR</span>
			</div>
			<div class="panel-body">
				<div class="flex items-center gap-3 mb-4">
					<span class="avatar size-10!">{preview.name.slice(0, 2).toUpperCase()}</span>
					<div><p>{preview.name}</p></div>
				</div>
				<form onsubmit={save}>
					<Field.Group
						><Field.Field
							><Field.Label for="display-name">Display name</Field.Label><Input
								id="display-name"
								bind:value={name}
								required
								maxlength={80}
								pattern=".*\S.*"
							/></Field.Field
						><Field.Field
							><div class="flex items-center gap-1">
								<Field.Label for="email">Email</Field.Label><InfoTip
									text="Example account. Authentication is not connected."
									label="About account email"
								/>
							</div>
							<Input id="email" value="admin@example.test" readonly /></Field.Field
						></Field.Group
					>
					<div class="flex justify-end mt-4">
						<Button size="sm" type="submit" disabled={!name.trim() || name.trim() === preview.name}
							>Save changes</Button
						>
					</div>
				</form>
			</div>
		</section>
		<section id="appearance" class="panel">
			<div class="panel-heading">
				<h2>Appearance</h2>
				<InfoTip
					text="Theme is saved on this device. Profile changes last until reload."
					label="About preferences"
				/>
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
									bind:group={theme}
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
								<span class={theme === option ? 'positive' : 'muted'}
									>{theme === option ? '●' : '○'} {option[0].toUpperCase() + option.slice(1)}</span
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
					<dd>Shared · admin / member</dd>
					<dt>Database</dt>
					<dd>SQLite</dd>
					<dt>Console</dt>
					<dd>v0.1 · UI preview</dd>
				</dl>
			</div>
		</section>
	</div>
</div>
