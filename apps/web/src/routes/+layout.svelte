<script lang="ts">
	import './layout.css';
	import favicon from '#lib/assets/favicon.svg';
	import type { LayoutProps } from './$types';
	import { page } from '$app/state';
	import { invalidateAll } from '$app/navigation';
	import { ModeWatcher } from 'mode-watcher';
	import * as Sidebar from '#lib/components/ui/sidebar/index.js';
	import * as Dialog from '#lib/components/ui/dialog/index.js';
	import * as Tooltip from '#lib/components/ui/tooltip/index.js';
	import InfoTip from '#lib/components/InfoTip.svelte';
	import { Toaster } from '#lib/components/ui/sonner/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { Input } from '#lib/components/ui/input/index.js';
	import AppSidebar from '#lib/components/AppSidebar.svelte';
	import { navigation } from '#lib/fixtures.js';

	let { children }: LayoutProps = $props();
	let searchOpen = $state(false);
	let search = $state('');
	const links = navigation.flatMap((group) => group.items);
	const filtered = $derived(
		links.filter((link) => link.title.toLowerCase().includes(search.toLowerCase())),
	);
	const isLogin = $derived(['/login', '/signup'].includes(page.url.pathname));
	function shortcut(event: KeyboardEvent) {
		if (!isLogin && (event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
			event.preventDefault();
			searchOpen = !searchOpen;
		}
	}
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<ModeWatcher defaultMode="dark" />
<svelte:window onkeydown={shortcut} />
<a href="#main" class="skip-link">Skip to content</a>
<Tooltip.Provider delayDuration={150}>
	{#if isLogin}
		<main id="main" tabindex="-1">{@render children()}</main>
	{:else}
		<Sidebar.Provider style="--sidebar-width: 13.25rem; --sidebar-width-icon: 3rem;">
			<AppSidebar
				onsearch={() => {
					search = '';
					searchOpen = true;
				}}
			/>
			<Sidebar.Inset>
				<div class="topbar">
					<div class="flex items-center gap-2">
						<Sidebar.Trigger /><span class="muted">console</span><span class="muted">/</span><span
							>{links.find((link) => link.href === page.url.pathname)?.title.toLowerCase() ??
								'account'}</span
						>
					</div>
					<div class="flex items-center gap-2">
						<span class="eyebrow">CONNECTED</span><InfoTip
							text="Data comes from your installation. Refresh to check for changes."
						/><Button size="xs" variant="outline" onclick={() => invalidateAll()}>Refresh ↻</Button>
					</div>
				</div>
				<main id="main" class="workspace" tabindex="-1">{@render children()}</main>
			</Sidebar.Inset>
		</Sidebar.Provider>
		<Dialog.Root bind:open={searchOpen}
			><Dialog.Content class="gap-3"
				><Dialog.Header
					><Dialog.Title>Go to page</Dialog.Title><Dialog.Description class="sr-only"
						>Search pages. Escape to close.</Dialog.Description
					></Dialog.Header
				><Input aria-label="Search pages" placeholder="Search..." bind:value={search} />
				<nav class="flex flex-col gap-1" aria-label="Search results">
					{#each filtered as link}<a
							class="search-result"
							href={link.href}
							onclick={() => (searchOpen = false)}
							><span>{link.title}</span><span class="muted">↵</span></a
						>{:else}<p class="muted py-3">No pages found.</p>{/each}
				</nav></Dialog.Content
			></Dialog.Root
		>
	{/if}
</Tooltip.Provider>
<Toaster />
