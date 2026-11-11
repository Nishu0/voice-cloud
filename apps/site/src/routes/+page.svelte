<script lang="ts">
	import Dither from '#lib/components/Dither.svelte';
	import Wordmark from '#lib/components/Wordmark.svelte';
	import Navbar from '#lib/components/Navbar.svelte';
	import { APP_URL, DOCS_URL, REPO_URL } from '#lib/links.js';

	let hero: HTMLElement;
	let pastHero = $state(false);

	const sections = [
		{
			id: 'router',
			title: 'Voice router',
			text: 'Connect local and cloud voice models through one router, with policies for language, latency, and cost.'
		},
		{
			id: 'self-host',
			title: 'Self-hosted',
			text: 'Run the router on your own infrastructure.'
		}
	];
	let current = $state(sections[0]);

	$effect(() => {
		const observer = new IntersectionObserver(([entry]) => (pastHero = !entry.isIntersecting));
		observer.observe(hero);
		return () => observer.disconnect();
	});

	// The navbar's second row follows whichever section has reached it.
	function track() {
		let active = sections[0];
		for (const section of sections) {
			const top = document.getElementById(section.id)?.getBoundingClientRect().top;
			if (top !== undefined && top <= 160) active = section;
		}
		current = active;
	}

	// Illustrative data for the console mock-ups below.
	const targets = [
		{ y: 44, name: 'whisper-large-v3', via: 'Local runtime', edge: 'priority 0' },
		{ y: 130, name: 'nova-3', via: 'Deepgram', edge: 'priority 1' },
		{ y: 216, name: 'universal', via: 'AssemblyAI', edge: 'priority 2' }
	];
	const connections = [
		{ name: 'Deepgram', kind: 'cloud', on: true },
		{ name: 'ElevenLabs', kind: 'cloud', on: true },
		{ name: 'Local runtime', kind: 'local · 127.0.0.1:8081', on: true },
		{ name: 'Cartesia', kind: 'cloud', on: false },
		{ name: 'OpenAI', kind: 'cloud', on: false }
	];
	const deployments = [
		{ model: 'nova-3', via: 'Deepgram', task: 'STT', langs: 'en · hi · es', health: 'healthy' },
		{ model: 'whisper-large-v3', via: 'Local runtime', task: 'STT', langs: 'en · ja', health: 'healthy' },
		{ model: 'sonic-2', via: 'Cartesia', task: 'TTS', langs: 'en', health: 'unknown' },
		{ model: 'eleven_flash_v2_5', via: 'ElevenLabs', task: 'TTS', langs: 'en · de · fr', health: 'healthy' },
		{ model: 'gpt-4o-mini', via: 'OpenAI', task: 'LLM', langs: '—', health: 'unknown' }
	];
	const activity = [
		{ time: '14:02', text: 'Connection Deepgram enabled' },
		{ time: '14:03', text: 'Deployment nova-3 created' },
		{ time: '14:03', text: 'Language hi added to nova-3' },
		{ time: '14:05', text: 'Policy support-stt created' },
		{ time: '14:05', text: 'Target whisper-large-v3 set · priority 0' },
		{ time: '14:06', text: 'Health check nova-3 · healthy' }
	];
	const providers =
		'Alibaba, Anthropic, AssemblyAI, Azure, Baseten, Cartesia, Cerebras, Deepgram, DeepInfra, ElevenLabs, Fish Audio, Gemini, Gladia, Google, Gradium, Hamsa, Hopper, Hume, Inworld, Maya Research, Meta, MiniMax, Modulate, Munsit, Nari Labs, Omi Health, OpenAI, Palabra, Paxa Labs, Rime, Smallest AI, Soniox, Speechify, Speechmatics, Together, xAI'.split(
			', '
		);
</script>

<svelte:head>
	<title>Voice Cloud — an open router for voice</title>
	<meta
		name="description"
		content="Connect local and cloud voice models through one self-hosted router, with policies for language, latency, and cost."
	/>
</svelte:head>

<div class="backdrop"><Dither /></div>
<svelte:window onscroll={track} />
<Navbar visible={pastHero} {sections} {current} />

