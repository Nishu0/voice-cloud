<script lang="ts">
	import Mark from './Mark.svelte';
	import { APP_URL, DOCS_URL } from '#lib/links.js';

	type Section = { id: string; title: string; text: string };

	// Hidden while the hero is on screen; floats in once it has scrolled away.
	// The second row names the section currently being read.
	let {
		visible,
		sections,
		current
	}: { visible: boolean; sections: Section[]; current: Section } = $props();
</script>

<header class:visible inert={!visible}>
	<nav aria-label="Main">
		<a class="brand" href="#top"><Mark size={24} /> Voice Cloud</a>
		<div class="sections">
			{#each sections as section}
				<a href="#{section.id}" aria-current={section.id === current.id ? 'true' : undefined}
					>{section.title}</a
				>
			{/each}
		</div>
		<div class="actions">
			<a href={DOCS_URL}>Docs</a>
			<a class="open" href={APP_URL}>Open app</a>
		</div>
	</nav>
	<div class="current">
		<b>{current.title}</b>
		<span>{current.text}</span>
	</div>
</header>

<style>
	header {
		position: fixed;
		top: 12px;
		left: 50%;
		z-index: 20;
		width: min(var(--page) - 2 * var(--gutter), 100% - 24px);
		border: 1px solid var(--card-line);
		border-radius: 20px;
		background: color-mix(in srgb, var(--card) 90%, transparent);
		backdrop-filter: blur(14px);
		box-shadow: 0 10px 30px rgb(11 11 40 / 0.22);
		color: var(--card-ink);
		opacity: 0;
		transform: translate(-50%, -140%);
		transition:
			transform 0.3s cubic-bezier(0.2, 0.7, 0.2, 1),
			opacity 0.2s;
	}
	header.visible {
		opacity: 1;
		transform: translate(-50%, 0);
	}
	nav {
		display: grid;
		grid-template-columns: 1fr auto 1fr;
		align-items: center;
		height: 52px;
		padding: 0 10px 0 18px;
		font-size: 13px;
	}
	a {
		color: var(--card-dim);
		transition: color 0.15s;
	}
	a:hover,
	a[aria-current] {
		color: var(--card-ink);
	}
	.brand {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		justify-self: start;
		color: var(--card-ink);
		font-size: 14px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.sections,
	.actions {
		display: flex;
		align-items: center;
		gap: 28px;
	}
	.actions {
		justify-self: end;
		gap: 18px;
	}
	.open {
		padding: 7px 14px;
		border-radius: 10px;
		background: var(--sky);
		color: #fff;
		font-weight: 600;
	}
	.open:hover {
		background: var(--sky-deep);
		color: #fff;
	}
	.current {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 24px;
		padding: 12px 18px 13px;
		border-top: 1px solid var(--card-line);
		font-size: 13px;
	}
	.current b {
		font-weight: 600;
		white-space: nowrap;
	}
	.current span {
		overflow: hidden;
		color: var(--card-dim);
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	@media (max-width: 720px) {
		nav {
			grid-template-columns: 1fr auto;
		}
		.sections,
		.current span {
			display: none;
		}
	}
</style>
