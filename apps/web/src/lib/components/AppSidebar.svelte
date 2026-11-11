<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import * as Sidebar from '#lib/components/ui/sidebar/index.js';
	import * as Dropdown from '#lib/components/ui/dropdown-menu/index.js';
	import { navigation } from '#lib/fixtures.js';
	import { usePreview } from '#lib/preview.svelte.js';
	import Terminal from 'phosphor-svelte/lib/TerminalWindow';
	import Squares from 'phosphor-svelte/lib/SquaresFour';
	import Code from 'phosphor-svelte/lib/Code';
	import Stack from 'phosphor-svelte/lib/Stack';
	import Key from 'phosphor-svelte/lib/Key';
	import List from 'phosphor-svelte/lib/ListDashes';
	import Pulse from 'phosphor-svelte/lib/Pulse';
	import Gear from 'phosphor-svelte/lib/GearSix';
	import Book from 'phosphor-svelte/lib/BookOpen';
	import Caret from 'phosphor-svelte/lib/CaretUpDown';
	import Search from 'phosphor-svelte/lib/MagnifyingGlass';
	let { onsearch }: { onsearch: () => void } = $props();
	const preview = usePreview();
	const sidebar = Sidebar.useSidebar();
	const icons = {
		overview: Squares,
		code: Code,
		models: Stack,
		key: Key,
		logs: List,
		activity: Pulse,
		settings: Gear,
	};
	const settings = $derived(page.url.pathname.startsWith('/settings'));
</script>

<Sidebar.Root collapsible="icon">
	<Sidebar.Header class="gap-0 p-2">
		<a href="/" class="brand" aria-label="Voice Cloud home"
			><span class="brand-mark"><Terminal size={19} /></span><span class="brand-label"
				>voice<span class="muted">cloud</span><small>GATEWAY / ROUTER</small></span
			><span class="brand-version">v0.1</span></a
		>
		<Dropdown.Root
			><Dropdown.Trigger class="installation"
				><span class="status-dot"></span><span class="truncate">default installation</span><Caret
					size={13}
				/></Dropdown.Trigger
			><Dropdown.Content align="start"
				><Dropdown.Group
					><Dropdown.Label>Shared installation</Dropdown.Label><Dropdown.Item
						onSelect={() => goto('/settings/account')}>default · administrator</Dropdown.Item
					></Dropdown.Group
				></Dropdown.Content
			></Dropdown.Root
		>
		<button class="search-trigger" onclick={onsearch}
			><Search size={14} /><span>Search pages...</span><kbd>⌘ K</kbd></button
		>
	</Sidebar.Header>
	<Sidebar.Content class="gap-1">
		{#if settings}
			<Sidebar.Group class="p-2"
				><Sidebar.GroupLabel>Settings</Sidebar.GroupLabel><Sidebar.Menu
					><Sidebar.MenuItem
						><Sidebar.MenuButton onclick={() => sidebar.setOpenMobile(false)}
							>{#snippet child({ props })}<a href="/" {...props}>← <span>Back to overview</span></a
								>{/snippet}</Sidebar.MenuButton
						></Sidebar.MenuItem
					><Sidebar.MenuItem
						><Sidebar.MenuButton isActive tooltipContent="Account"
							>{#snippet child({ props })}<a href="/settings/account" {...props}
									><Gear /><span>Account</span></a
								>{/snippet}</Sidebar.MenuButton
						></Sidebar.MenuItem
					></Sidebar.Menu
				></Sidebar.Group
			>
		{:else}
			{#each navigation as group}<Sidebar.Group class="p-2"
					><Sidebar.GroupLabel>{group.label}</Sidebar.GroupLabel><Sidebar.GroupContent
						><Sidebar.Menu
							>{#each group.items as item}{@const Icon =
									icons[item.icon as keyof typeof icons]}<Sidebar.MenuItem
									><Sidebar.MenuButton
										size="sm"
										isActive={page.url.pathname === item.href}
										tooltipContent={item.title}
										>{#snippet child({ props })}<a
												href={item.href}
												{...props}
												aria-current={page.url.pathname === item.href ? 'page' : undefined}
												onclick={() => sidebar.setOpenMobile(false)}
												><Icon /><span>{item.title}</span><span class="nav-number"
													>{item.shortcut}</span
												></a
											>{/snippet}</Sidebar.MenuButton
									></Sidebar.MenuItem
								>{/each}</Sidebar.Menu
						></Sidebar.GroupContent
					></Sidebar.Group
				>{/each}
		{/if}
	</Sidebar.Content>
	<Sidebar.Footer class="gap-1 p-2">
		<div class="sidebar-runtime">
			<span class="status-dot"></span><span>self-hosted</span><span class="muted">sqlite</span>
		</div>
		<Sidebar.Menu
			><Sidebar.MenuItem
				><Sidebar.MenuButton tooltipContent="Documentation"
					>{#snippet child({ props })}<a href="/quickstart" {...props}
							><Book /><span>Documentation</span><span class="nav-number">↗</span></a
						>{/snippet}</Sidebar.MenuButton
				></Sidebar.MenuItem
			><Sidebar.MenuItem
				><Dropdown.Root
					><Dropdown.Trigger class="account-trigger"
						><span class="avatar">{preview.name.slice(0, 2).toUpperCase()}</span><span
							class="account-label">{preview.name}<small>Administrator</small></span
						><Caret size={14} /></Dropdown.Trigger
					><Dropdown.Content side="top" align="start"
						><Dropdown.Group
							><Dropdown.Label>Local preview</Dropdown.Label><Dropdown.Item
								onSelect={() => goto('/settings/account')}>Account & appearance</Dropdown.Item
							><Dropdown.Item onSelect={() => goto('/api-keys')}>API keys</Dropdown.Item
							><Dropdown.Item onSelect={() => goto('/login')}>Sign in</Dropdown.Item
							></Dropdown.Group
						></Dropdown.Content
					></Dropdown.Root
				></Sidebar.MenuItem
			></Sidebar.Menu
		>
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