<section class="hero" id="top" bind:this={hero}>
	<div class="rule"></div>
	<Wordmark text="Voice Cloud" />
	<div class="below-word">
		<div class="rule short"></div>
		<div class="choices">
			<a href={APP_URL}>Open app <span aria-hidden="true">↗</span></a>
			<div class="rule"></div>
			<a href={DOCS_URL}>Read the docs <span aria-hidden="true">→</span></a>
		</div>
		<div class="rule"></div>
	</div>
</section>

<main>
	<section id="router">
		<div class="intro">
			<div>
				<h2>Local and cloud voice models behind one router</h2>
				<a class="button" href={APP_URL}>Open app</a>
			</div>
			<p>
				Add your providers, describe the models they serve, and decide which one answers. Built in
				Go, for voice applications and agents.
			</p>
		</div>
		<div class="cards">
			<article>
				<div class="mock" aria-hidden="true">
					<div class="panel">
						<header><b>support-stt · Policy</b><span>STT</span></header>
						<svg class="flow" viewBox="0 0 520 290">
							<defs>
								<pattern id="dots" width="16" height="16" patternUnits="userSpaceOnUse">
									<circle cx="1" cy="1" r="1" />
								</pattern>
							</defs>
							<rect width="520" height="290" fill="url(#dots)" />
							<path class="wire" d="M96 145h40" />
							{#each targets as target}
								<path class="wire" d="M262 145C310 145 300 {target.y + 29} 348 {target.y + 29}" />
								<text class="edge" x="302" y={target.y + (target.y < 130 ? 16 : target.y > 130 ? 52 : 6)}
									>{target.edge}</text
								>
								<rect class="node" x="348" y={target.y} width="160" height="58" rx="8" />
								<text x="362" y={target.y + 25}>{target.name}</text>
								<text class="sub" x="362" y={target.y + 43}>{target.via}</text>
							{/each}
							<rect class="node" x="12" y="121" width="84" height="48" rx="8" />
							<text x="26" y="150">Request</text>
							<rect class="node active" x="136" y="121" width="126" height="48" rx="8" />
							<text x="150" y="150">support-stt</text>
						</svg>
					</div>
				</div>
				<h3>Policies</h3>
				<p>
					Rank target deployments by priority. When the first choice is unavailable, the route
					falls through to the next.
				</p>
			</article>

			<article>
				<div class="mock" aria-hidden="true">
					<div class="panel">
						<header><b>Connections</b><span>{providers.length} seeded</span></header>
						{#each connections as connection}
							<div class="row">
								<b>{connection.name}</b>
								<span>{connection.kind}</span>
								<i class="switch" class:on={connection.on}></i>
							</div>
						{/each}
					</div>
				</div>
				<h3>Connections</h3>
				<p>
					Add a provider account or point at an endpoint of your own. Cloud and local models are
					configured the same way.
				</p>
			</article>

			<article>
				<div class="mock" aria-hidden="true">
					<div class="panel">
						<header><b>Deployments</b><span>Checked by the worker</span></header>
						{#each deployments as deployment}
							<div class="row deployment">
								<i class="dot" class:ok={deployment.health === 'healthy'}></i>
								<b>{deployment.model}<small>{deployment.via}</small></b>
								<span>{deployment.task}</span>
								<span>{deployment.langs}</span>
								<span>{deployment.health}</span>
							</div>
						{/each}
					</div>
				</div>
				<h3>Deployments</h3>
				<p>
					Serve a model through a connection, tag the languages it supports, and let the worker
					keep an eye on its health.
				</p>
			</article>

			<article>
				<div class="mock" aria-hidden="true">
					<div class="panel">
						<header><b>Activity</b><span>Today</span></header>
						{#each activity as event}
							<div class="row event">
								<span>{event.time}</span>
								<b>{event.text}</b>
							</div>
						{/each}
					</div>
				</div>
				<h3>Activity</h3>
				<p>See what changed in your installation, in the order it happened.</p>
			</article>
		</div>
	</section>

	<section id="self-host">
		<div class="intro">
			<div>
				<h2>Your hardware, one Compose file</h2>
				<a class="button" href={DOCS_URL}>Read the docs</a>
			</div>
			<p>
				There is no separate database server to operate. Voice Cloud is in early development; APIs
				and setup are still taking shape.
			</p>
		</div>
		<div class="cards">
			<article>
				<div class="mock">
					<div class="panel">
						<header><b>Terminal</b><span>repository root</span></header>
						<pre><code
								>docker compose -f docker/compose.yml up --build -d

<span># console  http://localhost:3000</span>
<span># api      http://localhost:8080</span>
<span># health   http://localhost:8080/healthz</span></code
							></pre>
					</div>
				</div>
				<h3>Compose</h3>
				<p>
					One backend container runs the Go API and its worker, and the console runs beside it.
					SQLite lives on a volume.
				</p>
			</article>

			<article>
				<div class="mock">
					<div class="panel">
						<header><b>Provider catalog</b><span>{providers.length} connections</span></header>
						<ul class="chips">
							{#each providers as name}<li>{name}</li>{/each}
						</ul>
					</div>
				</div>
				<h3>Providers</h3>
				<p>
					Seeded providers stay disabled until you configure them. Add any number of custom or
					local endpoints alongside.
				</p>
			</article>
		</div>
	</section>
</main>

<section class="closing">
	<div class="rule"></div>
	<div class="choices">
		<a href={APP_URL}>Open app <span aria-hidden="true">↗</span></a>
		<div class="rule"></div>
		<a href={DOCS_URL}>Read the docs <span aria-hidden="true">→</span></a>
	</div>
	<div class="rule"></div>
</section>

<footer>
	<span>Voice Cloud</span>
	<a href={REPO_URL}>GitHub</a>
</footer>

<style>
	/* The centre line: one hairline from the top of the hero to the footer. */
	.rule {
		width: 1px;
		min-height: 0;
		justify-self: center;
		background: var(--rule, rgb(255 255 255 / 0.3));
	}

	/* One dithered sky behind the whole page; it stays put while content scrolls. */
	.backdrop {
		position: fixed;
		inset: 0;
		z-index: -1;
	}

	.hero {
		position: relative;
		display: grid;
		grid-template-rows: 1fr auto 1fr;
		justify-items: center;
		min-height: 100svh;
	}
	.hero > .rule {
		margin-bottom: 36px;
	}
	.below-word {
		display: grid;
		grid-template-rows: 36px auto 1fr;
		justify-items: center;
		width: 100%;
		padding-top: 36px;
	}

	.choices {
		position: relative;
		display: grid;
		grid-template-columns: 1fr 1px 1fr;
		width: 100%;
		font-size: 17px;
		font-weight: 600;
		color: #fff;
	}
	.choices a {
		padding: 18px 32px;
		white-space: nowrap;
	}
	.choices a:first-child {
		text-align: right;
	}
	.choices a span {
		display: inline-block;
		margin-left: 6px;
		transition: transform 0.18s;
	}
	.choices a:hover {
		text-decoration: underline;
		text-underline-offset: 6px;
	}
	.choices a:hover span {
		transform: translateX(3px);
	}

	main {
		position: relative;
		overflow-x: clip;
		padding-bottom: 120px;
	}
	main::before {
		content: '';
		position: absolute;
		inset: 0 auto 0 50%;
		width: 1px;
		background: rgb(255 255 255 / 0.3);
	}
	main > section {
		position: relative;
		max-width: var(--page);
		margin: 0 auto;
		padding: 96px var(--gutter) 0;
	}

	.intro,
	.cards {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 32px 40px;
	}
	.intro {
		padding: 24px 4px 64px;
	}
	.intro h2 {
		margin: 0;
		color: #fff;
		max-width: 16ch;
		font-size: clamp(1.75rem, 3.4vw, 2.6rem);
		font-weight: 600;
		line-height: 1.15;
		letter-spacing: -0.045em;
		text-wrap: balance;
	}
	.intro p {
		max-width: 42ch;
		padding-top: 6px;
		color: #e4e0ff;
		font-size: 17px;
	}
	.button {
		display: inline-block;
		margin-top: 28px;
		padding: 10px 18px;
		border-radius: 10px;
		background: #fff;
		color: var(--ground);
		font-size: 14px;
		font-weight: 600;
		transition: background 0.15s;
	}
	.button:hover {
		background: #e4e0ff;
	}

	article {
		display: flex;
		flex-direction: column;
		min-width: 0;
		padding: 28px;
		border: 1px solid var(--line);
		border-radius: 22px;
		background: var(--raised);
		color: var(--ink);
		/* Cards are light: remap the page tokens for everything inside. */
		--ink: var(--card-ink);
		--dim: var(--card-dim);
		--line: var(--card-line);
		--raised: var(--card);
		--ground: var(--card-panel);
		--lilac: var(--sky);
	}
	article h3 {
		margin: 28px 0 6px;
		color: var(--dim);
		font-size: 13px;
		font-weight: 400;
	}
	article > p {
		max-width: 44ch;
		font-size: 15px;
		line-height: 1.6;
	}

	/* Console mock-ups */
	.mock {
		height: 312px;
		overflow: hidden;
		font-size: 13px;
		mask-image: linear-gradient(#000 82%, transparent);
	}
	.panel {
		border: 1px solid var(--line);
		border-radius: 12px;
		background: var(--ground);
		overflow: hidden;
	}
	.panel header,
	.row {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 14px 18px;
	}
	.panel header {
		justify-content: space-between;
		border-bottom: 1px solid var(--line);
	}
	.panel b {
		font-weight: 500;
	}
	.panel span,
	.panel small {
		color: var(--dim);
		font-size: 12px;
	}
	.row + .row {
		border-top: 1px solid var(--line);
	}
	.row > span:nth-child(2) {
		flex: 1;
	}
	.switch {
		position: relative;
		flex: none;
		width: 30px;
		height: 18px;
		border-radius: 9px;
		background: var(--line);
	}
	.switch::after {
		content: '';
		position: absolute;
		top: 3px;
		left: 3px;
		width: 12px;
		height: 12px;
		border-radius: 50%;
		background: var(--dim);
	}
	.switch.on {
		background: var(--sky);
	}
	.switch.on::after {
		left: 15px;
		background: #fff;
	}
	.deployment {
		display: grid;
		grid-template-columns: 6px minmax(0, 1.6fr) 34px minmax(0, 1fr) 62px;
		padding-block: 10px;
	}
	.deployment small {
		display: block;
	}
	.deployment b,
	.deployment span {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: var(--dim);
	}
	.dot.ok {
		background: #1f9d6b;
	}
	.event span {
		flex: none !important;
		font-variant-numeric: tabular-nums;
	}

	.flow {
		display: block;
		width: 100%;
		font-size: 13px;
		fill: var(--ink);
	}
	.flow pattern circle {
		fill: var(--line);
	}
	.flow .node {
		fill: var(--raised);
		stroke: var(--line);
	}
	.flow .node.active {
		stroke: var(--lilac);
	}
	.flow .wire {
		fill: none;
		stroke: var(--lilac);
		stroke-opacity: 0.7;
	}
	.flow .sub,
	.flow .edge {
		fill: var(--dim);
		font-size: 11px;
	}
	.flow .edge {
		text-anchor: middle;
	}

	pre {
		margin: 0;
		padding: 18px;
		overflow-x: auto;
		font: inherit;
		line-height: 1.9;
	}
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin: 0;
		padding: 18px;
		list-style: none;
		font-size: 12px;
	}
	.chips li {
		padding: 2px 9px;
		border: 1px solid var(--line);
		border-radius: 999px;
		color: var(--dim);
	}

	.closing {
		position: relative;
		display: grid;
		grid-template-rows: 1fr auto 1fr;
		height: 420px;
	}

	footer {
		display: flex;
		justify-content: space-between;
		padding: 22px max(var(--gutter), calc((100% - var(--page)) / 2 + var(--gutter)));
		border-top: 1px solid rgb(255 255 255 / 0.3);
		color: #fff;
		font-size: 12px;
	}
	/* Light chips keep the text readable when a cloud drifts behind it. */
	footer > * {
		padding: 4px 12px;
		border-radius: 999px;
		background: var(--card);
		color: var(--card-ink);
	}
	footer a:hover {
		background: #fff;
	}

	@media (max-width: 800px) {
		main::before {
			display: none;
		}
		main > section {
			padding-top: 64px;
		}
		.intro,
		.cards {
			grid-template-columns: 1fr;
		}
		.intro {
			padding: 48px 4px 40px;
		}
		article {
			padding: 20px;
		}
		.choices a {
			padding: 16px 18px;
			font-size: 14px;
		}
	}
</style>
